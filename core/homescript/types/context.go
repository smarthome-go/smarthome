package types

import (
	"time"
)

//
// Execution context
//

type HMS_CONTEXT_KIND uint8

const (
	HMS_PROGRAM_KIND_USER HMS_CONTEXT_KIND = iota
	HMS_PROGRAM_KIND_DEVICE_DRIVER
	HMS_PROGRAM_KIND_AUTOMATION
)

type ExecutionContext interface {
	Kind() HMS_CONTEXT_KIND
	Username() *string
	UserArgs() map[string]string
}

type ExecutionContextUser struct {
	UsernameData string
	// Arguments entered by the user. (for instance via the web UI)
	UserArguments map[string]string
}

func (u ExecutionContextUser) Kind() HMS_CONTEXT_KIND      { return HMS_PROGRAM_KIND_USER }
func (u ExecutionContextUser) Username() *string           { return &u.UsernameData }
func (u ExecutionContextUser) UserArgs() map[string]string { return u.UserArguments }

func NewExecutionContextUser(username string, userArguments map[string]string) ExecutionContextUser {
	return ExecutionContextUser{
		UsernameData:  username,
		UserArguments: userArguments,
	}
}

//
// Driver context.
//

type ExecutionContextDriver struct {
	DriverVendor string
	DriverModel  string
	// This can be `nil`, for instance if a health check is performed on the driver, without a device attached.
	DeviceID *string
}

//
// Automation context.
//

type ExecutionContextAutomation struct {
	UserContext ExecutionContextUser
	// This is != nil if the trigger of the automation was a notification.
	NotificationContext *ExecutionContextNotification

	// TODO: make this general???
	MaximumHMSRuntime *time.Duration
}

type ExecutionContextNotification struct {
	Id          uint
	Title       string
	Description string
	Level       uint8
}

func (a ExecutionContextAutomation) Kind() HMS_CONTEXT_KIND      { return HMS_PROGRAM_KIND_AUTOMATION }
func (a ExecutionContextAutomation) Username() *string           { return &a.UserContext.UsernameData }
func (a ExecutionContextAutomation) UserArgs() map[string]string { return a.UserContext.UserArguments }
