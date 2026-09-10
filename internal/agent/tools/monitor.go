package tools

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/charmbracelet/crush/internal/shell"
)

const (
	MonitorToolName = "monitor"
)

//go:embed monitor.md
var monitorDescription string

type MonitorParams struct {
	Command     string `json:"command" description:"The background command to execute"`
	Description string `json:"description" description:"A brief description of what the monitor does"`
	WorkingDir  string `json:"working_dir,omitempty" description:"Optional working directory"`
}

type MonitorResponseMetadata struct {
	MonitorID   string `json:"monitor_id"`
	Command     string `json:"command"`
	Description string `json:"description"`
}

// NewMonitorTool creates a tool that starts a background process and streams its output
// back to the agent via wakeup events.
func NewMonitorTool(wakeups pubsub.Publisher[pubsub.WakeupEvent]) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		MonitorToolName,
		monitorDescription,
		func(ctx context.Context, params MonitorParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for monitor")
			}
			if params.Command == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("missing command")
			}

			bgManager := shell.GetBackgroundShellManager()

			// Start the process in the background.
			bgShell, err := bgManager.Start(context.Background(), params.WorkingDir, nil, params.Command, params.Description)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("failed to start monitor: %w", err)
			}

			// Start a goroutine to poll and stream new lines as wakeups.
			go streamMonitorOutput(bgShell, sessionID, wakeups)

			metadata := MonitorResponseMetadata{
				MonitorID:   bgShell.ID,
				Command:     params.Command,
				Description: params.Description,
			}

			msg := fmt.Sprintf("Started monitor %s for command `%s`.\nOutput will stream back to you automatically.\nUse `job_kill` with ID %s to stop it.", bgShell.ID, params.Command, bgShell.ID)

			return fantasy.WithResponseMetadata(fantasy.NewTextResponse(msg), metadata), nil
		},
	)
}

func streamMonitorOutput(bgShell *shell.BackgroundShell, sessionID string, wakeups pubsub.Publisher[pubsub.WakeupEvent]) {
	lastPos := 0
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stdout, stderr, done, _ := bgShell.GetOutput()
			combined := stdout
			if stderr != "" {
				combined += "\n" + stderr
			}

			// If we have new output, extract and send the new part.
			currentLen := len(combined)
			if currentLen > lastPos {
				newOutput := combined[lastPos:]
				// Only send chunks if they have a newline or if the process finished.
				if done || strings.Contains(newOutput, "\n") {
					wakeupPrompt := fmt.Sprintf("<monitor_output id=\"%s\">\n%s\n</monitor_output>", bgShell.ID, strings.TrimSpace(newOutput))
					wakeups.Publish(pubsub.UpdatedEvent, pubsub.WakeupEvent{
						SessionID: sessionID,
						Reason:    fmt.Sprintf("Monitor %s output", bgShell.ID),
						Prompt:    wakeupPrompt,
					})
					lastPos = currentLen
				}
			}

			// Stop polling if the process exited.
			if done {
				wakeups.Publish(pubsub.UpdatedEvent, pubsub.WakeupEvent{
					SessionID: sessionID,
					Reason:    fmt.Sprintf("Monitor %s exited", bgShell.ID),
					Prompt:    fmt.Sprintf("<monitor_exited id=\"%s\" />", bgShell.ID),
				})
				return
			}
		}
	}
}
