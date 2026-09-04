package tools

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"charm.land/fantasy"
	"github.com/stretchr/testify/require"
)

type mockScheduler struct {
	scheduleCalls []scheduleCall
	cancelCalls   []string
}

type scheduleCall struct {
	sessionID string
	delay     time.Duration
	reason    string
	prompt    string
}

func (m *mockScheduler) Schedule(sessionID string, delay time.Duration, reason, prompt string) {
	m.scheduleCalls = append(m.scheduleCalls, scheduleCall{sessionID, delay, reason, prompt})
}

func (m *mockScheduler) Cancel(sessionID string) {
	m.cancelCalls = append(m.cancelCalls, sessionID)
}

func runScheduleWakeupTool(t *testing.T, tool fantasy.AgentTool, ctx context.Context, params ScheduleWakeupParams) fantasy.ToolResponse {
	t.Helper()
	input, err := json.Marshal(params)
	require.NoError(t, err)

	resp, err := tool.Run(ctx, fantasy.ToolCall{
		ID:    "test-call",
		Name:  ScheduleWakeupToolName,
		Input: string(input),
	})
	require.NoError(t, err)
	return resp
}

func TestScheduleWakeupTool(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock := &mockScheduler{}
		tool := NewScheduleWakeupTool(mock)
		ctx := context.WithValue(context.Background(), SessionIDContextKey, "sess-1")

		resp := runScheduleWakeupTool(t, tool, ctx, ScheduleWakeupParams{
			DelaySeconds: 120,
			Reason:       "waiting for CI",
			Prompt:       "check ci",
		})

		require.False(t, resp.IsError)
		require.Len(t, mock.scheduleCalls, 1)
		require.Equal(t, "sess-1", mock.scheduleCalls[0].sessionID)
		require.Equal(t, 120*time.Second, mock.scheduleCalls[0].delay)
		require.Equal(t, "waiting for CI", mock.scheduleCalls[0].reason)
		require.Equal(t, "check ci", mock.scheduleCalls[0].prompt)
	})

	t.Run("clamp delay too small", func(t *testing.T) {
		t.Parallel()
		mock := &mockScheduler{}
		tool := NewScheduleWakeupTool(mock)
		ctx := context.WithValue(context.Background(), SessionIDContextKey, "sess-1")

		runScheduleWakeupTool(t, tool, ctx, ScheduleWakeupParams{DelaySeconds: 30})

		require.Len(t, mock.scheduleCalls, 1)
		require.Equal(t, 60*time.Second, mock.scheduleCalls[0].delay)
	})

	t.Run("clamp delay too large", func(t *testing.T) {
		t.Parallel()
		mock := &mockScheduler{}
		tool := NewScheduleWakeupTool(mock)
		ctx := context.WithValue(context.Background(), SessionIDContextKey, "sess-1")

		runScheduleWakeupTool(t, tool, ctx, ScheduleWakeupParams{DelaySeconds: 5000})

		require.Len(t, mock.scheduleCalls, 1)
		require.Equal(t, 3600*time.Second, mock.scheduleCalls[0].delay)
	})

	t.Run("cancel", func(t *testing.T) {
		t.Parallel()
		mock := &mockScheduler{}
		tool := NewScheduleWakeupTool(mock)
		ctx := context.WithValue(context.Background(), SessionIDContextKey, "sess-1")

		resp := runScheduleWakeupTool(t, tool, ctx, ScheduleWakeupParams{Stop: true})

		require.False(t, resp.IsError)
		require.Len(t, mock.cancelCalls, 1)
		require.Equal(t, "sess-1", mock.cancelCalls[0])
		require.Empty(t, mock.scheduleCalls)
	})

	t.Run("missing session", func(t *testing.T) {
		t.Parallel()
		mock := &mockScheduler{}
		tool := NewScheduleWakeupTool(mock)

		input, err := json.Marshal(ScheduleWakeupParams{DelaySeconds: 120})
		require.NoError(t, err)

		_, err = tool.Run(context.Background(), fantasy.ToolCall{
			ID:    "test-call",
			Name:  ScheduleWakeupToolName,
			Input: string(input),
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "session ID is required")
	})
}
