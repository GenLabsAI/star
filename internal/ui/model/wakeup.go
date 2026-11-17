package model

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/charmbracelet/crush/internal/ui/dialog"
	"github.com/charmbracelet/crush/internal/ui/util"
)

// handleWakeupScheduled surfaces a newly armed wakeup. Timed
// schedules get the countdown toast (with cancel); event-only
// schedules get a quiet info line since there is no time to count
// down to.
func (m *UI) handleWakeupScheduled(ev pubsub.WakeupScheduledEvent) tea.Cmd {
	if ev.EndTime.IsZero() {
		where := "locally"
		if ev.Cloud {
			where = "on the cloud machine"
		}
		return util.ReportInfo("Event wakeup armed for " + ev.Event + " (" + where + ")")
	}
	m.wakeupEndTime = ev.EndTime
	m.dialog.OpenDialog(dialog.NewWakeupToast(m.com, ev.SessionID, ev.Reason, ev.EndTime))
	return nil
}

// closeWakeupToast retires the countdown toast when the wakeup fires
// or is canceled elsewhere.
func (m *UI) closeWakeupToast() {
	m.dialog.CloseDialog(dialog.WakeupToastID)
}
