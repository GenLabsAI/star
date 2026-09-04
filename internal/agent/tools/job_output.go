package tools

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/shell"
)

const (
	JobOutputToolName = "job_output"
)

//go:embed job_output.md
var jobOutputDescription string

type JobOutputParams struct {
	ShellID string `json:"shell_id" description:"The ID of the background shell to retrieve output from"`
	Wait    bool   `json:"wait" description:"If true, block until the background shell completes, the keyword (if set) appears in the output, or the timeout elapses"`
	Timeout int    `json:"timeout,omitempty" description:"Required when wait=true. Seconds to wait before returning current output (max 600)."`
	Keyword string `json:"keyword,omitempty" description:"Optional. When wait=true, return as soon as this substring appears in stdout/stderr, instead of waiting for completion."`
}

type JobOutputResponseMetadata struct {
	ShellID          string `json:"shell_id"`
	Command          string `json:"command"`
	Description      string `json:"description"`
	Done             bool   `json:"done"`
	WorkingDirectory string `json:"working_directory"`
}

func NewJobOutputTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		JobOutputToolName,
		jobOutputDescription,
		func(ctx context.Context, params JobOutputParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.ShellID == "" {
				return fantasy.NewTextErrorResponse("missing shell_id"), nil
			}

			bgManager := shell.GetBackgroundShellManager()
			bgShell, ok := bgManager.Get(params.ShellID)
			if !ok {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("background shell not found: %s", params.ShellID)), nil
			}

			var keywordFound bool
			if params.Wait {
				if params.Timeout <= 0 {
					return fantasy.NewTextErrorResponse("timeout is required when wait=true; set a value between 1 and 600 seconds"), nil
				}
				timeoutSecs := params.Timeout
				if timeoutSecs > 600 {
					timeoutSecs = 600
				}
				waitCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
				defer cancel()
				if params.Keyword != "" {
					bgShell.WaitForKeyword(waitCtx, params.Keyword)
					keywordFound = strings.Contains(bgShell.PeekOutput(), params.Keyword)
				} else {
					bgShell.WaitContext(waitCtx)
				}
			}

			stdout, stderr, done, err := bgShell.GetOutput()

			var outputParts []string
			if stdout != "" {
				outputParts = append(outputParts, stdout)
			}
			if stderr != "" {
				outputParts = append(outputParts, stderr)
			}

			status := "running"
			if done {
				status = "completed"
				if err != nil {
					exitCode := shell.ExitCode(err)
					if exitCode != 0 {
						outputParts = append(outputParts, fmt.Sprintf("Exit code %d", exitCode))
					}
				}
			}

			output := strings.Join(outputParts, "\n")
			output = TruncateOutput(output)

			metadata := JobOutputResponseMetadata{
				ShellID:          params.ShellID,
				Command:          bgShell.Command,
				Description:      bgShell.Description,
				Done:             done,
				WorkingDirectory: bgShell.WorkingDir,
			}

			if output == "" {
				output = BashNoOutput
			}

			header := fmt.Sprintf("Status: %s", status)
			if params.Keyword != "" && params.Wait {
				if keywordFound {
					header += fmt.Sprintf(" (keyword %q found)", params.Keyword)
				} else if !done {
					header += fmt.Sprintf(" (timed out waiting for keyword %q)", params.Keyword)
				}
			}
			result := fmt.Sprintf("%s\n\n%s", header, output)
			return fantasy.WithResponseMetadata(fantasy.NewTextResponse(result), metadata), nil
		},
	)
}
