package tools

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/shell"
)

const (
	JobKillToolName = "job_kill"
)

//go:embed job_kill.md
var jobKillDescription string

type JobKillParams struct {
	ShellID string `json:"shell_id" description:"The ID of the background shell to terminate"`
}

type JobKillResponseMetadata struct {
	ShellID     string `json:"shell_id"`
	Command     string `json:"command"`
	Description string `json:"description"`
}

func NewJobKillTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		JobKillToolName,
		jobKillDescription,
		func(ctx context.Context, params JobKillParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.ShellID == "" {
				return fantasy.NewTextErrorResponse("missing shell_id"), nil
			}

			bgManager := shell.GetBackgroundShellManager()

			bgShell, ok := bgManager.Get(params.ShellID)
			if !ok {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("background shell not found: %s", params.ShellID)), nil
			}

			metadata := JobKillResponseMetadata{
				ShellID:     params.ShellID,
				Command:     bgShell.Command,
				Description: bgShell.Description,
			}

			// Bound the wait on the turn's context: a job whose process tree
			// outlives the kill (a detached child holding its output pipe
			// open) would otherwise block this tool, and a tool that ignores
			// cancellation leaves the turn stuck with no way to stop it.
			err := bgManager.KillContext(ctx, params.ShellID)
			if err != nil && !errors.Is(err, shell.ErrJobStillRunning) {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			result := fmt.Sprintf("Background shell %s terminated successfully", params.ShellID)
			if err != nil {
				// The job is untracked and was signalled; only its exit is
				// unconfirmed. Reporting this as a failure would send the
				// model chasing a job that no longer exists.
				result = fmt.Sprintf("Background shell %s was signalled to stop and is no longer tracked; some of its processes have not exited yet", params.ShellID)
			}
			return fantasy.WithResponseMetadata(fantasy.NewTextResponse(result), metadata), nil
		},
	)
}
