package tools

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"charm.land/fantasy"
)

//go:embed schedule_wakeup.md
var scheduleWakeupDescription string

const ScheduleWakeupToolName = "schedule_wakeup"

type ScheduleWakeupParams struct {
	DelaySeconds int    `json:"delay_seconds,omitempty" description:"Seconds to wait before waking up (60-3600)"`
	Reason       string `json:"reason,omitempty" description:"A short explanation of why the agent is waiting (shown in UI)"`
	Prompt       string `json:"prompt,omitempty" description:"An optional prompt to inject when the agent wakes up"`
	Stop         bool   `json:"stop,omitempty" description:"Set to true to stop an existing wakeup"`
}

type WakeupScheduler interface {
	Schedule(sessionID string, delay time.Duration, reason, prompt string)
	Cancel(sessionID string)
}

func NewScheduleWakeupTool(scheduler WakeupScheduler) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ScheduleWakeupToolName,
		scheduleWakeupDescription,
		func(ctx context.Context, params ScheduleWakeupParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for schedule_wakeup")
			}

			if params.Stop {
				scheduler.Cancel(sessionID)
				return fantasy.NewTextResponse("Wakeup cancelled."), nil
			}

			// Clamp to [60, 3600]
			if params.DelaySeconds < 60 {
				params.DelaySeconds = 60
			} else if params.DelaySeconds > 3600 {
				params.DelaySeconds = 3600
			}

			scheduler.Schedule(sessionID, time.Duration(params.DelaySeconds)*time.Second, params.Reason, params.Prompt)
			return fantasy.NewTextResponse(fmt.Sprintf("Wakeup scheduled for %d seconds from now.", params.DelaySeconds)), nil
		},
	)
}
