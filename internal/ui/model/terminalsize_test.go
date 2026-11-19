package model

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

// A resize the terminal never announced would otherwise leave every frame at
// the old width, wrapping rows against the new window.
func TestTerminalSizePollAppliesAMissedResize(t *testing.T) {
	t.Parallel()

	u := newFrameTestUI(t)
	u.width, u.height = 100, 30

	cmd := u.terminalSizePollCmd(terminalSizePollMsg{width: 140, height: 40})
	require.NotNil(t, cmd)
	require.Equal(t, tea.WindowSizeMsg{Width: 140, Height: 40}, cmd(),
		"a size the terminal reports must be applied as a resize")
}

// An unchanged size must not churn the layout every tick.
func TestTerminalSizePollIgnoresAnUnchangedResize(t *testing.T) {
	t.Parallel()

	u := newFrameTestUI(t)
	u.width, u.height = 100, 30

	require.Nil(t, u.terminalSizePollCmd(terminalSizePollMsg{width: 100, height: 30}),
		"an unchanged size must not trigger a resize")
	require.Nil(t, u.terminalSizePollCmd(terminalSizePollMsg{}),
		"a failed size query must not trigger a resize")
}
