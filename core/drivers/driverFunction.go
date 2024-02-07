package drivers

import "github.com/smarthome-go/smarthome/core/homescript"

//
// Driver actions
//

type DriverActionOutput struct {
	Payload   DriverActionOutputPayload `json:"payload"`
	HmsErrors []homescript.HmsError     `json:"hmsErrors"`
}

type DriverActionKind uint8

const (
	DriverActionKindHealthCheck DriverActionKind = iota
	DriverActionKindSetPower
	DriverActionKindReportPowerState
	DriverActionKindReportPowerDraw
	DriverActionKindDim
)

type DriverAction interface {
	Kind() DriverActionKind
}

type DriverActionOutputPayload interface {
	Kind() DriverActionKind
}

//
// Healthcheck action
//

type DriverActionHealthCheck struct {
}

func (self DriverActionHealthCheck) Kind() DriverActionKind {
	return DriverActionKindHealthCheck
}

type DriverActionHealthCheckOutput struct {
	Healthy bool
	Errors  []string
}

func (self DriverActionHealthCheckOutput) Kind() DriverActionKind {
	return DriverActionKindHealthCheck
}

//
// Power state action
//

type DriverActionGetPowerState struct{}

func (self DriverActionGetPowerState) Kind() DriverActionKind {
	return DriverActionKindReportPowerState
}

type DriverActionGetPowerStateOutput struct {
	State bool `json:"state"`
}

func (self DriverActionGetPowerStateOutput) Kind() DriverActionKind {
	return DriverActionKindReportPowerState
}

//
// Power draw action
//

type DriverActionGetPowerDraw struct{}

func (self DriverActionGetPowerDraw) Kind() DriverActionKind {
	return DriverActionKindReportPowerDraw
}

type DriverActionGetPowerDrawOutput struct {
	Watts uint `json:"watts"`
}

func (self DriverActionGetPowerDrawOutput) Kind() DriverActionKind {
	return DriverActionKindReportPowerDraw
}

//
// Set power action
//

type DriverActionPower struct {
	State bool
}

func (self DriverActionPower) Kind() DriverActionKind {
	return DriverActionKindSetPower
}

type DriverActionPowerOutput struct {
	Changed bool `json:"changed"`
}

func (self DriverActionPowerOutput) Kind() DriverActionKind {
	return DriverActionKindSetPower
}
