package model

// Visual states for the collapsed tool-run summary. Run with:
//
//	go test ./internal/ui/model -run TestToolGroupScreenshots -count=1 -v
//	python3 artifacts/render_tui.py --pattern 'tui-tool-group*.json' artifacts/

import (
	"encoding/json"
	"testing"

	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
)

type toolGroupCall struct {
	id, name, input, output string
}

func shellCall(t *testing.T, id, command, output string) toolGroupCall {
	t.Helper()
	input, err := json.Marshal(tools.BashParams{Command: command})
	if err != nil {
		panic(err)
	}
	return toolGroupCall{id: id, name: tools.BashToolName, input: string(input), output: output}
}

// toolGroupMessages is a realistic run: commands and reads and searches fold
// into one summary, the edit breaks it, and the build checks start a new one.
func toolGroupMessages(t *testing.T) []message.Message {
	t.Helper()
	calls := []toolGroupCall{
		shellCall(t, "call-1", "git status --short",
			" M internal/auth/middleware.go\n M internal/auth/middleware_test.go\n?? internal/auth/token.go"),
		shellCall(t, "call-2", "go test ./internal/auth/...",
			"ok      github.com/charmbracelet/crush/internal/auth/oauth   0.31s\n"+
				"--- FAIL: TestRefreshToken (0.01s)\n"+
				"    middleware_test.go:88: Error: Not equal:\n"+
				"        expected: \"Bearer abc123\"\n"+
				"        actual  : \"Bearer \"\n"+
				"--- FAIL: TestRotateSession (0.02s)\n"+
				"    middleware_test.go:140: Error: token was not rotated\n"+
				"FAIL\nFAIL    github.com/charmbracelet/crush/internal/auth    0.58s\nFAIL\nexit status 1"),
		{
			id: "call-3", name: tools.ViewToolName,
			input:  `{"file_path":"internal/auth/middleware.go","offset":180,"limit":60}`,
			output: "   180  func (m *Middleware) Refresh(ctx context.Context, sess *Session) error {\n   181      token, err := m.tokens.Rotate(ctx, sess.Token)\n   182      if err != nil {\n   183          return fmt.Errorf(\"rotate token: %w\", err)\n   ...\n   212      sess.Token = token.Value\n   213      return m.sessions.Save(ctx, sess)",
		},
		{
			id: "call-4", name: tools.ViewToolName,
			input:  `{"file_path":"internal/auth/middleware_test.go","offset":70,"limit":80}`,
			output: "    70  func TestRefreshToken(t *testing.T) {\n    71      sess := newSession(\"Bearer abc123\")\n    72      require.NoError(t, mw.Refresh(ctx, sess))\n    88      require.Equal(t, \"Bearer abc123\", sess.Token)\n   140      require.NotEqual(t, oldExpiry, sess.Expiry)",
		},
		{
			id: "call-5", name: tools.WebSearchToolName,
			input:  `{"query":"oauth refresh token rotation avoid reusing old token"}`,
			output: "1. OAuth 2.0 Security Best Current Practice - Refresh token rotation\n   https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics\n2. Rotating refresh tokens without racing tabs\n   https://auth0.com/blog/refresh-token-rotation",
		},
		{
			id: "call-6", name: tools.SourcegraphToolName,
			input:  `{"query":"repo:GenLabsAI/star refresh token rotation"}`,
			output: "internal/auth/middleware.go:212  sess.Token = token.Value\ninternal/auth/oauth/rotate.go:54  next, err := rotate(old)",
		},
		{
			id: "call-7", name: tools.EditToolName,
			input:  `{"file_path":"internal/auth/middleware.go","old_string":"sess.Token = token.Value","new_string":"sess.Token = token.Value\n\tsess.Expiry = token.Expiry"}`,
			output: "Applied 1 edit to internal/auth/middleware.go",
		},
		shellCall(t, "call-8", "go build ./...", ""),
		shellCall(t, "call-9", "gofmt -l internal/auth", "internal/auth/middleware.go"),
	}

	parts := []message.ContentPart{
		message.TextContent{Text: "I'll trace the refresh path, check prior art, then patch it."},
	}
	results := make([]message.ContentPart, 0, len(calls))
	for _, call := range calls {
		parts = append(parts, message.ToolCall{
			ID:       call.id,
			Name:     call.name,
			Input:    call.input,
			Finished: true,
		})
		results = append(results, message.ToolResult{
			ToolCallID: call.id,
			Name:       call.name,
			Content:    call.output,
		})
	}

	return []message.Message{
		{
			ID:        "tool-group-user",
			SessionID: "sess-1",
			Role:      message.User,
			Parts: []message.ContentPart{message.TextContent{
				Text: "Fix the refresh flow and keep the transcript quiet - I mostly care about what changed.",
			}},
			CreatedAt: 1,
		},
		{
			ID:        "tool-group-assistant",
			SessionID: "sess-1",
			Role:      message.Assistant,
			Parts:     parts,
			CreatedAt: 2,
		},
		{
			ID:        "tool-group-results",
			SessionID: "sess-1",
			Role:      message.Tool,
			Parts:     results,
			CreatedAt: 3,
		},
	}
}

func seedToolGroup(u *UI, t *testing.T) {
	t.Helper()
	_ = u.setSessionMessages(toolGroupMessages(t))
	u.state = uiChat
	u.focus = uiFocusMain
	u.chat.Focus()
	u.updateLayoutAndSize()
	// user, assistant, then the first summary row.
	u.chat.SetSelected(2)
	u.chat.ScrollToSelected()
}

func setupToolGroupCollapsed(u *UI, t *testing.T) {
	seedToolGroup(u, t)
}

func setupToolGroupExpanded(u *UI, t *testing.T) {
	seedToolGroup(u, t)
	u.chat.ToggleExpandedSelectedItem()
	u.chat.ScrollBy(-200)
}

func setupToolGroupScrolled(u *UI, t *testing.T) {
	setupToolGroupExpanded(u, t)
	u.chat.ScrollBy(12)
}

func TestToolGroupScreenshots(t *testing.T) {
	dir := TUIShotDir(t)
	shots := []TUIShot{
		{Name: "tool-group-collapsed-140", Width: 140, Height: 45, Setup: func(u *UI) { setupToolGroupCollapsed(u, t) }},
		{Name: "tool-group-collapsed-100", Width: 100, Height: 30, Setup: func(u *UI) { setupToolGroupCollapsed(u, t) }},
		{Name: "tool-group-expanded-140", Width: 140, Height: 45, Setup: func(u *UI) { setupToolGroupExpanded(u, t) }},
		{Name: "tool-group-scrolled-140", Width: 140, Height: 45, Setup: func(u *UI) { setupToolGroupScrolled(u, t) }},
	}
	CaptureTUIShots(t, dir, shots)
}
