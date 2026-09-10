package tools

import (
	"context"
	_ "embed"
	"fmt"

	"charm.land/fantasy"

	"github.com/charmbracelet/crush/internal/worktree"
)

//go:embed enter_worktree.md
var enterWorktreeDescription string

const EnterWorktreeToolName = "enter_worktree"

// EnterWorktreeParams holds the parameters for the enter_worktree tool.
type EnterWorktreeParams struct {
	Branch string `json:"branch,omitempty" description:"Optional new branch name for the worktree"`
}

// NewEnterWorktreeTool creates the enter_worktree tool.
func NewEnterWorktreeTool(manager worktree.Manager) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		EnterWorktreeToolName,
		enterWorktreeDescription,
		func(ctx context.Context, params EnterWorktreeParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for enter_worktree")
			}
			path, err := manager.Create(ctx, sessionID, params.Branch)
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			return fantasy.NewTextResponse(fmt.Sprintf("Entered worktree: %s", path)), nil
		},
	)
}
