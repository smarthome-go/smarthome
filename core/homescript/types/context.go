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
	Clone() ExecutionContext
}

type ExecutionContextUser struct {
	Filename     string
	UsernameData string
	// Arguments entered by the user. (for instance via the web UI)
	UserArguments map[string]string
}

func (u ExecutionContextUser) Kind() HMS_CONTEXT_KIND      { return HMS_PROGRAM_KIND_USER }
func (u ExecutionContextUser) Username() *string           { return &u.UsernameData }
func (u ExecutionContextUser) UserArgs() map[string]string { return u.UserArguments }
func (u ExecutionContextUser) Clone() ExecutionContext {
	newMap := make(map[string]string)

	for k, v := range u.UserArguments {
		newMap[k] = v
	}

	return ExecutionContextUser{
		Filename:      u.Filename,
		UsernameData:  u.UsernameData,
		UserArguments: newMap,
	}
}

func NewExecutionContextUser(
	filename string,
	username string,
	userArguments map[string]string,
) ExecutionContextUser {
	return ExecutionContextUser{
		Filename:      filename,
		UsernameData:  username,
		UserArguments: userArguments,
	}
}

func NewExecutionContextUserNoFilename(
	username string,
	userArguments map[string]string,
) ExecutionContextUser {
	return ExecutionContextUser{
		Filename:      "",
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

func (d ExecutionContextDriver) Kind() HMS_CONTEXT_KIND      { return HMS_PROGRAM_KIND_DEVICE_DRIVER }
func (d ExecutionContextDriver) Username() *string           { return nil }
func (d ExecutionContextDriver) UserArgs() map[string]string { return nil }
func (d ExecutionContextDriver) Clone() ExecutionContext {
	var id *string
	if d.DeviceID != nil {
		i := *d.DeviceID
		id = &i
	}

	return ExecutionContextDriver{
		DriverVendor: d.DriverVendor,
		DriverModel:  d.DriverModel,
		DeviceID:     id,
	}
}

func NewExecutionContextDriver(vendor, model string, deviceID *string) ExecutionContextDriver {
	return ExecutionContextDriver{
		DriverVendor: vendor,
		DriverModel:  model,
		DeviceID:     deviceID,
	}
}

//
// Automation context.
//

type ExecutionContextAutomation struct {
	UserContext ExecutionContextUser
	Inner       ExecutionContextAutomationInner
}

func NewExecutionContextAutomation(
	user ExecutionContextUser,
	inner ExecutionContextAutomationInner,
) ExecutionContextAutomation {
	return ExecutionContextAutomation{
		UserContext: user,
		Inner:       inner,
	}
}

type ExecutionContextAutomationInner struct {
	// This is != nil if the trigger of the automation was a notification.
	NotificationContext *ExecutionContextNotification

	// TODO: make this general???
	MaximumHMSRuntime *time.Duration
}

func (i ExecutionContextAutomationInner) Clone() ExecutionContextAutomationInner {
	var n *ExecutionContextNotification

	if i.NotificationContext != nil {
		nT := (*i.NotificationContext).Clone()
		n = &nT
	}

	var mrt *time.Duration
	if i.MaximumHMSRuntime != nil {
		mrtT := *i.MaximumHMSRuntime
		mrt = &mrtT
	}

	return ExecutionContextAutomationInner{
		NotificationContext: n,
		MaximumHMSRuntime:   mrt,
	}
}

type ExecutionContextNotification struct {
	Id          uint
	Title       string
	Description string
	Level       uint8
}

func (n ExecutionContextNotification) Clone() ExecutionContextNotification {
	return ExecutionContextNotification{
		Id:          n.Id,
		Title:       n.Title,
		Description: n.Description,
		Level:       n.Level,
	}
}

func (a ExecutionContextAutomation) Kind() HMS_CONTEXT_KIND      { return HMS_PROGRAM_KIND_AUTOMATION }
func (a ExecutionContextAutomation) Username() *string           { return &a.UserContext.UsernameData }
func (a ExecutionContextAutomation) UserArgs() map[string]string { return a.UserContext.UserArguments }
func (a ExecutionContextAutomation) Clone() ExecutionContext {
	return ExecutionContextAutomation{
		UserContext: a.UserContext.Clone().(ExecutionContextUser),
		Inner:       a.Inner.Clone(),
	}
}
