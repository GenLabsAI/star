package model

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/charmbracelet/crush/internal/permission"
)

// TestEnterYeehawWhileAgentBusy pins the step-away flow: submit a task, then
// arm yeehaw while the turn is still running. The session mode is read when
// the turn ends, so arming must not be gated on the agent being idle and
// must not cancel the run in flight.
func TestEnterYeehawWhileAgentBusy(t *testing.T) {
	pinTTLs(t)

	ws := &countingWorkspace{ready: true, agentBusy: true}
	m := newBusyUI(ws)
	warmCaches(m, true)
	require.True(t, m.isAgentBusy(), "the agent must be busy for this test")

	m.enterYeehaw()

	require.True(t, m.yeehawActive, "yeehaw must arm while the agent is still running")
	require.Equal(t, permission.ModeYeehaw, ws.PermissionSessionMode(m.session.ID),
		"arming must put the session in view on yeehaw")
	require.True(t, ws.PermissionSkipRequests(),
		"arming must skip permission requests so the run stays unattended")
	require.True(t, m.yoloModeCached(), "the UI cache must reflect the skipped requests")
	require.Zero(t, ws.cancelCalls, "arming yeehaw must not cancel the run in flight")
	require.Equal(t, uiFocusMain, m.focus, "focus must leave the retired editor")
}

// TestEnterYeehawIdleStillWorks guards the ordinary case against regressions
// from relaxing the busy gate.
func TestEnterYeehawIdleStillWorks(t *testing.T) {
	pinTTLs(t)

	ws := &countingWorkspace{ready: true}
	m := newBusyUI(ws)
	warmCaches(m, false)
	require.False(t, m.isAgentBusy())

	m.enterYeehaw()

	require.True(t, m.yeehawActive, "yeehaw must arm from idle")
	require.Equal(t, permission.ModeYeehaw, ws.PermissionSessionMode(m.session.ID))
}
