package tools

import (
	"context"
	"testing"
	"time"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/stretchr/testify/require"
)

func TestMonitorToolStreamsOutput(t *testing.T) {
	t.Parallel()

	wakeups := pubsub.NewBroker[pubsub.WakeupEvent]()
	defer wakeups.Shutdown()
	events := wakeups.Subscribe(t.Context())

	tool := NewMonitorTool(wakeups)
	ctx := context.WithValue(t.Context(), SessionIDContextKey, "session-1")
	response, err := tool.Run(ctx, fantasy.ToolCall{
		ID:    "call-1",
		Name:  MonitorToolName,
		Input: `{"command":"printf 'first\\nsecond\\n'","description":"test monitor"}`,
	})
	require.NoError(t, err)
	require.Contains(t, response.Content, "Started monitor")

	var output string
	for range 2 {
		select {
		case event := <-events:
			output += event.Payload.Prompt
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for monitor output")
		}
	}
	require.Contains(t, output, "first")
	require.Contains(t, output, "second")
	require.Contains(t, output, "monitor_exited")
}

func TestMonitorToolRequiresSession(t *testing.T) {
	t.Parallel()

	wakeups := pubsub.NewBroker[pubsub.WakeupEvent]()
	defer wakeups.Shutdown()
	tool := NewMonitorTool(wakeups)

	_, err := tool.Run(t.Context(), fantasy.ToolCall{
		ID:    "call-1",
		Name:  MonitorToolName,
		Input: `{"command":"echo test"}`,
	})
	require.ErrorContains(t, err, "session ID is required")
}
