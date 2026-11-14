package model

// Visual states for cloud mode: the CLOUD badge next to the help hints
// and a cloud turn whose ACP frames arrived as native tool calls. Run:
//
//	go test ./internal/ui/model -run TestCloudScreenshots -count=1 -v
//	python3 artifacts/render_tui.py --pattern 'tui-cloud*.json' artifacts/

import (
	"encoding/json"
	"testing"

	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
)

// cloudMessages is a cloud turn exactly as the ACP translation writes
// it: assistant text, the tool call the cloud agent made, and the tool
// result it produced on the cloud machine.
func cloudMessages() []message.Message {
	bashInput, _ := jsonMarshal(tools.BashParams{Command: "uname -a && pwd && ls"})
	viewInput := `{"file_path":"/tmp/cloud-demo/NOTES.md"}`
	toolParts := []message.ContentPart{
		message.TextContent{Text: "I'm on the cloud machine. Let me look around before answering."},
		message.ToolCall{
			ID:       "call-cloud-1",
			Name:     tools.BashToolName,
			Input:    bashInput,
			Finished: true,
		},
		message.ToolCall{
			ID:       "call-cloud-2",
			Name:     tools.ViewToolName,
			Input:    viewInput,
			Finished: true,
		},
	}
	results := []message.ContentPart{
		message.ToolResult{
			ToolCallID: "call-cloud-1",
			Name:       tools.BashToolName,
			Content:    "Linux cloud-runner 6.8.0 #1 SMP x86_64 GNU/Linux\n/home/runner/work/star\nNOTES.md\nsrc",
		},
		message.ToolResult{
			ToolCallID: "call-cloud-2",
			Name:       tools.ViewToolName,
			Content:    "     1  # Cloud workspace\n     2  This machine is not your laptop.\n     3  Prompts run here; the TUI streams the results back.",
		},
	}
	return []message.Message{
		{
			ID:        "cloud-user",
			SessionID: "sess-1",
			Role:      message.User,
			Parts: []message.ContentPart{message.TextContent{
				Text: "Run a quick check on the cloud machine and tell me what you see.",
			}},
			CreatedAt: 1,
		},
		{
			ID:        "cloud-assistant",
			SessionID: "sess-1",
			Role:      message.Assistant,
			Parts:     toolParts,
			CreatedAt: 2,
		},
		{
			ID:        "cloud-results",
			SessionID: "sess-1",
			Role:      message.Tool,
			Parts:     results,
			CreatedAt: 3,
		},
	}
}

func jsonMarshal(v any) (string, error) {
	data, err := json.Marshal(v)
	return string(data), err
}

func seedCloudTurn(u *UI) {
	_ = u.setSessionMessages(cloudMessages())
	u.state = uiChat
	u.focus = uiFocusMain
	u.chat.Focus()
	u.updateLayoutAndSize()
	u.chat.SetSelected(u.chat.Len() - 1)
	u.chat.ScrollToSelected()
}

func TestCloudScreenshots(t *testing.T) {
	dir := TUIShotDir(t)

	cloudEditor := TUIShotSetup(func(u *UI) {
		u.setCloudMode("sess-1", true)
		u.status.SetCloudMode(true)
		u.textarea.SetValue("/cloud")
		u.updateLayoutAndSize()
	})

	cloudRunning := TUIShotSetup(func(u *UI) {
		u.setCloudMode("sess-1", true)
		u.status.SetCloudMode(true)
		u.textarea.SetValue("")
		seedCloudTurn(u)
	})

	cloudExpanded := TUIShotSetup(func(u *UI) {
		cloudRunning(u)
		u.chat.ToggleExpandedSelectedItem()
		u.chat.ScrollBy(-200)
	})

	shots := []TUIShot{
		{Name: "cloud-editor-140", Width: 140, Height: 45, Setup: cloudEditor},
		{Name: "cloud-turn-140", Width: 140, Height: 45, Setup: cloudRunning},
		{Name: "cloud-turn-expanded-140", Width: 140, Height: 45, Setup: cloudExpanded},
		{Name: "cloud-turn-100", Width: 100, Height: 30, Setup: cloudRunning},
	}
	CaptureTUIShots(t, dir, shots)
}
