package dialog

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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
