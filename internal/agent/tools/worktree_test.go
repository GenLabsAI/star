package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"charm.land/fantasy"
	"github.com/stretchr/testify/require"
)

type mockWorktreeManager struct {
	createCalls []createCall
	removeCalls []string
	err         error
}

type createCall struct {
	sessionID string
	branch    string
}

func (m *mockWorktreeManager) Create(ctx context.Context, sessionID, branch string) (string, error) {
	m.createCalls = append(m.createCalls, createCall{sessionID, branch})
	if m.err != nil {
		return "", m.err
	}
	return "/fake/" + sessionID, nil
}

func (m *mockWorktreeManager) Remove(ctx context.Context, sessionID string) error {
	m.removeCalls = append(m.removeCalls, sessionID)
	return m.err
}

func (m *mockWorktreeManager) Cleanup(ctx context.Context) error {
	return nil
}

func TestEnterWorktreeTool(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		mgr := &mockWorktreeManager{}
		tool := NewEnterWorktreeTool(mgr)
		ctx := context.WithValue(context.Background(), SessionIDContextKey, "sess-1")

		input, err := json.Marshal(EnterWorktreeParams{Branch: "feature"})
		require.NoError(t, err)

		resp, err := tool.Run(ctx, fantasy.ToolCall{
			ID:    "call-1",
			Name:  EnterWorktreeToolName,
			Input: string(input),
		})
		require.NoError(t, err)
		require.False(t, resp.IsError)
		require.Contains(t, resp.Content, "/fake/sess-1")

		require.Len(t, mgr.createCalls, 1)
		require.Equal(t, "sess-1", mgr.createCalls[0].sessionID)
		require.Equal(t, "feature", mgr.createCalls[0].branch)
	})

	t.Run("missing session id", func(t *testing.T) {
		mgr := &mockWorktreeManager{}
		tool := NewEnterWorktreeTool(mgr)

		input, _ := json.Marshal(EnterWorktreeParams{})
		_, err := tool.Run(context.Background(), fantasy.ToolCall{
			ID:    "call-1",
			Name:  EnterWorktreeToolName,
			Input: string(input),
		})
		require.ErrorContains(t, err, "session ID is required")
	})

	t.Run("manager error", func(t *testing.T) {
		mgr := &mockWorktreeManager{err: fmt.Errorf("boom")}
		tool := NewEnterWorktreeTool(mgr)
		ctx := context.WithValue(context.Background(), SessionIDContextKey, "sess-1")

		input, _ := json.Marshal(EnterWorktreeParams{})
		resp, err := tool.Run(ctx, fantasy.ToolCall{
			ID:    "call-1",
			Name:  EnterWorktreeToolName,
			Input: string(input),
		})
		require.NoError(t, err)
		require.True(t, resp.IsError)
		require.Contains(t, resp.Content, "boom")
	})
}

func TestExitWorktreeTool(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		mgr := &mockWorktreeManager{}
		tool := NewExitWorktreeTool(mgr)
		ctx := context.WithValue(context.Background(), SessionIDContextKey, "sess-1")

		input, err := json.Marshal(ExitWorktreeParams{})
		require.NoError(t, err)

		resp, err := tool.Run(ctx, fantasy.ToolCall{
			ID:    "call-1",
			Name:  ExitWorktreeToolName,
			Input: string(input),
		})
		require.NoError(t, err)
		require.False(t, resp.IsError)

		require.Len(t, mgr.removeCalls, 1)
		require.Equal(t, "sess-1", mgr.removeCalls[0])
	})
}
