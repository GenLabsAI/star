package pubsub

import (
	"time"
)

type WakeupEvent struct {
	SessionID  string
	ScheduleID string
	Reason     string
	Prompt     string
	Event      string
	Cloud      bool
}

type WakeupScheduledEvent struct {
	SessionID  string
	ScheduleID string
	Reason     string
	Prompt     string
	Event      string
	Cloud      bool
	Repeat     string
	EndTime    time.Time
}

type WakeupCanceledEvent struct {
	SessionID string
}
