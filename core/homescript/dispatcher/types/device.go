package types

import "slices"

//
// Device Triggers.
//

type DeviceFilterKindCode uint8

const (
	DeviceFilterKindClass DeviceFilterKindCode = iota
	DeviceFilterKindID
)

type CallbackTriggerDeviceAction struct {
	FilterKind    DeviceFilterKind
	Topics        []string
	TopicWildcard bool
}

func (self CallbackTriggerDeviceAction) Kind() CallBackTriggerKind { return OnDeviceActionTriggerKind }
func (self CallbackTriggerDeviceAction) Eq(other CallBackTrigger) bool {
	if other.Kind() != OnDeviceActionTriggerKind {
		return false
	}

	otherD := other.(CallbackTriggerDeviceAction)

	if otherD.TopicWildcard != self.TopicWildcard {
		return false
	}

	if self.Topics == nil {
		return true
	}

	//  Other is a subset of self.
	for _, t := range otherD.Topics {
		if !slices.Contains(self.Topics, t) {
			return false
		}
	}

	//  Self is a subset of other.
	for _, t := range self.Topics {
		if !slices.Contains(otherD.Topics, t) {
			return false
		}
	}

	return false
}

type DeviceFilterKind interface {
	Kind() DeviceFilterKindCode
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
