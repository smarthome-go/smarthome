package types

//
// Device Triggers.
//

type DeviceFilterKindCode uint8

const (
	DeviceFilterKindClass DeviceFilterKindCode = iota
	DeviceFilterKindID
)

type CallbackTriggerDeviceAction struct {
	FilterKind DeviceFilterKind
	Topics     *[]string
}

func (self CallbackTriggerDeviceAction) Kind() CallBackTriggerKind { return OnDeviceActionTriggerKind }

type DeviceFilterKind interface {
	Kind() DeviceFilterKind
}

type DeviceFilterClass struct {
	Model  string
	Vendor string
}

func (c DeviceFilterClass) Kind() DeviceFilterKindCode {
	return DeviceFilterKindClass
}

type DeviceFilterIndividual struct {
	ID string
}

func (i DeviceFilterIndividual) Kind() DeviceFilterKindCode {
	return DeviceFilterKindID
}
