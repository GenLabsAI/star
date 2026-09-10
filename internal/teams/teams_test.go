package teams

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := &Store{dir: t.TempDir()}

	_, err := store.Join(ctx, "test-team", Agent{Name: "architect", Role: "architect", SessionID: "session-1", Status: AgentIdle})
	require.NoError(t, err)
	_, err = store.Join(ctx, "test-team", Agent{Name: "coder", Role: "coder", SessionID: "session-2", Status: AgentIdle})
	require.NoError(t, err)

	agents, err := store.ListAgents(ctx, "test-team")
	require.NoError(t, err)
	require.Len(t, agents, 2)

	task, err := store.CreateTask(ctx, "test-team", "Build the API", "Implement REST endpoints", nil)
	require.NoError(t, err)
	require.Equal(t, TaskPending, task.Status)

	claimed, err := store.ClaimTask(ctx, "test-team", task.ID, "coder")
	require.NoError(t, err)
	require.Equal(t, TaskClaimed, claimed.Status)
	require.Equal(t, "coder", claimed.Owner)

	_, err = store.ClaimTask(ctx, "test-team", task.ID, "architect")
	require.ErrorIs(t, err, ErrTaskClaimed)

	updated, err := store.UpdateTask(ctx, "test-team", task.ID, "coder", TaskRunning, "")
	require.NoError(t, err)
	require.Equal(t, TaskRunning, updated.Status)

	updated, err = store.UpdateTask(ctx, "test-team", task.ID, "coder", TaskCompleted, "Done!")
	require.NoError(t, err)
	require.Equal(t, TaskCompleted, updated.Status)
	require.Equal(t, "Done!", updated.Result)

	_, err = store.UpdateTask(ctx, "test-team", task.ID, "coder", TaskRunning, "")
	require.ErrorIs(t, err, ErrInvalidTransition)
}

func TestMessaging(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := &Store{dir: t.TempDir()}

	_, err := store.Join(ctx, "msg-team", Agent{Name: "alice", Role: "coder", SessionID: "s1", Status: AgentIdle})
	require.NoError(t, err)
	_, err = store.Join(ctx, "msg-team", Agent{Name: "bob", Role: "reviewer", SessionID: "s2", Status: AgentIdle})
	require.NoError(t, err)

	_, err = store.SendMessage(ctx, "msg-team", Message{From: "alice", To: "bob", Content: "Ready for review"})
	require.NoError(t, err)

	messages, err := store.Inbox(ctx, "msg-team", "bob", true, true)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, "Ready for review", messages[0].Content)

	messages, err = store.Inbox(ctx, "msg-team", "bob", true, false)
	require.NoError(t, err)
	require.Empty(t, messages)
}

func TestBlockedTasks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := &Store{dir: t.TempDir()}

	_, err := store.Join(ctx, "dep-team", Agent{Name: "worker", Role: "coder", SessionID: "s1", Status: AgentIdle})
	require.NoError(t, err)
	parent, err := store.CreateTask(ctx, "dep-team", "Setup DB", "", nil)
	require.NoError(t, err)
	child, err := store.CreateTask(ctx, "dep-team", "Write queries", "", []string{parent.ID})
	require.NoError(t, err)

	_, err = store.ClaimTask(ctx, "dep-team", child.ID, "worker")
	require.ErrorIs(t, err, ErrInvalidTransition)

	_, err = store.ClaimTask(ctx, "dep-team", parent.ID, "worker")
	require.NoError(t, err)
	_, err = store.UpdateTask(ctx, "dep-team", parent.ID, "worker", TaskRunning, "")
	require.NoError(t, err)
	_, err = store.UpdateTask(ctx, "dep-team", parent.ID, "worker", TaskCompleted, "done")
	require.NoError(t, err)

	_, err = store.ClaimTask(ctx, "dep-team", child.ID, "worker")
	require.NoError(t, err)
}

func TestSafeNameValidation(t *testing.T) {
	t.Parallel()
	_, err := safeName("")
	require.Error(t, err)
	_, err = safeName("../evil")
	require.Error(t, err)
	_, err = safeName("good-name_123")
	require.NoError(t, err)
}
