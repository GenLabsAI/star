package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/stretchr/testify/require"
)

type mockWakeupPublisher struct {
	events []pubsub.Event[pubsub.WakeupEvent]
	mu     sync.Mutex
}

func (m *mockWakeupPublisher) Publish(eventType pubsub.EventType, payload pubsub.WakeupEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, pubsub.Event[pubsub.WakeupEvent]{
		Type:    eventType,
		Payload: payload,
	})
}

func (m *mockWakeupPublisher) PublishMustDeliver(_ context.Context, eventType pubsub.EventType, payload pubsub.WakeupEvent) {
	m.Publish(eventType, payload)
}

type mockWakeupScheduledPublisher struct {
	events []pubsub.Event[pubsub.WakeupScheduledEvent]
	mu     sync.Mutex
}

func (m *mockWakeupScheduledPublisher) Publish(eventType pubsub.EventType, payload pubsub.WakeupScheduledEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, pubsub.Event[pubsub.WakeupScheduledEvent]{
		Type:    eventType,
		Payload: payload,
	})
}

func (m *mockWakeupScheduledPublisher) PublishMustDeliver(_ context.Context, eventType pubsub.EventType, payload pubsub.WakeupScheduledEvent) {
	m.Publish(eventType, payload)
}

type mockWakeupCanceledPublisher struct {
	events []pubsub.Event[pubsub.WakeupCanceledEvent]
	mu     sync.Mutex
}

func (m *mockWakeupCanceledPublisher) Publish(eventType pubsub.EventType, payload pubsub.WakeupCanceledEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, pubsub.Event[pubsub.WakeupCanceledEvent]{
		Type:    eventType,
		Payload: payload,
	})
}

func (m *mockWakeupCanceledPublisher) PublishMustDeliver(_ context.Context, eventType pubsub.EventType, payload pubsub.WakeupCanceledEvent) {
	m.Publish(eventType, payload)
}

func TestWakeupScheduler(t *testing.T) {
	t.Parallel()

	pubWakeup := &mockWakeupPublisher{}
	pubWakeupScheduled := &mockWakeupScheduledPublisher{}
	pubWakeupCanceled := &mockWakeupCanceledPublisher{}

	s := NewWakeupScheduler(pubWakeup, pubWakeupScheduled, pubWakeupCanceled)

	// Test Schedule
	delay := 50 * time.Millisecond
	s.Schedule("sess1", delay, "reason1", "prompt1")

	// Verify WakeupScheduledEvent was published
	require.Len(t, pubWakeupScheduled.events, 1)
	require.Equal(t, pubsub.CreatedEvent, pubWakeupScheduled.events[0].Type)
	require.Equal(t, "sess1", pubWakeupScheduled.events[0].Payload.SessionID)
	require.Equal(t, "reason1", pubWakeupScheduled.events[0].Payload.Reason)
	require.Equal(t, "prompt1", pubWakeupScheduled.events[0].Payload.Prompt)

	// Wait for timer to fire
	time.Sleep(delay * 2)

	// Verify WakeupEvent was published
	pubWakeup.mu.Lock()
	require.Len(t, pubWakeup.events, 1)
	require.Equal(t, pubsub.UpdatedEvent, pubWakeup.events[0].Type)
	require.Equal(t, "sess1", pubWakeup.events[0].Payload.SessionID)
	require.Equal(t, "reason1", pubWakeup.events[0].Payload.Reason)
	require.Equal(t, "prompt1", pubWakeup.events[0].Payload.Prompt)
	pubWakeup.mu.Unlock()

	// Test Cancel
	s.Schedule("sess2", delay*10, "reason2", "prompt2")
	s.Cancel("sess2")

	// Verify WakeupCanceledEvent was published
	require.Len(t, pubWakeupCanceled.events, 1)
	require.Equal(t, pubsub.DeletedEvent, pubWakeupCanceled.events[0].Type)
	require.Equal(t, "sess2", pubWakeupCanceled.events[0].Payload.SessionID)

	// Wait to ensure timer doesn't fire
	time.Sleep(delay * 2)
	pubWakeup.mu.Lock()
	require.Len(t, pubWakeup.events, 1) // still just the first one
	pubWakeup.mu.Unlock()
}
