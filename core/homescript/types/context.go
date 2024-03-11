package types

import "time"

type AutomationContext struct {
	NotificationContext *NotificationContext
	MaximumHMSRuntime   *time.Duration
}

type NotificationContext struct {
	Id          uint
	Title       string
	Description string
	Level       uint8
}
