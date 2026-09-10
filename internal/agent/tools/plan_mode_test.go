package tools

import (
	"context"
	"encoding/json"
	"testing"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/permission"
	"github.com/stretchr/testify/require"
)

func runPlanModeTool(t *testing.T, tool fantasy.AgentTool, ctx context.Context) (fantasy.ToolResponse, error) {
	t.Helper()
	input, err := json.Marshal(PlanModeParams{})
	require.NoError(t, err)
	return tool.Run(ctx, fantasy.ToolCall{ID: "test-call", Name: tool.Info().Name, Input: string(input)})
}

func TestPlanModeTools(t *testing.T) {
	t.Parallel()

	permissions := permission.NewPermissionService(t.TempDir(), false, nil)
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "sess-1")

	response, err := runPlanModeTool(t, NewEnterPlanModeTool(permissions), ctx)
	require.NoError(t, err)
	require.False(t, response.IsError)
	require.Equal(t, permission.ModePlan, permissions.SessionMode("sess-1"))

	response, err = runPlanModeTool(t, NewExitPlanModeTool(permissions), ctx)
	require.NoError(t, err)
	require.False(t, response.IsError)
	require.Equal(t, permission.ModeNormal, permissions.SessionMode("sess-1"))
}

func TestPlanModeToolsRequireSession(t *testing.T) {
	t.Parallel()

	permissions := permission.NewPermissionService(t.TempDir(), false, nil)
	_, err := runPlanModeTool(t, NewEnterPlanModeTool(permissions), context.Background())
	require.ErrorContains(t, err, "session ID is required")
}
