package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/stretchr/testify/require"
)

type wakeupHarness struct {
	wakeup    *WakeupScheduler
	scheduled *mockWakeupScheduledPublisher
	fired     *mockWakeupPublisher
	canceled  *mockWakeupCanceledPublisher

	mu       sync.Mutex
	executed []WakeupSchedule
}

func newWakeupHarness(store WakeupStore) *wakeupHarness {
	h := &wakeupHarness{
		scheduled: &mockWakeupScheduledPublisher{},
		fired:     &mockWakeupPublisher{},
		canceled:  &mockWakeupCanceledPublisher{},
	}
	h.wakeup = NewWakeupSchedulerWithStore(h.fired, h.scheduled, h.canceled, store)
	h.wakeup.SetExecutor(func(_ context.Context, sched WakeupSchedule) error {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.executed = append(h.executed, sched)
		return nil
	})
	return h
}

func (h *wakeupHarness) executedCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.executed)
}

func TestWakeupPersistsAcrossRestarts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	first := newWakeupHarness(NewWakeupStore(dir))
	result, err := first.wakeup.ScheduleAdvanced(context.Background(), tools.ScheduleRequest{
		SessionID: "sess-1",
		Delay:     60 * time.Second,
		Reason:    "later",
		Prompt:    "wake up",
	})
	require.NoError(t, err)
	first.wakeup.Close()

	// A fresh scheduler over the same data dir sees the schedule and
	// fires it via RunDue when the time comes.
	second := newWakeupHarness(NewWakeupStore(dir))
	schedules, err := second.wakeup.List(context.Background())
	require.NoError(t, err)
	require.Len(t, schedules, 1)
	require.Equal(t, result.Schedule.ID, schedules[0].ID)

	fired, err := second.wakeup.RunDue(context.Background())
	require.NoError(t, err)
	require.Empty(t, fired, "not due yet")

	// Event schedules survive restarts too and fire on trigger.
	_, err = second.wakeup.ScheduleAdvanced(context.Background(), tools.ScheduleRequest{
		SessionID: "sess-1",
		Event:     "ci-failed",
		Prompt:    "bisect it",
	})
	require.NoError(t, err)
	n, err := second.wakeup.FireEvent(context.Background(), "ci-failed", "", "")
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Equal(t, 1, second.executedCount())
	second.wakeup.Close()
}

func TestWakeupRepeatAdvances(t *testing.T) {
	t.Parallel()

	h := newWakeupHarness(NewMemoryWakeupStore())
	defer h.wakeup.Close()
	result, err := h.wakeup.ScheduleAdvanced(context.Background(), tools.ScheduleRequest{
		SessionID:     "sess-1",
		RepeatSeconds: 3600,
		Prompt:        "check again",
	})
	require.NoError(t, err)
	require.Contains(t, result.Message, "every 1h0m0s")

	// Claiming at the fire advances the stored schedule instead of
	// removing it.
	first := time.Unix(result.Schedule.NextFireUnix, 0)
	claimed, ok, err := h.wakeup.store.ClaimDue(context.Background(), result.Schedule.ID, first)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, first.Add(time.Hour).Unix(), claimed.NextFireUnix)

	schedules, err := h.wakeup.List(context.Background())
	require.NoError(t, err)
	require.Len(t, schedules, 1)
}

func TestWakeupOneShotClaimedOnce(t *testing.T) {
	t.Parallel()

	h := newWakeupHarness(NewMemoryWakeupStore())
	defer h.wakeup.Close()
	_, err := h.wakeup.ScheduleAdvanced(context.Background(), tools.ScheduleRequest{
		SessionID: "sess-1",
		Delay:     10 * time.Millisecond,
		Prompt:    "once",
	})
	require.NoError(t, err)

	deadline := time.Now().Add(2 * time.Second)
	for h.executedCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	require.Equal(t, 1, h.executedCount())

	// The timer re-arms harmlessly after the claim, but there is
	// nothing left to claim: no double fire.
	fired, err := h.wakeup.RunDue(context.Background())
	require.NoError(t, err)
	require.Empty(t, fired)
	schedules, err := h.wakeup.List(context.Background())
	require.NoError(t, err)
	require.Empty(t, schedules)
}

func TestWakeupBusyRearmsOneShot(t *testing.T) {
	t.Parallel()

	h := newWakeupHarness(NewMemoryWakeupStore())
	defer h.wakeup.Close()
	h.wakeup.SetExecutor(func(_ context.Context, _ WakeupSchedule) error {
		return ErrWakeupBusy
	})
	schedBefore, err := h.wakeup.ScheduleAdvanced(context.Background(), tools.ScheduleRequest{
		SessionID: "sess-1",
		Delay:     10 * time.Millisecond,
		Prompt:    "once",
	})
	require.NoError(t, err)
	_ = schedBefore
	before := schedBefore.Schedule.NextFireUnix

	deadline := time.Now().Add(2 * time.Second)
	for len(h.scheduled.events) < 2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	require.Len(t, h.scheduled.events, 2, "busy one-shot re-arms with a new scheduled event")

	schedules, err := h.wakeup.List(context.Background())
	require.NoError(t, err)
	require.Len(t, schedules, 1)
	require.Greater(t, schedules[0].NextFireUnix, before)
}

func TestWakeupCancelRemovesDurable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	h := newWakeupHarness(NewWakeupStore(dir))
	defer h.wakeup.Close()
	_, err := h.wakeup.ScheduleAdvanced(context.Background(), tools.ScheduleRequest{
		SessionID: "sess-1",
		Event:     "ci-failed",
		Prompt:    "bisect",
	})
	require.NoError(t, err)
	h.wakeup.Cancel("sess-1")

	schedules, err := h.wakeup.List(context.Background())
	require.NoError(t, err)
	require.Empty(t, schedules)

	// A restart sees nothing either.
	restarted := newWakeupHarness(NewWakeupStore(dir))
	defer restarted.wakeup.Close()
	schedules, err = restarted.wakeup.List(context.Background())
	require.NoError(t, err)
	require.Empty(t, schedules)
}

func TestWakeupValidation(t *testing.T) {
	t.Parallel()

	h := newWakeupHarness(NewMemoryWakeupStore())
	defer h.wakeup.Close()
	ctx := context.Background()

	for _, req := range []tools.ScheduleRequest{
		{Delay: time.Minute, Prompt: "x"},                                    // no session
		{SessionID: "s", Delay: time.Minute},                                 // no prompt
		{SessionID: "s", Prompt: "x", RepeatSeconds: 30},                     // repeat too short
		{SessionID: "s", Prompt: "x", Cron: "0 9 * * *", Delay: time.Minute}, // cron + delay
		{SessionID: "s", Prompt: "x", Cron: "banana"},                        // bad cron
		{SessionID: "s", Prompt: "x", DailyAt: "25:00"},                      // bad daily time
		{SessionID: "s", Prompt: "x", Cron: "0 9 * * *", DailyAt: "09:00"},   // cron + daily
		{SessionID: "s", Prompt: "x", Event: "e", Cron: "0 9 * * *", DailyAt: "09:00"},
		{SessionID: "s", Event: "e"}, // event without prompt still needs one
	} {
		_, err := h.wakeup.ScheduleAdvanced(ctx, req)
		require.Error(t, err, "req %+v", req)
	}
}

func TestWakeupLongHorizonSurvivesProcessDeath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dir := t.TempDir()

	// An agent schedules a wakeup 24 hours out: far past any single
	// job's lifetime. The scheduling process then dies outright.
	first := newWakeupHarness(NewWakeupStore(dir))
	result, err := first.wakeup.ScheduleAdvanced(ctx, tools.ScheduleRequest{
		SessionID: "sess-1",
		Delay:     24 * time.Hour,
		Reason:    "daily triage",
		Prompt:    "morning triage",
	})
	require.NoError(t, err)
	first.wakeup.Close()

	// A fresh short-lived sweeper (what a cron entry runs every few
	// minutes) sees the schedule but correctly fires nothing yet.
	sweeper := newWakeupHarness(NewWakeupStore(dir))
	fired, err := sweeper.wakeup.RunDue(ctx)
	require.NoError(t, err)
	require.Empty(t, fired)
	require.Equal(t, 0, sweeper.executedCount())
	sweeper.wakeup.Close()

	// The next sweep starts with no knowledge of the past: a fresh
	// scheduler loads the (still future) schedule and arms its timer.
	after := newWakeupHarness(NewWakeupStore(dir))
	defer after.wakeup.Close()

	// Twenty-four hours pass with no Star process alive anywhere.
	// Simulate the elapsed time by backdating the stored fire.
	store := NewWakeupStore(dir)
	schedules, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, schedules, 1)
	schedules[0].NextFireUnix = time.Now().Add(-time.Minute).Unix()
	require.NoError(t, store.Save(ctx, schedules[0]))

	// A short-lived sweep claims and fires it exactly once.
	fired, err = after.wakeup.RunDue(ctx)
	require.NoError(t, err)
	require.Len(t, fired, 1)
	require.Equal(t, result.Schedule.ID, fired[0].ID)
	require.Equal(t, 1, after.executedCount())

	// And it does not fire again on the following sweep.
	fired, err = after.wakeup.RunDue(ctx)
	require.NoError(t, err)
	require.Empty(t, fired)
}

func TestWakeupRunAfterCloseReapsOverdue(t *testing.T) {
	t.Parallel()

	// Mirrors the one-shot CLI commands (`wakeup run`): the fresh
	// dispatcher disarms constructor timers first so no catch-up
	// timer can steal the claim and die with the exiting process
	// mid-turn. Reaping is explicit and exactly-once.
	ctx := context.Background()
	dir := t.TempDir()

	first := newWakeupHarness(NewWakeupStore(dir))
	result, err := first.wakeup.ScheduleAdvanced(ctx, tools.ScheduleRequest{
		SessionID: "sess-1",
		Delay:     time.Hour,
		Prompt:    "later",
	})
	require.NoError(t, err)
	first.wakeup.Close()

	store := NewWakeupStore(dir)
	schedules, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, schedules, 1)
	schedules[0].NextFireUnix = time.Now().Add(-time.Minute).Unix()
	require.NoError(t, store.Save(ctx, schedules[0]))

	second := newWakeupHarness(NewWakeupStore(dir))
	second.wakeup.Close()
	fired, err := second.wakeup.RunDue(ctx)
	require.NoError(t, err)
	require.Len(t, fired, 1)
	require.Equal(t, result.Schedule.ID, fired[0].ID)
	require.Equal(t, 1, second.executedCount())
}

func TestWakeupArmStoredFiresCatchUp(t *testing.T) {
	t.Parallel()

	// Long-lived hosts arm explicitly after init: an overdue stored
	// schedule fires via its catch-up timer without any RunDue sweep.
	ctx := context.Background()
	dir := t.TempDir()

	first := newWakeupHarness(NewWakeupStore(dir))
	_, err := first.wakeup.ScheduleAdvanced(ctx, tools.ScheduleRequest{
		SessionID: "sess-1",
		Delay:     time.Hour,
		Prompt:    "catch up",
	})
	require.NoError(t, err)
	first.wakeup.Close()

	// The fire time passes with no host around.
	store := NewWakeupStore(dir)
	schedules, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, schedules, 1)
	schedules[0].NextFireUnix = time.Now().Add(-time.Minute).Unix()
	require.NoError(t, store.Save(ctx, schedules[0]))

	// A fresh host process arms stored timers and the fire arrives
	// via catch-up without any RunDue sweep.
	second := newWakeupHarness(NewWakeupStore(dir))
	defer second.wakeup.Close()
	require.NoError(t, second.wakeup.ArmStored(ctx))

	deadline := time.Now().Add(2 * time.Second)
	for second.executedCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	require.Equal(t, 1, second.executedCount())
}
