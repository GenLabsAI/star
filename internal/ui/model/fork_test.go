package model

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/chat"
	"github.com/charmbracelet/crush/internal/ui/dialog"
	"github.com/stretchr/testify/require"
)

func TestForkButtonClickRequestsFork(t *testing.T) {
	u := NewScreenshotUI(t)
	shotSize(u, 140, 45)
	setupForkButton(u)

	_ = u.View()
	area := u.chat.forkButtonArea
	require.False(t, area.Empty(), "focused selected message should expose a Fork button")

	handled, cmd := u.chat.HandleMouseDown(area.Min.X, area.Min.Y)
	require.True(t, handled)
	require.NotNil(t, cmd)
	msg, ok := cmd().(chat.ForkRequestMsg)
	require.True(t, ok)
	require.Equal(t, "66666666-6666-4666-8666-666666666666", msg.MessageID)

	u.chat.Blur()
	_ = u.View()
	require.True(t, u.chat.forkButtonArea.Empty(), "Fork button hides when the chat loses focus")
}

func TestForkButtonForksThroughWorkspace(t *testing.T) {
	u := NewScreenshotUI(t)
	shotSize(u, 140, 45)
	setupForkButton(u)
	_ = u.View()

	area := u.chat.forkButtonArea
	require.False(t, area.Empty())
	handled, clickCmd := u.chat.HandleMouseDown(area.Min.X, area.Min.Y)
	require.True(t, handled)
	require.NotNil(t, clickCmd)

	request, ok := clickCmd().(chat.ForkRequestMsg)
	require.True(t, ok)
	_, forkCmd := u.Update(request)
	require.NotNil(t, forkCmd)

	// Update batches the fork with unrelated refresh commands; run the fork
	// command itself so the stub does not have to implement every subsystem.
	created, ok := u.forkSession("sess-1", request.MessageID)().(forkCreatedMsg)
	require.True(t, ok, "fork command should report the new session")
	require.Equal(t, "forked-session", created.sessionID)

	ws, ok := u.com.Workspace.(*testWorkspace)
	require.True(t, ok)
	require.Equal(t, [][2]string{{"sess-1", "66666666-6666-4666-8666-666666666666"}}, ws.forkCalls)
}

func TestClickingMessageRevealsForkButton(t *testing.T) {
	u := NewScreenshotUI(t)
	shotSize(u, 140, 45)
	seedForkMessages(u, forkChatMessages())
	u.chat.SetSelected(5)
	u.chat.ScrollToSelected()

	// Start in the editor, as the app does before the user clicks the chat.
	u.focus = uiFocusEditor
	_ = u.textarea.Focus()
	u.chat.Blur()
	_ = u.View()

	offset, ok := u.chat.list.ItemViewportOffset(5)
	require.True(t, ok)
	u.Update(tea.MouseClickMsg{
		X:      u.layout.main.Min.X + 4,
		Y:      u.layout.main.Min.Y + offset + 1,
		Button: tea.MouseLeft,
	})
	_ = u.View()

	require.Equal(t, uiFocusMain, u.focus)
	require.True(t, u.chat.Focused())
	require.False(t, u.chat.forkButtonArea.Empty(), "clicking a message should reveal its Fork button")
}

func TestForkPointsFromMessagesNewestFirst(t *testing.T) {
	msgs := []message.Message{
		forkTestMessage("old-user", message.User, "old   question\nwith whitespace", 1),
		forkTestMessage("tool-result", message.Tool, "ignored", 2),
		forkTestMessage("answer", message.Assistant, "the answer", 3),
		forkTestMessage("hidden", message.User, "hidden", 4),
		forkTestMessage("empty", message.Assistant, "", 5),
		forkTestMessage("new-user", message.User, "newest question", 6),
	}
	msgs[3].Parts[0] = message.TextContent{Text: "hidden", Hidden: true}

	points := forkPointsFromMessages(msgs)
	require.Len(t, points, 3)
	require.Equal(t, "new-user", points[0].MessageID)
	require.Equal(t, "You", points[0].Role)
	require.Equal(t, "newest question", points[0].Preview)
	require.True(t, points[0].Newest)
	require.Equal(t, "answer", points[1].MessageID)
	require.Equal(t, "Assistant", points[1].Role)
	require.False(t, points[1].Newest)
	require.Equal(t, "old question with whitespace", points[2].Preview)
}

func TestForkDialogForksSelectedPoint(t *testing.T) {
	u := NewScreenshotUI(t)
	dlg := dialog.NewForkFrom(u.com, []dialog.ForkPoint{
		{MessageID: "newest", Role: "You", Preview: "newest", Newest: true},
		{MessageID: "older", Role: "You", Preview: "older"},
	})

	action := dlg.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Equal(t, dialog.ActionForkSession{MessageID: "newest"}, action)

	dlg.Select(1)
	action = dlg.HandleMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Equal(t, dialog.ActionForkSession{MessageID: "older"}, action)
}
