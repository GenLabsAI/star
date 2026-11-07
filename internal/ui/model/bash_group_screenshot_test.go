package model

// Visual states for collapsed Bash tool groups. Run with:
//
//	go test ./internal/ui/model -run TestBashGroupScreenshots -count=1 -v
//	python3 artifacts/render_tui.py --pattern 'tui-bash-group*.json' artifacts/

import (
	"encoding/json"
	"testing"

	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
)

type bashGroupCall struct {
	id      string
	command string
	output  string
}

// bashGroupMessages seeds one assistant turn that runs four shell commands,
// plus a result message so the items settle into the grouped summary.
func bashGroupMessages() []message.Message {
	calls := []bashGroupCall{
		{
			id:      "call-1",
			command: "git status --short",
			output:  " M internal/auth/middleware.go\n M internal/auth/middleware_test.go\n?? internal/auth/token.go",
		},
		{
			id:      "call-2",
			command: "go test ./internal/auth/...",
			output: "ok      github.com/charmbracelet/crush/internal/auth/oauth   0.31s\n" +
				"--- FAIL: TestRefreshToken (0.01s)\n" +
				"    middleware_test.go:88:\n" +
				"        Error: Not equal:\n" +
				"        expected: \"Bearer abc123\"\n" +
				"        actual  : \"Bearer \"\n" +
				"    middleware_test.go:91:\n" +
				"        Error: received unexpected status 401\n" +
				"--- FAIL: TestRotateSession (0.02s)\n" +
				"    middleware_test.go:140:\n" +
				"        Error: token was not rotated\n" +
				"        expected a new expiry, got the previous one\n" +
				"    middleware_test.go:141:\n" +
				"        Error: session stayed bound to the old token\n" +
				"        goroutine 42 [running]:\n" +
				"        auth.(*Middleware).Refresh(0xc0001a2000)\n" +
				"            internal/auth/middleware.go:212 +0x1d4\n" +
				"        auth.(*Middleware).ServeHTTP(0xc0001a2000, {0x1f2a3c0, 0xc0000e8000})\n" +
				"            internal/auth/middleware.go:148 +0x96\n" +
				"        net/http.HandlerFunc.ServeHTTP(0x1e2b1c0, {0x1f2a3c0, 0xc0000e8000})\n" +
				"            net/http/server.go:2136 +0x29\n" +
				"        net/http.(*ServeMux).ServeHTTP(0xc00012a000, {0x1f2a3c0, 0xc0000e8000})\n" +
				"            net/http/server.go:2514 +0x1c2\n" +
				"    middleware_test.go:88: retry 1 of 3\n" +
				"    middleware_test.go:88: retry 2 of 3\n" +
				"    middleware_test.go:88: retry 3 of 3\n" +
				"FAIL\n" +
				"FAIL    github.com/charmbracelet/crush/internal/auth    0.58s\n" +
				"FAIL\n" +
				"exit status 1",
		},
		{
			id:      "call-3",
			command: "go build ./...",
			output:  "",
		},
		{
			id:      "call-4",
			command: "gofmt -l internal/auth",
			output:  "internal/auth/middleware.go",
		},
	}

	parts := []message.ContentPart{
		message.TextContent{Text: "Found the refresh bug. I'll patch the middleware and run the checks."},
	}
	results := make([]message.ContentPart, 0, len(calls))
	for _, call := range calls {
		input, err := json.Marshal(tools.BashParams{Command: call.command})
		if err != nil {
			panic(err)
		}
		parts = append(parts, message.ToolCall{
			ID:       call.id,
			Name:     tools.BashToolName,
			Input:    string(input),
			Finished: true,
		})
		results = append(results, message.ToolResult{
			ToolCallID: call.id,
			Name:       tools.BashToolName,
			Content:    call.output,
		})
	}

	return []message.Message{
		{
			ID:        "bash-group-user",
			SessionID: "sess-1",
			Role:      message.User,
			Parts: []message.ContentPart{message.TextContent{
				Text: "Refactor the auth middleware and get the suite green. Keep the transcript quiet - I mostly care about what changed.",
			}},
			CreatedAt: 1,
		},
		{
			ID:        "bash-group-assistant",
			SessionID: "sess-1",
			Role:      message.Assistant,
			Parts:     parts,
			CreatedAt: 2,
		},
		{
			ID:        "bash-group-results",
			SessionID: "sess-1",
			Role:      message.Tool,
			Parts:     results,
			CreatedAt: 3,
		},
	}
}

func seedBashGroup(u *UI) {
	_ = u.setSessionMessages(bashGroupMessages())
	u.state = uiChat
	u.focus = uiFocusMain
	u.chat.Focus()
	u.updateLayoutAndSize()
	// user, assistant, then the single grouped row.
	u.chat.SetSelected(2)
	u.chat.ScrollToSelected()
}

func setupBashGroupCollapsed(u *UI) {
	seedBashGroup(u)
}

func setupBashGroupExpanded(u *UI) {
	seedBashGroup(u)
	u.chat.ToggleExpandedSelectedItem()
	// Expanding from a followed view pins to the bottom; walk back up so the
	// frame shows the group header and the first commands.
	u.chat.ScrollBy(-60)
}

func setupBashGroupScrolled(u *UI) {
	setupBashGroupExpanded(u)
	u.chat.ScrollBy(9)
}

func TestBashGroupScreenshots(t *testing.T) {
	dir := TUIShotDir(t)
	shots := []TUIShot{
		{Name: "bash-group-collapsed-140", Width: 140, Height: 45, Setup: setupBashGroupCollapsed},
		{Name: "bash-group-collapsed-100", Width: 100, Height: 30, Setup: setupBashGroupCollapsed},
		{Name: "bash-group-expanded-140", Width: 140, Height: 45, Setup: setupBashGroupExpanded},
		{Name: "bash-group-scrolled-140", Width: 140, Height: 45, Setup: setupBashGroupScrolled},
	}
	CaptureTUIShots(t, dir, shots)
}
