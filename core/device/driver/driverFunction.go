package driver

import "github.com/smarthome-go/smarthome/core/homescript/types"

//
// Driver actions
//

type DriverActionOutput struct {
	Payload   DriverActionOutputPayload `json:"payload"`
	HmsErrors []types.HmsError          `json:"hmsErrors"`
}

type DriverActionKind uint8

const (
	DriverActionKindHealthCheck DriverActionKind = iota
	DriverActionKindReportSensorReadings
	DriverActionKindSetPower
	DriverActionKindReportPowerState
	DriverActionKindReportPowerDraw
	DriverActionKindReportDim
	DriverActionKindDim
	DriverActionKindReportColor
	DriverActionKindSetColor
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

func (action DriverActionHealthCheck) Kind() DriverActionKind {
	return DriverActionKindHealthCheck
}

type DriverActionHealthCheckOutput struct {
	Healthy bool
	Errors  []string
}

func (action DriverActionHealthCheckOutput) Kind() DriverActionKind {
	return DriverActionKindHealthCheck
}

//
// Sensor action response
//

type DriverActionReportSensorReadingsOutput struct {
	Label       string `json:"label"`
	Value       any    `json:"value"`
	HmsTypeKind string `json:"hmsType"`
	Unit        string `json:"unit"`
}

func (action DriverActionReportSensorReadingsOutput) Kind() DriverActionKind {
	return DriverActionKindReportSensorReadings
}

//
// Power state action
//

type DriverActionGetPowerState struct{}

func (action DriverActionGetPowerState) Kind() DriverActionKind {
	return DriverActionKindReportPowerState
}

type DriverActionGetPowerStateOutput struct {
	State bool `json:"state"`
}

func (action DriverActionGetPowerStateOutput) Kind() DriverActionKind {
	return DriverActionKindReportPowerState
}

//
// Power draw action
//

type DriverActionGetPowerDraw struct{}

func (action DriverActionGetPowerDraw) Kind() DriverActionKind {
	return DriverActionKindReportPowerDraw
}

type DriverActionGetPowerDrawOutput struct {
	Watts uint `json:"watts"`
}

func (action DriverActionGetPowerDrawOutput) Kind() DriverActionKind {
	return DriverActionKindReportPowerDraw
}

//
// Set power action
//

type DriverActionPower struct {
	State bool
}

func (action DriverActionPower) Kind() DriverActionKind {
	return DriverActionKindSetPower
}

type DriverActionPowerOutput struct {
	Changed bool `json:"changed"`
}

func (action DriverActionPowerOutput) Kind() DriverActionKind {
	return DriverActionKindSetPower
}

//
// Report dimmable percent
//

type DriverActionReportDim struct{}

func (action DriverActionReportDim) Kind() DriverActionKind {
	return DriverActionKindReportDim
}

type DriverActionReportRange struct {
	Lower int64 `json:"lower"`
	// Is exclusive: backend may perform `+1` or `-1` to make this fit (x..y) vs (x..=y).
	Upper int64 `json:"upper"`
}

type DriverActionReportDimOutput struct {
	Value int64                   `json:"value"`
	Label string                  `json:"label"`
	Range DriverActionReportRange `json:"range"`
}

func (action DriverActionReportDimOutput) Kind() DriverActionKind {
	return DriverActionKindReportDim
}

//
// Dim action
//

type DriverActionDim struct {
	Value int64
	Label string
}

func (action DriverActionDim) Kind() DriverActionKind {
	return DriverActionKindDim
}

type DriverActionDimOutput struct {
	Changed bool `json:"changed"`
}

func (action DriverActionDimOutput) Kind() DriverActionKind {
	return DriverActionKindDim
}

//
// Report Color
//

type DriverActionReportColor struct{}

func (action DriverActionReportColor) Kind() DriverActionKind {
	return DriverActionKindReportColor
}

type DriverActionReportColorOutput struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

func (action DriverActionReportColorOutput) Kind() DriverActionKind {
	return DriverActionKindReportColor
}

//
// Set Color
//

type DriverActionSetColor struct {
	R uint8
	G uint8
	B uint8
}

func (action DriverActionSetColor) Kind() DriverActionKind {
	return DriverActionKindSetColor
}

type DriverActionSetColorOutput struct {
}

func (action DriverActionSetColorOutput) Kind() DriverActionKind {
	return DriverActionKindSetColor
}
