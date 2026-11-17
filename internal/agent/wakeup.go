package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/pubsub"
)

// ErrWakeupBusy reports that a wakeup fired while its session was
// already running. One-shot schedules re-arm shortly after; repeating
// and cron schedules simply skip the missed occurrence.
var ErrWakeupBusy = errors.New("Wakeup skipped: session is busy")

// ErrWakeupCloudUnsupported reports a cloud wakeup routed to a
// coordinator without a cloud-capable executor (e.g. server mode).
var ErrWakeupCloudUnsupported = errors.New("Cloud wakeups need a local workspace")

// WakeupExecutor runs one claimed schedule: a local agent run, a
// cloud turn, or whatever the host wired up. It runs on the
// schedule's timer goroutine.
type WakeupExecutor func(ctx context.Context, schedule WakeupSchedule) error

// ScheduleResult is the stored schedule plus its human summary.
// The request type lives in tools (see tools.ScheduleRequest) so
// the tool can build it without an import cycle.
type ScheduleResult struct {
	Schedule WakeupSchedule
	Message  string
}

// WakeupScheduler owns durable wakeup schedules: it persists them,
// arms in-process timers, claims due fires atomically so parallel
// Star processes cannot double-fire, and hands claimed schedules to
// the executor. The original one-shot Schedule/Cancel pair keeps its
// signature; ScheduleAdvanced adds repeat, cron, event, and cloud
// routing.
type WakeupScheduler struct {
	pubWakeup          pubsub.Publisher[pubsub.WakeupEvent]
	pubWakeupScheduled pubsub.Publisher[pubsub.WakeupScheduledEvent]
	pubWakeupCanceled  pubsub.Publisher[pubsub.WakeupCanceledEvent]
	store              WakeupStore
	executor           WakeupExecutor

	mu     sync.Mutex
	timers map[string]*time.Timer
}

func NewWakeupScheduler(pubWakeup pubsub.Publisher[pubsub.WakeupEvent], pubWakeupScheduled pubsub.Publisher[pubsub.WakeupScheduledEvent], pubWakeupCanceled pubsub.Publisher[pubsub.WakeupCanceledEvent]) *WakeupScheduler {
	return NewWakeupSchedulerWithStore(pubWakeup, pubWakeupScheduled, pubWakeupCanceled, NewMemoryWakeupStore())
}

// NewWakeupSchedulerWithStore builds a scheduler backed by a durable
// store. Construction arms nothing: long-lived hosts (TUI, ACP,
// server, `wakeup serve`) call ArmStored after init so catch-up timers
// only ever live in processes that outlive them. One-shot commands
// (`wakeup run`, `trigger`, `list`) reap explicitly instead, so an
// overdue schedule can never be claimed by a timer that dies with an
// exiting process mid-turn. The caller owns lifetime: timers stop
// when ctx ends (via Close, or automatically through
// context.AfterFunc).
func NewWakeupSchedulerWithStore(pubWakeup pubsub.Publisher[pubsub.WakeupEvent], pubWakeupScheduled pubsub.Publisher[pubsub.WakeupScheduledEvent], pubWakeupCanceled pubsub.Publisher[pubsub.WakeupCanceledEvent], store WakeupStore) *WakeupScheduler {
	if store == nil {
		store = NewMemoryWakeupStore()
	}
	return &WakeupScheduler{
		pubWakeup:          pubWakeup,
		pubWakeupScheduled: pubWakeupScheduled,
		pubWakeupCanceled:  pubWakeupCanceled,
		store:              store,
		timers:             make(map[string]*time.Timer),
	}
}

// ArmStored arms in-process timers for every timed schedule in the
// store, including catch-up fires for overdue ones. Long-lived hosts
// call it once after init; one-shot commands must not.
func (s *WakeupScheduler) ArmStored(ctx context.Context) error {
	loaded, err := s.store.List(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sched := range loaded {
		if sched.Timed() {
			s.armLocked(sched, now)
		}
	}
	return nil
}

// AttachContext stops all timers when ctx ends. Hosts call it with
// their process/workspace context so no timer outlives shutdown.
func (s *WakeupScheduler) AttachContext(ctx context.Context) {
	context.AfterFunc(ctx, func() { s.Close() })
}

// SetExecutor overrides schedule execution. The default runs a local
// agent turn through runLocal, which hosts set to their coordinator
// run; cloud schedules without a cloud-capable executor fail loudly
// instead of silently never running.
func (s *WakeupScheduler) SetExecutor(executor WakeupExecutor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executor = executor
}

// Schedule keeps the original one-shot contract: a single local
// timer that publishes a wakeup event. It now also persists, so the
// wakeup survives restarts.
func (s *WakeupScheduler) Schedule(sessionID string, delay time.Duration, reason, prompt string) {
	_, err := s.ScheduleAdvanced(context.Background(), tools.ScheduleRequest{
		SessionID: sessionID,
		Delay:     delay,
		Reason:    reason,
		Prompt:    prompt,
	})
	if err != nil {
		slog.Warn("Failed to schedule wakeup", "error", err)
	}
}

// ScheduleAdvanced stores a wakeup and arms it. Timed schedules get
// a timer; event schedules wait for FireEvent; combined schedules do
// both (whichever claims first wins, one-shot timers are consumed).
func (s *WakeupScheduler) ScheduleAdvanced(ctx context.Context, req tools.ScheduleRequest) (ScheduleResult, error) {
	now := time.Now()
	sched, err := resolveSchedule(req, now)
	if err != nil {
		return ScheduleResult{}, err
	}
	if err := s.store.Save(ctx, sched); err != nil {
		return ScheduleResult{}, err
	}
	s.mu.Lock()
	s.armLocked(sched, now)
	s.mu.Unlock()
	s.pubWakeupScheduled.Publish(pubsub.CreatedEvent, pubsub.WakeupScheduledEvent{
		SessionID:  sched.SessionID,
		ScheduleID: sched.ID,
		Reason:     sched.Reason,
		Prompt:     sched.Prompt,
		Event:      sched.Event,
		Cloud:      sched.Cloud,
		Repeat:     sched.Repeat(),
		EndTime:    fireTime(sched),
	})
	return ScheduleResult{Schedule: sched, Message: scheduleMessage(sched)}, nil
}

// Cancel removes every schedule for a session, both persisted and
// armed. It keeps the original single-timer semantics (one canceled
// event when anything was pending).
func (s *WakeupScheduler) Cancel(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()
	removed, err := s.store.DeleteSession(ctx, sessionID)
	if err != nil {
		slog.Warn("Failed to cancel saved wakeups", "error", err)
	}
	disarmed := s.disarmSessionLocked(sessionID)
	if disarmed || removed > 0 {
		s.pubWakeupCanceled.Publish(pubsub.DeletedEvent, pubsub.WakeupCanceledEvent{
			SessionID: sessionID,
		})
	}
}

// CancelSchedule removes one schedule by id.
func (s *WakeupScheduler) CancelSchedule(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.timers[id]; ok {
		t.Stop()
		delete(s.timers, id)
	}
	return s.store.Delete(ctx, id)
}

// RunDue claims and fires every due timed schedule. Cron and systemd
// invocations call this so wakeups fire even when no long-running
// Star process is around to host the timers.
func (s *WakeupScheduler) RunDue(ctx context.Context) ([]WakeupSchedule, error) {
	schedules, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var fired []WakeupSchedule
	for _, sched := range schedules {
		if !sched.Timed() || sched.NextFireUnix > now.Unix() {
			continue
		}
		if claimed, ok := s.fireSchedule(sched.ID, now); ok {
			fired = append(fired, claimed)
		}
	}
	return fired, nil
}

// FireEvent fires every armed schedule for a named event (optionally
// restricted to one session). A prompt override replaces the stored
// prompt for this firing only. It returns the number of schedules
// fired; callers surface a busy session instead of silently dropping
// it.
func (s *WakeupScheduler) FireEvent(ctx context.Context, event, sessionID, prompt string) (int, error) {
	if event == "" {
		return 0, fmt.Errorf("An event name is required")
	}
	schedules, err := s.store.List(ctx)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	fired := 0
	for _, sched := range schedules {
		if sched.Event != event {
			continue
		}
		if sessionID != "" && sched.SessionID != sessionID {
			continue
		}
		if prompt != "" {
			sched.Prompt = prompt
		}
		oneShot := sched.RepeatSeconds == 0 && sched.Cron == ""
		if oneShot {
			// One-shot schedules are consumed by whichever fires
			// first, so a racing timer (or a second process on
			// the same event) cannot run them again.
			consumed, err := s.store.Consume(ctx, sched.ID)
			if err != nil {
				slog.Warn("Failed to consume event wakeup", "schedule", sched.ID, "error", err)
				continue
			}
			if !consumed {
				continue
			}
			s.disarmTimer(sched.ID)
		}
		if err := s.execute(sched); err != nil {
			if errors.Is(err, ErrWakeupBusy) {
				if oneShot {
					// A busy session must not eat the wakeup:
					// put it back (re-armed shortly when timed).
					if sched.Timed() {
						sched.NextFireUnix = now.Add(time.Minute).Unix()
					}
					if serr := s.store.Save(ctx, sched); serr != nil {
						slog.Warn("Failed to restore busy wakeup", "schedule", sched.ID, "error", serr)
					} else if sched.Timed() {
						s.mu.Lock()
						s.armLocked(sched, now)
						s.mu.Unlock()
					}
				}
				return fired, fmt.Errorf("Wakeup for session %s: %w", sched.SessionID, err)
			}
			slog.Warn("Event wakeup failed", "schedule", sched.ID, "error", err)
			continue
		}
		fired++
	}
	return fired, nil
}

// disarmTimer stops and drops the armed timer for one schedule.
func (s *WakeupScheduler) disarmTimer(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.timers[id]; ok {
		t.Stop()
		delete(s.timers, id)
	}
}

// List returns every stored schedule.
func (s *WakeupScheduler) List(ctx context.Context) ([]WakeupSchedule, error) {
	return s.store.List(ctx)
}

// Close stops every armed timer.
func (s *WakeupScheduler) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, t := range s.timers {
		t.Stop()
		delete(s.timers, id)
	}
}

// armLocked arms (or re-arms) the timer for a timed schedule. The
// caller holds s.mu.
func (s *WakeupScheduler) armLocked(sched WakeupSchedule, now time.Time) {
	if t, ok := s.timers[sched.ID]; ok {
		t.Stop()
		delete(s.timers, sched.ID)
	}
	if !sched.Timed() {
		return
	}
	delay := time.Until(time.Unix(sched.NextFireUnix, 0))
	if delay < 0 {
		delay = 0
	}
	id := sched.ID
	s.timers[id] = time.AfterFunc(delay, func() {
		s.fireSchedule(id, time.Now())
	})
}

// disarmSessionLocked stops every armed timer for a session. The
// caller holds s.mu. It reports whether any timer was armed.
func (s *WakeupScheduler) disarmSessionLocked(sessionID string) bool {
	// Timers are keyed by schedule id, so consult the store for the
	// session's ids.
	schedules, err := s.store.List(context.Background())
	if err != nil {
		slog.Warn("Failed to list saved wakeups while canceling", "error", err)
		return false
	}
	disarmed := false
	for _, sched := range schedules {
		if sched.SessionID != sessionID {
			continue
		}
		if t, ok := s.timers[sched.ID]; ok {
			t.Stop()
			delete(s.timers, sched.ID)
			disarmed = true
		}
	}
	return disarmed
}

// fireSchedule claims a timed schedule (so parallel Star processes
// cannot both run it), re-arms repeats, publishes the wakeup event,
// and executes. It reports whether this process won the claim.
func (s *WakeupScheduler) fireSchedule(id string, now time.Time) (WakeupSchedule, bool) {
	s.mu.Lock()
	if t, ok := s.timers[id]; ok {
		t.Stop()
		delete(s.timers, id)
	}
	s.mu.Unlock()

	claimed, ok, err := s.store.ClaimDue(context.Background(), id, now)
	if err != nil {
		slog.Warn("Failed to claim wakeup", "schedule", id, "error", err)
		return WakeupSchedule{}, false
	}
	if !ok {
		return WakeupSchedule{}, false
	}
	if claimed.Repeating() {
		// ClaimDue advanced the stored fire; re-arm from the
		// advanced copy. One-shots stay consumed.
		s.mu.Lock()
		s.armLocked(claimed, now)
		s.mu.Unlock()
	}
	s.pubWakeup.Publish(pubsub.UpdatedEvent, pubsub.WakeupEvent{
		SessionID:  claimed.SessionID,
		ScheduleID: claimed.ID,
		Reason:     claimed.Reason,
		Prompt:     claimed.Prompt,
		Event:      claimed.Event,
		Cloud:      claimed.Cloud,
	})
	if err := s.execute(claimed); err != nil {
		if errors.Is(err, ErrWakeupBusy) && claimed.RepeatSeconds == 0 && claimed.Cron == "" {
			// A busy session must not eat a one-shot wakeup:
			// re-arm it shortly after instead of dropping it.
			retry := claimed
			retry.NextFireUnix = now.Add(time.Minute).Unix()
			if serr := s.store.Save(context.Background(), retry); serr != nil {
				slog.Warn("Failed to re-arm busy wakeup", "schedule", claimed.ID, "error", serr)
				return claimed, true
			}
			s.mu.Lock()
			s.armLocked(retry, now)
			s.mu.Unlock()
			s.pubWakeupScheduled.Publish(pubsub.CreatedEvent, pubsub.WakeupScheduledEvent{
				SessionID:  retry.SessionID,
				ScheduleID: retry.ID,
				Reason:     retry.Reason,
				Prompt:     retry.Prompt,
				Event:      retry.Event,
				Cloud:      retry.Cloud,
				Repeat:     retry.Repeat(),
				EndTime:    fireTime(retry),
			})
			return claimed, true
		}
		slog.Warn("Wakeup execution failed", "schedule", claimed.ID, "error", err)
	}
	return claimed, true
}

// execute hands a claimed schedule to the executor.
func (s *WakeupScheduler) execute(sched WakeupSchedule) error {
	s.mu.Lock()
	executor := s.executor
	s.mu.Unlock()
	if executor == nil {
		return fmt.Errorf("No wakeup executor configured")
	}
	if sched.Prompt == "" {
		return fmt.Errorf("Wakeup has no prompt")
	}
	return executor(context.Background(), sched)
}

// resolveSchedule validates a request and computes the first fire.
func resolveSchedule(req tools.ScheduleRequest, now time.Time) (WakeupSchedule, error) {
	if req.SessionID == "" {
		return WakeupSchedule{}, fmt.Errorf("A session ID is required")
	}
	if req.Prompt == "" {
		return WakeupSchedule{}, fmt.Errorf("A wakeup prompt is required")
	}
	sched := WakeupSchedule{
		ID:          NewWakeupScheduleID(),
		SessionID:   req.SessionID,
		Reason:      req.Reason,
		Prompt:      req.Prompt,
		Event:       req.Event,
		Cloud:       req.Cloud,
		CreatedUnix: now.Unix(),
	}
	timed := false
	switch {
	case req.Cron != "":
		if req.DailyAt != "" {
			return WakeupSchedule{}, fmt.Errorf("Use either cron or daily-at, not both")
		}
		if req.Delay > 0 {
			return WakeupSchedule{}, fmt.Errorf("Cron schedules ignore delay-seconds; drop it")
		}
		next, err := NextCronOccurrence(req.Cron, now)
		if err != nil {
			return WakeupSchedule{}, err
		}
		sched.Cron = req.Cron
		sched.NextFireUnix = next.Unix()
		timed = true
	case req.DailyAt != "":
		if req.Delay > 0 {
			return WakeupSchedule{}, fmt.Errorf("Daily schedules ignore delay-seconds; drop it")
		}
		cron, err := DailyAtCron(req.DailyAt)
		if err != nil {
			return WakeupSchedule{}, err
		}
		next, err := NextDailyAt(req.DailyAt, now)
		if err != nil {
			return WakeupSchedule{}, err
		}
		sched.Cron = cron
		sched.NextFireUnix = next.Unix()
		timed = true
	case req.RepeatSeconds > 0:
		if req.RepeatSeconds < 60 || req.RepeatSeconds > 30*24*3600 {
			return WakeupSchedule{}, fmt.Errorf("Repeat must be between 60 seconds and 30 days")
		}
		delay := req.Delay
		if delay <= 0 {
			delay = time.Duration(req.RepeatSeconds) * time.Second
		}
		sched.RepeatSeconds = req.RepeatSeconds
		sched.NextFireUnix = now.Add(delay).Unix()
		timed = true
	case req.Delay > 0 || req.Event == "":
		delay := req.Delay
		if delay <= 0 {
			delay = time.Minute
		}
		sched.NextFireUnix = now.Add(delay).Unix()
		timed = true
	}
	if !timed && sched.Event == "" {
		return WakeupSchedule{}, fmt.Errorf("Specify a delay, repeat, cron, daily time, or event")
	}
	return sched, nil
}

func fireTime(sched WakeupSchedule) time.Time {
	if !sched.Timed() {
		return time.Time{}
	}
	return time.Unix(sched.NextFireUnix, 0)
}

func scheduleMessage(sched WakeupSchedule) string {
	where := "locally"
	if sched.Cloud {
		where = "on the cloud machine"
	}
	what := "Wakeup scheduled " + where
	switch {
	case sched.Event != "" && sched.Timed():
		what += fmt.Sprintf("; fires %s or on event %q, whichever comes first", sched.Repeat(), sched.Event)
	case sched.Event != "":
		what += fmt.Sprintf(" on event %q", sched.Event)
	default:
		what += fmt.Sprintf("; fires %s", sched.Repeat())
	}
	if sched.Timed() {
		what += fmt.Sprintf(" (next %s)", fireTime(sched).Format(time.RFC822))
	}
	return what + "."
}
