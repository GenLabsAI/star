package model

import (
	"image"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"

	"github.com/charmbracelet/crush/internal/permission"
	"github.com/charmbracelet/crush/internal/session"
)

// yeehawProbe delivers a busy-state probe the way applyBusyState would
// receive one: the mode reported belongs to forSession.
func yeehawProbe(m *UI, sessionID string, mode permission.Mode) {
	m.applyBusyState(busyStateMsg{
		gen:        m.busyFetchGen,
		forSession: sessionID,
		ready:      true,
		agentBusy:  true,
		mode:       mode,
	})
}

// TestYeehawFollowsSessionSwitch: the horse and badge belong to the session
// in view, not to the UI. Leaving a yeehaw session hides them, re-opening it
// brings them back, and a probe about a different session has no say.
func TestYeehawFollowsSessionSwitch(t *testing.T) {
	pinTTLs(t)

	ws := &countingWorkspace{
		ready:     true,
		agentBusy: true,
		sessionModes: map[string]permission.Mode{
			"A": permission.ModeYeehaw,
			"B": permission.ModeNormal,
		},
	}
	m := newBusyUI(ws)
	m.session = &session.Session{ID: "A"}
	warmCaches(m, true)

	yeehawProbe(m, "A", permission.ModeYeehaw)
	require.True(t, m.yeehawActive, "a session on autopilot must show the horse")
	require.Equal(t, uiFocusMain, m.focus, "the retired editor must not keep focus")

	// Switch to a normal session: the horse must not follow.
	m.session = &session.Session{ID: "B"}
	yeehawProbe(m, "B", permission.ModeNormal)
	require.False(t, m.yeehawActive, "the horse must not show on a normal session")
	require.Equal(t, uiFocusEditor, m.focus, "a normal session must restore the editor")

	// A probe about another session must not drive the view.
	yeehawProbe(m, "A", permission.ModeYeehaw)
	require.False(t, m.yeehawActive, "only the session in view may show the horse")

	// Re-opening the yeehaw session restores it, without replaying the fade.
	m.session = &session.Session{ID: "A"}
	yeehawProbe(m, "A", permission.ModeYeehaw)
	require.True(t, m.yeehawActive, "re-opening a yeehaw session must bring the horse back")
	require.Zero(t, m.yeehawFadeFrames, "re-entering must not replay the fade")
	require.Equal(t, uiFocusMain, m.focus)
}

// TestDoubleEscOnOtherSessionLeavesYeehawAlone: stopping the session in view
// must not cancel or demote the session that is still on autopilot.
func TestDoubleEscOnOtherSessionLeavesYeehawAlone(t *testing.T) {
	pinTTLs(t)

	ws := &countingWorkspace{
		ready:     true,
		agentBusy: true,
		sessionModes: map[string]permission.Mode{
			"A": permission.ModeYeehaw,
			"B": permission.ModeNormal,
		},
	}
	m := newBusyUI(ws)
	m.session = &session.Session{ID: "A"}
	warmCaches(m, true)
	yeehawProbe(m, "A", permission.ModeYeehaw)
	require.True(t, m.yeehawActive)

	// Switch to B and stop it with double esc.
	m.session = &session.Session{ID: "B"}
	yeehawProbe(m, "B", permission.ModeNormal)
	require.False(t, m.yeehawActive, "B is a normal session")

	m.cancelAgent()
	m.cancelAgent()

	require.Equal(t, 1, ws.cancelCalls, "double esc must cancel the session in view")
	require.Equal(t, permission.ModeYeehaw, ws.sessionModes["A"],
		"stopping another session must leave A's autopilot mode alone")
	require.Equal(t, permission.ModeNormal, ws.sessionModes["B"],
		"the session in view must not be promoted to yeehaw")
	require.False(t, m.yeehawActive, "B must not inherit the horse")
}

// TestLeavingYeehawClearsTheStaleHorseCells: the frame canvas is reused across
// frames, so the rows the horse lane occupied are never painted again once the
// lane collapses. Leaving yeehaw must drop the buffer (and any memoized frame)
// or the horse stays stamped on the screen of the session we switched to.
func TestLeavingYeehawClearsTheStaleHorseCells(t *testing.T) {
	pinTTLs(t)

	ws := &countingWorkspace{
		ready:     true,
		agentBusy: true,
		sessionModes: map[string]permission.Mode{
			"A": permission.ModeYeehaw,
			"B": permission.ModeNormal,
		},
	}
	m := newBusyUI(ws)
	m.session = &session.Session{ID: "A"}
	warmCaches(m, true)
	yeehawProbe(m, "A", permission.ModeYeehaw)
	require.True(t, m.yeehawActive)

	// Paint the lane the way a real frame does, into the reusable canvas.
	m.width, m.height = 100, 30
	canvas := m.frameCanvas()
	uv.NewStyledString(m.yeehawHorseView(90)).Draw(canvas, image.Rect(0, 20, 100, 27))
	require.Contains(t, canvas.Render(), yeehawBadgeText, "the horse lane must be painted")

	// Switch to a normal session: the horse must leave the screen.
	m.session = &session.Session{ID: "B"}
	yeehawProbe(m, "B", permission.ModeNormal)
	require.False(t, m.yeehawActive)
	require.NotContains(t, m.frameCanvas().Render(), yeehawBadgeText)
}
