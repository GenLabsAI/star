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
	DelaySeconds  int    `json:"delay_seconds,omitempty" description:"Seconds to wait before waking up (60-3600 for one-shot wakeups)"`
	RepeatSeconds int    `json:"repeat_seconds,omitempty" description:"Repeat the wakeup every N seconds (60-2592000); first fire defaults to one interval out unless delay_seconds is set"`
	Cron          string `json:"cron,omitempty" description:"Cron schedule (five fields, or @hourly/@daily/@weekly/@every <duration>); cannot combine with delay_seconds or daily_at"`
	DailyAt       string `json:"daily_at,omitempty" description:"Local wall-clock time HH:MM to wake up every day; cannot combine with delay_seconds or cron"`
	Event         string `json:"event,omitempty" description:"Named event that wakes the agent (e.g. ci-failed); with no timer fields the wakeup stays armed until the event fires"`
	Reason        string `json:"reason,omitempty" description:"A short explanation of why the agent is waiting (shown in UI)"`
	Prompt        string `json:"prompt,omitempty" description:"An optional prompt to inject when the agent wakes up"`
	Cloud         bool   `json:"cloud,omitempty" description:"Run the wakeup on the cloud machine (for cloud-mode sessions); default runs locally"`
	Stop          bool   `json:"stop,omitempty" description:"Set to true to stop an existing wakeup"`
}

// ScheduleRequest describes one durable wakeup: a timed fire
// (delay, repeat, cron, or daily time), an armed named event, or
// both. It lives in tools so the agent package (which owns the
// dispatcher) can consume it without an import cycle.
type ScheduleRequest struct {
	SessionID     string
	Delay         time.Duration
	RepeatSeconds int64
	Cron          string
	DailyAt       string
	Event         string
	Reason        string
	Prompt        string
	Cloud         bool
}

type WakeupScheduler interface {
	Schedule(sessionID string, delay time.Duration, reason, prompt string)
	Cancel(sessionID string)
}

// ExtendedWakeupScheduler is implemented by the durable dispatcher;
// the tool uses it when the agent passes scheduling options beyond
// a plain one-shot delay.
type ExtendedWakeupScheduler interface {
	WakeupScheduler
	ScheduleAdvanced(ctx context.Context, req ScheduleRequest) (ScheduleResult, error)
}

// ScheduleResult is the stored schedule id plus its human summary.
// The schedule itself is owned by the agent package; the tool only
// needs the message.
type ScheduleResult struct {
	ScheduleID string
	Message    string
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

			if advancedParams(params) {
				extended, ok := scheduler.(ExtendedWakeupScheduler)
				if !ok {
					return fantasy.ToolResponse{}, fmt.Errorf("This scheduler does not support repeating, cron, event, or cloud wakeups")
				}
				result, err := extended.ScheduleAdvanced(ctx, ScheduleRequest{
					SessionID:     sessionID,
					Delay:         time.Duration(params.DelaySeconds) * time.Second,
					RepeatSeconds: int64(params.RepeatSeconds),
					Cron:          params.Cron,
					DailyAt:       params.DailyAt,
					Event:         params.Event,
					Reason:        params.Reason,
					Prompt:        params.Prompt,
					Cloud:         params.Cloud,
				})
				if err != nil {
					return fantasy.ToolResponse{}, err
				}
				return fantasy.NewTextResponse(result.Message), nil
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

// advancedParams reports whether the agent asked for anything beyond
// a plain one-shot delay. Plain delays keep the legacy clamp
// behavior; anything else goes through the durable scheduler.
func advancedParams(params ScheduleWakeupParams) bool {
	return params.RepeatSeconds != 0 || params.Cron != "" || params.DailyAt != "" ||
		params.Event != "" || params.Cloud
}
