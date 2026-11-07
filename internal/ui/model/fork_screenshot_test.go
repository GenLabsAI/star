package model

// Visual states for the fork affordance and Fork From dialog. Run with:
//
//	go test ./internal/ui/model -run TestForkScreenshots -count=1 -v
//	python3 artifacts/render_tui.py --pattern 'tui-fork*.json' --font /path/to/monospace.ttf
//
// The frames are written as per-cell JSON grids under artifacts/ and rendered
// to PNGs by the shared harness.

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/charmbracelet/crush/internal/ui/dialog"
)

func forkTestMessage(id string, role message.MessageRole, text string, createdAt int64) message.Message {
	return message.Message{
		ID:        id,
		SessionID: "sess-1",
		Role:      role,
		Parts:     []message.ContentPart{message.TextContent{Text: text}},
		CreatedAt: createdAt,
	}
}

// forkChatMessages is a short, realistic exchange used for the in-chat Fork
// button shot.
func forkChatMessages() []message.Message {
	return []message.Message{
		forkTestMessage("11111111-1111-4111-8111-111111111111", message.User,
			"Our checkout retries are double-charging customers. Can you trace where the retry is scheduled?", 1),
		forkTestMessage("22222222-2222-4222-8222-222222222222", message.Assistant,
			"I'll start by looking at the payment service and the retry scheduler. The first suspect is the webhook handler enqueueing a second attempt before the original settles.", 2),
		forkTestMessage("33333333-3333-4333-8333-333333333333", message.User,
			"The duplicate charges all share the same idempotency key prefix. Check the queue consumer too.", 3),
		forkTestMessage("44444444-4444-4444-8444-444444444444", message.Assistant,
			"Found it: the consumer re-enqueues on any timeout, including attempts that already committed. I'll add an idempotency guard keyed by the charge ID.", 4),
		forkTestMessage("55555555-5555-4555-8555-555555555555", message.User,
			"Good. Add a regression test with a simulated timeout after commit.", 5),
		forkTestMessage("66666666-6666-4666-8666-666666666666", message.Assistant,
			"Added the guard and a test that forces the post-commit timeout. The charging path now sees the recorded charge and skips the retry. All checkout tests pass.", 6),
	}
}

// forkLongMessages gives the Fork From dialog enough rows to scroll.
func forkLongMessages() []message.Message {
	msgs := make([]message.Message, 0, 24)
	for i := 0; i < 12; i++ {
		userID := fmt.Sprintf("%08d-1111-4111-8111-%012d", i*2+1, i*2+1)
		assistantID := fmt.Sprintf("%08d-2222-4222-8222-%012d", i*2+2, i*2+2)
		msgs = append(msgs,
			forkTestMessage(userID, message.User,
				fmt.Sprintf("Checkpoint %d: investigate the checkout retry path.", i+1), int64(i*2+1)),
			forkTestMessage(assistantID, message.Assistant,
				fmt.Sprintf("Checkpoint %d: traced the retry to the payment consumer and wrote down the sequence.", i+1), int64(i*2+2)),
		)
	}
	return msgs
}

func seedForkMessages(u *UI, msgs []message.Message) {
	_ = u.setSessionMessages(msgs)
	u.state = uiChat
	u.focus = uiFocusMain
	u.chat.Focus()
	u.updateLayoutAndSize()
}

func setupForkButton(u *UI) {
	seedForkMessages(u, forkChatMessages())
	// Select the last assistant message (user/assistant alternate).
	u.chat.SetSelected(5)
	u.chat.ScrollToSelected()
}

func setupForkCommand(u *UI) {
	seedForkMessages(u, forkChatMessages())
	u.com.Config().Options.TUI = &config.TUIOptions{}
	commands, err := dialog.NewCommands(u.com, "sess-1", true, false, false, nil, nil)
	if err != nil {
		panic(err)
	}
	u.dialog.OpenDialog(commands)
	// The first draw rebuilds the command list and clears its filter; do it
	// before typing so the capture keeps the query and filtered result.
	_ = u.View()
	for _, r := range "fork" {
		commands.HandleMsg(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

func setupForkDialog(u *UI, selected int) {
	msgs := forkLongMessages()
	seedForkMessages(u, msgs)
	dlg := dialog.NewForkFrom(u.com, forkPointsFromMessages(msgs))
	if selected > 0 {
		dlg.Select(selected)
	}
	u.dialog.OpenDialog(dlg)
}

func setupForkDialogNewest(u *UI) { setupForkDialog(u, 0) }
func setupForkDialogOlder(u *UI)  { setupForkDialog(u, 12) }

func setupForkedSession(u *UI) {
	msgs := forkChatMessages()[:4]
	u.session = &session.Session{ID: "sess-fork", Title: "Fork of Checkout retries"}
	seedForkMessages(u, msgs)
	// The fork result is the subject here; hide the action pill so it does not
	// sit over the copied transcript.
	u.chat.Blur()
}

func TestForkScreenshots(t *testing.T) {
	dir := TUIShotDir(t)
	shots := []TUIShot{
		{Name: "fork-button-140", Width: 140, Height: 45, Setup: setupForkButton},
		{Name: "fork-button-100", Width: 100, Height: 30, Setup: setupForkButton},
		{Name: "fork-command-140", Width: 140, Height: 45, Setup: setupForkCommand},
		{Name: "fork-dialog-140", Width: 140, Height: 45, Setup: setupForkDialogNewest},
		{Name: "fork-dialog-scrolled-140", Width: 140, Height: 45, Setup: setupForkDialogOlder},
		{Name: "forked-session-140", Width: 140, Height: 45, Setup: setupForkedSession},
	}
	CaptureTUIShots(t, dir, shots)
}
