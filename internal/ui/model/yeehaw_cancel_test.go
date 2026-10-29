package model

import (
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestCancelAgentExitsYeehawMode pins the yeehaw exit story: with the
// textarea retired there is nowhere to type /yeehaw, so stopping the
// agent must also hand the editor back. Covers both the busy case
// (double-esc cancels and exits) and the idle case (a lone esc after
// the work finished just exits).
func TestCancelAgentExitsYeehawMode(t *testing.T) {
	pinTTLs(t)

	newYeehawUI := func(busy bool) (*countingWorkspace, *UI) {
		ws := &countingWorkspace{ready: true, agentBusy: busy}
		m := newBusyUI(ws)
		warmCaches(m, busy)
		m.setYeehawMode(true)
		require.True(t, m.yeehawActive, "yeehaw must engage for the test")
		require.Equal(t, uiFocusMain, m.focus, "yeehaw must park focus off the retired editor")
		return ws, m
	}

	// Busy: double-esc cancels the agent and restores the editor.
	ws, m := newYeehawUI(true)
	require.Zero(t, m.promptQueue)
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.True(t, m.isCanceling, "first esc must arm cancellation")
	require.True(t, m.yeehawActive, "arming must not exit yeehaw yet")
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.Equal(t, 1, ws.cancelCalls, "second esc must cancel the agent")
	require.False(t, m.yeehawActive, "canceling an unattended run must restore the editor")
	require.Equal(t, uiFocusEditor, m.focus, "focus must return to the editor")
	require.False(t, m.yoloModeCached(), "exiting yeehaw must disarm skip-requests")

	// Idle: work already finished, horse still lapping — esc just exits.
	ws, m = newYeehawUI(false)
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.Zero(t, ws.cancelCalls, "idle esc must not cancel anything")
	require.False(t, m.yeehawActive, "idle esc must restore the editor")
	require.Equal(t, uiFocusEditor, m.focus, "focus must return to the editor")
}
