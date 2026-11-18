package model

// Double-click is the primary gesture for a tool-run summary; a single click
// must only select the row. Run with:
//
//	go test ./internal/ui/model -run TestChatDoubleClickExpandsToolGroup -count=1 -v

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestChatDoubleClickExpandsToolGroup(t *testing.T) {
	u := NewScreenshotUI(t)
	shotSize(u, 140, 45)
	seedToolGroup(u, t)

	// Locate a row inside the grouped item using the same hit test the mouse
	// handler uses.
	x := 8
	y := -1
	for row := 0; row < u.layout.main.Dy(); row++ {
		if idx, _ := u.chat.list.ItemIndexAtPosition(x, row); idx == 2 {
			y = row
			break
		}
	}
	require.NotEqual(t, -1, y, "group row should be hit-testable")

	// Expanding grows the item and the chat follows the bottom, so assert on
	// the group's render rather than whatever the viewport happens to show.
	rendered := func() string {
		return ansi.Strip(u.chat.list.ItemAt(2).Render(120))
	}

	handled, _ := u.chat.HandleMouseDown(x, y)
	require.True(t, handled)
	require.NotContains(t, rendered(), "Bash git status --short", "single click must not expand the run")

	handled, _ = u.chat.HandleMouseDown(x, y)
	require.True(t, handled)
	require.Contains(t, rendered(), "Bash git status --short", "double click expands the run")
	require.Contains(t, rendered(), "double-click to collapse")

	// The counter resets after a double-click, so the next pair folds it back.
	_, _ = u.chat.HandleMouseDown(x, y)
	_, _ = u.chat.HandleMouseDown(x, y)
	require.NotContains(t, rendered(), "Bash git status --short", "double click collapses the run again")
	require.Contains(t, rendered(), "double-click to expand")
}
