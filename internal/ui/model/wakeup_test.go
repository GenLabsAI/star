package model

import (
	tea "charm.land/bubbletea/v2"
	"testing"
	"time"

	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/charmbracelet/crush/internal/ui/dialog"
	"github.com/stretchr/testify/require"
)

func TestWakeupToastLifecycle(t *testing.T) {
	t.Parallel()

	u := newFrameTestUI(t)

	// A timed schedule opens the countdown toast.
	_, _ = u.Update(pubsub.Event[pubsub.WakeupScheduledEvent]{
		Type: pubsub.CreatedEvent,
		Payload: pubsub.WakeupScheduledEvent{
			SessionID: "sess-1",
			Reason:    "waiting on CI",
			EndTime:   time.Now().Add(time.Minute),
		},
	})
	require.True(t, u.dialog.ContainsDialog(dialog.WakeupToastID))

	// Firing the wakeup retires the toast; the dispatcher runs the
	// turn itself, so the TUI only gets out of the way.
	_, _ = u.Update(pubsub.Event[pubsub.WakeupEvent]{
		Type:    pubsub.UpdatedEvent,
		Payload: pubsub.WakeupEvent{SessionID: "sess-1"},
	})
	require.False(t, u.dialog.ContainsDialog(dialog.WakeupToastID))
}

func TestWakeupEventOnlyArmsQuietly(t *testing.T) {
	t.Parallel()

	u := newFrameTestUI(t)

	// An event-only schedule has no countdown target: no toast, just
	// an info message.
	_, cmd := u.Update(pubsub.Event[pubsub.WakeupScheduledEvent]{
		Type: pubsub.CreatedEvent,
		Payload: pubsub.WakeupScheduledEvent{
			SessionID: "sess-1",
			Event:     "ci-failed",
		},
	})
	require.NotNil(t, cmd)
	require.False(t, u.dialog.ContainsDialog(dialog.WakeupToastID))
}

func TestWakeupToastCancel(t *testing.T) {
	t.Parallel()

	u := newFrameTestUI(t)
	ws := u.com.Workspace.(*testWorkspace)

	_, _ = u.Update(pubsub.Event[pubsub.WakeupScheduledEvent]{
		Type: pubsub.CreatedEvent,
		Payload: pubsub.WakeupScheduledEvent{
			SessionID: "sess-1",
			EndTime:   time.Now().Add(time.Minute),
		},
	})
	require.True(t, u.dialog.ContainsDialog(dialog.WakeupToastID))

	// Esc on the open toast cancels through the dialog action path.
	_, _ = u.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.Equal(t, []string{"sess-1"}, ws.wakeupCancelCalls)
	require.False(t, u.dialog.ContainsDialog(dialog.WakeupToastID))
}
