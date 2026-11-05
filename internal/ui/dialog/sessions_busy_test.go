package dialog

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/charmbracelet/crush/internal/session"
	"github.com/charmbracelet/crush/internal/ui/styles"
)

// TestRenderItemBusyGlyphLeftOfTitle: the busy indicator is drawn to the left
// of the session name (so a background yeehaw run is visible at a glance),
// while the timestamp keeps the right-hand column.
func TestRenderItemBusyGlyphLeftOfTitle(t *testing.T) {
	t.Parallel()

	out := renderItem(ListItemStyles{}, "fix the parser", "2h ago", false, true, 40, nil, nil, "⠹")
	require.Contains(t, out, "⠹", "the busy spinner must render")
	require.Less(t, strings.Index(out, "⠹"), strings.Index(out, "fix the parser"),
		"the busy spinner belongs to the left of the session name")
	require.Less(t, strings.Index(out, "fix the parser"), strings.Index(out, "2h ago"),
		"the timestamp keeps the right-hand column")
}

// TestRenderItemIdleHasNoSpinner: idle sessions stay clean.
func TestRenderItemIdleHasNoSpinner(t *testing.T) {
	t.Parallel()

	out := renderItem(ListItemStyles{}, "fix the parser", "2h ago", false, false, 40, nil, nil, "⠹")
	require.NotContains(t, out, "⠹")
}

// TestSessionItemBusyFlagRedrawsImmediately: flipping the busy flag must drop
// the cached render, otherwise a finished background run keeps its spinner
// until the list is rebuilt.
func TestSessionItemBusyFlagRedrawsImmediately(t *testing.T) {
	t.Parallel()

	items := sessionItems(&styles.Styles{}, sessionsModeNormal, func(string) bool { return false },
		session.Session{ID: "s1", Title: "background yeehaw"})
	item := items[0].(*SessionItem)

	require.NotContains(t, item.Render(40), "⠹", "an idle session has no spinner")

	item.SetBusyGlyph("⠹")
	item.SetBusy(true)
	require.Contains(t, item.Render(40), "⠹", "a busy session shows the spinner")

	item.SetBusy(false)
	require.NotContains(t, item.Render(40), "⠹", "the spinner must disappear as soon as the run ends")
}
