package workspace

import (
	"context"
	"fmt"

	"github.com/charmbracelet/crush/internal/agent"
)

// wakeupDispatcherProvider is implemented by coordinators exposing
// their durable wakeup dispatcher.
type wakeupDispatcherProvider interface {
	WakeupDispatcher() *agent.WakeupScheduler
}

// wakeupExecutorSetter is implemented by coordinators that expose
// their durable wakeup dispatcher. It is asserted, not part of the
// Coordinator interface, so test doubles are unaffected.
type wakeupExecutorSetter interface {
	SetWakeupExecutor(agent.WakeupExecutor)
}

// ArmWakeupTimers arms in-process timers for every stored schedule,
// / including catch-up fires for overdue ones. Call it once in
// long-lived hosts (TUI, ACP, `wakeup serve`) after init. One-shot
// commands (`wakeup run`, `trigger`, `list`) must not arm: they reap
// explicitly, and a catch-up timer in an exiting process would steal
// the claim and die mid-turn.
func (w *AppWorkspace) ArmWakeupTimers(ctx context.Context) error {
	if w.app == nil || w.app.AgentCoordinator == nil {
		return nil
	}
	provider, ok := w.app.AgentCoordinator.(wakeupDispatcherProvider)
	if !ok {
		return nil
	}
	return provider.WakeupDispatcher().ArmStored(ctx)
}

// StartWakeupDispatch routes fired wakeup schedules through this
// workspace: cloud schedules run on the cloud machine, everything
// else runs as a normal local turn. A session that is already busy
// is left alone so a wakeup never cancels or interleaves with an
// in-flight turn; one-shot wakeups re-arm shortly after.
func (w *AppWorkspace) StartWakeupDispatch() {
	if w.app == nil || w.app.AgentCoordinator == nil {
		return
	}
	setter, ok := w.app.AgentCoordinator.(wakeupExecutorSetter)
	if !ok {
		return
	}
	setter.SetWakeupExecutor(func(ctx context.Context, sched agent.WakeupSchedule) error {
		if sched.SessionID == "" {
			return fmt.Errorf("Wakeup has no session")
		}
		if sched.Prompt == "" {
			return fmt.Errorf("Wakeup has no prompt")
		}
		if w.AgentIsSessionBusy(sched.SessionID) {
			return agent.ErrWakeupBusy
		}
		if sched.Cloud {
			return w.AgentRunCloud(ctx, sched.SessionID, sched.Prompt)
		}
		return w.AgentRun(ctx, sched.SessionID, sched.Prompt)
	})
}
