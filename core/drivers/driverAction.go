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
	DriverActionKindPower
	DriverActionKindReportPowerUsage
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
// Power action
//

type DriverActionPower struct {
	State bool
}

type DriverActionPowerOutput struct {
	Changed bool
}

func (self DriverActionPower) Kind() DriverActionKind {
	return DriverActionKindPower
}

func (self DriverActionPowerOutput) Kind() DriverActionKind {
	return DriverActionKindPower
}

//
// Power usage action
//

type DriverActionGetPowerUsage struct{}

func (self DriverActionGetPowerUsage) Kind() DriverActionKind {
	return DriverActionKindReportPowerUsage
}
