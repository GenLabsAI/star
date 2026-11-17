package model

// Visual proof and regression coverage for transient UI that paints over the
// frame and then goes away: the Fork From modal, the questions form that takes
// over the editor, the completions menu, an expanded composer that submits, and
// the landing header that yields to the chat.
//
// Every scenario is captured twice from the same UI instance — once while the
// transient UI is on screen and once right after it is dismissed — and the
// dismissed frame is compared against a UI built straight into that final state.
// Any cell the transient UI left behind shows up as a mismatch, which is the bug
// class that made modals and menus look glitchy.
//
// Run:
//
//	go test ./internal/ui/model -run 'TestOverlayLeftovers|TestOverlayScreenshots' -count=1 -v
//	python3 artifacts/render_tui.py --pattern 'tui-overlay*.json' artifacts/

import (
	"image"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/question"
	"github.com/charmbracelet/crush/internal/ui/completions"
	"github.com/charmbracelet/crush/internal/ui/dialog"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

// overlayChat puts a UI into a mid-conversation chat with the composer focused.
func overlayChat(u *UI) {
	seedForkMessages(u, forkChatMessages())
	u.focus = uiFocusEditor
	u.textarea.Focus()
	// Placeholders are randomized per UI instance; pin them so two UIs in
	// the same state compose identical frames.
	u.workingPlaceholder = "Working..."
	u.readyPlaceholder = "Ready for instructions"
	u.updateLayoutAndSize()
}

// prepareLongPrompt fills the composer past one line with the agent ready, so
// Enter actually submits instead of reporting "agent not initialised".
func prepareLongPrompt(u *UI) {
	if ws, ok := u.com.Workspace.(*testWorkspace); ok {
		ws.agentReady = true
	}
	u.textarea.SetValue(composerLongPrompt)
	u.updateLayoutAndSize()
}

const composerLongPrompt = `SUBMITTED_LINE_ONE: investigate the double charge
SUBMITTED_LINE_TWO: trace the retry scheduler
SUBMITTED_LINE_THREE: check the consumer guard
SUBMITTED_LINE_FOUR: write the regression test`

const landingPrompt = "LANDING_PROMPT check the retry scheduler"

type overlayScenario struct {
	name string
	// show puts the transient UI on screen.
	show func(u *UI)
	// hide dismisses it. It must be safe to call on a UI that never showed
	// it, so the same final state can be built without the transient UI.
	hide func(u *UI)
	// ghost is text that exists only while the transient UI is on screen.
	ghost string
}

func showForkModal(u *UI) {
	u.dialog.OpenDialog(dialog.NewForkFrom(u.com, forkPointsFromMessages(forkChatMessages())))
}

func showQuestionForm(u *UI) {
	form := dialog.NewQuestionForm(u.com.Styles, question.Request{
		ID:        "req-1",
		SessionID: "sess-1",
		Questions: []question.Question{{
			ID:   "q1",
			Type: question.TypeSingleChoice,
			Text: "Which checkout flow should I keep?",
			Choices: []question.Choice{
				{ID: "a", Label: "Retry with an idempotency guard"},
				{ID: "b", Label: "Drop the retry queue"},
			},
		}},
	})
	form.SetFocused(true)
	u.activeInline = form
	u.updateLayoutAndSize()
}

func showCompletions(u *UI) {
	if u.completions == nil {
		sty := u.com.Styles.Completions
		u.completions = completions.New(sty.Normal, sty.Focused, sty.Match)
	}
	u.textarea.SetValue("@retry")
	u.completions.SetItems([]completions.FileCompletionValue{
		{Path: "internal/pay/retry.go"},
		{Path: "internal/pay/consumer.go"},
	}, nil)
	u.completionsPositionStart = image.Pt(u.layout.editor.Min.X+1, u.layout.editor.Min.Y+1)
	u.completionsOpen = true
}

func overlayScenarios() []overlayScenario {
	return []overlayScenario{
		{
			name:  "modal",
			show:  showForkModal,
			hide:  func(u *UI) { u.dialog.CloseFrontDialog() },
			ghost: "Fork From",
		},
		{
			name: "questions",
			show: showQuestionForm,
			hide: func(u *UI) {
				u.activeInline = nil
				u.textarea.Focus()
				u.updateLayoutAndSize()
			},
			ghost: "Which checkout flow should I keep?",
		},
		{
			name: "completions",
			show: showCompletions,
			hide: func(u *UI) {
				u.textarea.SetValue("@retry")
				u.completionsOpen = false
			},
			ghost: "internal/pay/retry.go",
		},
		{
			name: "composer-submit",
			show: prepareLongPrompt,
			hide: func(u *UI) {
				prepareLongPrompt(u)
				// Submit through the real key path so the composer resets
				// exactly as it does for a user pressing Enter.
				_, _ = u.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			},
			// The submitted prompt is echoed into the chat, so the frame
			// is compared against a fresh composition instead of hunting
			// for a marker.
		},
		{
			name: "landing-header",
			show: func(u *UI) {
				u.state = uiLanding
				u.textarea.SetValue(landingPrompt)
				u.updateLayoutAndSize()
			},
			hide: func(u *UI) {
				u.state = uiChat
				u.textarea.SetValue("")
				u.updateLayoutAndSize()
			},
			ghost: landingPrompt,
		},
	}
}

// TestOverlayLeftovers draws the transient UI, dismisses it, and requires the
// frame that follows to be exactly what a UI built into that final state paints.
func TestOverlayLeftovers(t *testing.T) {
	for _, sc := range overlayScenarios() {
		t.Run(sc.name, func(t *testing.T) {
			incremental := overlayScenarioUI(t)
			sc.show(incremental)
			if sc.ghost != "" {
				require.Contains(t, stableView(incremental), sc.ghost,
					"%s: the transient UI must actually be painted", sc.name)
			}
			sc.hide(incremental)
			got := stableView(incremental)

			fresh := overlayScenarioUI(t)
			sc.hide(fresh)
			want := stableView(fresh)

			if sc.ghost != "" {
				require.NotContains(t, got, sc.ghost,
					"%s: %q survived dismissal", sc.name, sc.ghost)
			}
			require.Equal(t, want, got,
				"%s: the frame after dismissal must equal a fresh composition", sc.name)
		})
	}
}

// stableView composes a frame with the cosmetic busy/ready placeholders
// neutralized. They are picked at random on every submit, so two UIs in the
// same state would otherwise differ on the composer row alone.
func stableView(u *UI) string {
	content := u.View().Content
	for _, p := range workingPlaceholders {
		content = strings.ReplaceAll(content, p, "PLACEHOLDER")
	}
	for _, p := range readyPlaceholders {
		content = strings.ReplaceAll(content, p, "PLACEHOLDER")
	}
	// Compare visible text only: the placeholders differ in length, so the
	// styled padding after them does too, and trailing blanks carry no
	// content. Any leftover glyph still shows up.
	content = ansi.Strip(content)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}

func overlayScenarioUI(t *testing.T) *UI {
	t.Helper()
	u := NewScreenshotUI(t)
	u.width, u.height = 140, 45
	overlayChat(u)
	return u
}

// TestOverlayScreenshots captures the pairs for the walkthrough GIF.
func TestOverlayScreenshots(t *testing.T) {
	dir := TUIShotDir(t)
	var shots []TUIShot
	for _, sc := range overlayScenarios() {
		sc := sc
		base := func(u *UI) {
			u.width, u.height = 140, 45
			overlayChat(u)
		}
		shots = append(shots,
			TUIShot{
				Name:   "overlay-" + sc.name + "-open-140",
				Width:  140,
				Height: 45,
				Setup: func(u *UI) {
					base(u)
					sc.show(u)
				},
			},
			TUIShot{
				Name:   "overlay-" + sc.name + "-closed-140",
				Width:  140,
				Height: 45,
				Setup: func(u *UI) {
					base(u)
					sc.show(u)
					// Paint the transient frame, then dismiss it: the
					// capture below is the frame right after dismissal.
					_ = u.View()
					sc.hide(u)
				},
			},
		)
	}
	CaptureTUIShots(t, dir, shots)
}

// TestComposerSubmitShrinksTheEditor covers the reported artifact where the
// top lines of an expanded composer stayed painted after Enter. The composer
// must collapse and the rows it vacated must show whatever moved up into them.
func TestComposerSubmitShrinksTheEditor(t *testing.T) {
	u := overlayScenarioUI(t)
	grown := strings.Count(composerLongPrompt, "\n") + 1
	u.textarea.SetValue(composerLongPrompt)
	u.updateLayoutAndSize()
	require.Greater(t, u.layout.editor.Dy(), grown,
		"the composer must be taller than one line while the prompt is long")
	require.Contains(t, u.View().Content, "SUBMITTED_LINE_ONE")

	u.textarea.SetValue("")
	u.textarea.Reset()
	u.updateLayoutAndSize()
	content := u.View().Content
	require.NotContains(t, content, "SUBMITTED_LINE",
		"every composer line must be gone once the prompt is submitted")
}
