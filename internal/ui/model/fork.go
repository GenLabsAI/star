package model

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/dialog"
	"github.com/charmbracelet/crush/internal/ui/util"
)

// forkPointsLoadedMsg carries the fork points for the dialog. Messages are
// read off the Update goroutine because a long transcript is expensive.
type forkPointsLoadedMsg struct {
	sessionID string
	points    []dialog.ForkPoint
}

// forkCreatedMsg reports that a fork session was created and should be loaded.
type forkCreatedMsg struct {
	sessionID string
}

// forkPreviewLimit bounds the snippet shown per message in the fork dialog.
const forkPreviewLimit = 160

// loadForkPoints reads the current session's transcript and builds the
// newest-first list of messages the user can fork from.
func (m *UI) loadForkPoints(sessionID string) tea.Cmd {
	return func() tea.Msg {
		msgs, err := m.com.Workspace.ListMessages(context.Background(), sessionID)
		if err != nil {
			return util.ReportError(err)()
		}
		return forkPointsLoadedMsg{sessionID: sessionID, points: forkPointsFromMessages(msgs)}
	}
}

// forkPointsFromMessages builds the newest-first list of fork points. Only
// visible user and assistant text is offered; tool results and hidden
// messages belong to the turn around them and are copied with it.
func forkPointsFromMessages(msgs []message.Message) []dialog.ForkPoint {
	points := make([]dialog.ForkPoint, 0, len(msgs))
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		if msg.Role != message.User && msg.Role != message.Assistant {
			continue
		}
		content := msg.Content()
		if content.Hidden {
			continue
		}
		preview := forkPreview(content.Text)
		if preview == "" {
			continue
		}
		points = append(points, dialog.ForkPoint{
			MessageID: msg.ID,
			Role:      forkRoleLabel(msg.Role),
			Preview:   preview,
			Newest:    len(points) == 0,
		})
	}
	return points
}

// forkSession creates the forked session and reports its ID so Update can
// switch to it. The copy runs off the Update goroutine.
func (m *UI) forkSession(sessionID, messageID string) tea.Cmd {
	return func() tea.Msg {
		forked, err := m.com.Workspace.ForkSession(context.Background(), sessionID, messageID)
		if err != nil {
			return util.ReportError(err)()
		}
		return forkCreatedMsg{sessionID: forked.ID}
	}
}

// forkPreview collapses whitespace and truncates a message so the dialog shows
// a readable preamble instead of a wall of markdown.
func forkPreview(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	runes := []rune(text)
	if len(runes) > forkPreviewLimit {
		return string(runes[:forkPreviewLimit-1]) + "…"
	}
	return text
}

func forkRoleLabel(role message.MessageRole) string {
	switch role {
	case message.User:
		return "You"
	case message.Assistant:
		return "Assistant"
	default:
		return string(role)
	}
}
