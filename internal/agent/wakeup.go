package agent

import (
	"sync"
	"time"

	"github.com/charmbracelet/crush/internal/pubsub"
)

type WakeupScheduler struct {
	pubWakeup          pubsub.Publisher[pubsub.WakeupEvent]
	pubWakeupScheduled pubsub.Publisher[pubsub.WakeupScheduledEvent]
	pubWakeupCanceled  pubsub.Publisher[pubsub.WakeupCanceledEvent]
	mu                 sync.Mutex
	timers             map[string]*time.Timer
}

func NewWakeupScheduler(pubWakeup pubsub.Publisher[pubsub.WakeupEvent], pubWakeupScheduled pubsub.Publisher[pubsub.WakeupScheduledEvent], pubWakeupCanceled pubsub.Publisher[pubsub.WakeupCanceledEvent]) *WakeupScheduler {
	return &WakeupScheduler{
		pubWakeup:          pubWakeup,
		pubWakeupScheduled: pubWakeupScheduled,
		pubWakeupCanceled:  pubWakeupCanceled,
		timers:             make(map[string]*time.Timer),
	}
}

func (s *WakeupScheduler) Schedule(sessionID string, delay time.Duration, reason, prompt string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t, ok := s.timers[sessionID]; ok {
		t.Stop()
	}

	s.pubWakeupScheduled.Publish(pubsub.CreatedEvent, pubsub.WakeupScheduledEvent{
		SessionID: sessionID,
		Reason:    reason,
		Prompt:    prompt,
		EndTime:   time.Now().Add(delay),
	})

	s.timers[sessionID] = time.AfterFunc(delay, func() {
		s.pubWakeup.Publish(pubsub.UpdatedEvent, pubsub.WakeupEvent{
			SessionID: sessionID,
			Reason:    reason,
			Prompt:    prompt,
		})

		s.mu.Lock()
		delete(s.timers, sessionID)
		s.mu.Unlock()
	})
}

func (s *WakeupScheduler) Cancel(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t, ok := s.timers[sessionID]; ok {
		t.Stop()
		delete(s.timers, sessionID)
		s.pubWakeupCanceled.Publish(pubsub.DeletedEvent, pubsub.WakeupCanceledEvent{
			SessionID: sessionID,
		})
	}
}
