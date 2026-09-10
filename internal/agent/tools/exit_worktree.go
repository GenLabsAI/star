package tools

import (
	"context"
	_ "embed"
	"fmt"

	"charm.land/fantasy"

	"github.com/charmbracelet/crush/internal/worktree"
)

//go:embed exit_worktree.md
var exitWorktreeDescription string

const ExitWorktreeToolName = "exit_worktree"

// ExitWorktreeParams holds the parameters for the exit_worktree tool.
type ExitWorktreeParams struct{}

// NewExitWorktreeTool creates the exit_worktree tool.
func NewExitWorktreeTool(manager worktree.Manager) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ExitWorktreeToolName,
		exitWorktreeDescription,
		func(ctx context.Context, params ExitWorktreeParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for exit_worktree")
			}
			if err := manager.Remove(ctx, sessionID); err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			return fantasy.NewTextResponse("Exited worktree and returned to original directory."), nil
		},
	)
}
