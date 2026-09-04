package pubsub

import (
	"time"
)

type WakeupEvent struct {
	SessionID string
	Reason    string
	Prompt    string
}

type WakeupScheduledEvent struct {
	SessionID string
	Reason    string
	Prompt    string
	EndTime   time.Time
}

type WakeupCanceledEvent struct {
	SessionID string
}
