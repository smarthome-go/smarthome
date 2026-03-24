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

func (action CallbackTriggerDeviceAction) Kind() CallBackTriggerKind {
	return OnDeviceActionTriggerKind
}
func (action CallbackTriggerDeviceAction) Eq(other CallBackTrigger) bool {
	if other.Kind() != OnDeviceActionTriggerKind {
		return false
	}

	otherD := other.(CallbackTriggerDeviceAction)

	if otherD.TopicWildcard != action.TopicWildcard {
		return false
	}

	if action.Topics == nil {
		return true
	}

	//  Other is a subset of self.
	for _, topic := range otherD.Topics {
		if !slices.Contains(action.Topics, topic) {
			return false
		}
	}

	//  Self is a subset of other.
	for _, topic := range action.Topics {
		if !slices.Contains(otherD.Topics, topic) {
			return false
		}
	}

	return false
}
func (action CallbackTriggerDeviceAction) Clone() CallBackTrigger {
	return CallbackTriggerDeviceAction{
		FilterKind:    action.FilterKind.Clone(),
		Topics:        slices.Clone(action.Topics),
		TopicWildcard: action.TopicWildcard,
	}
}

type DeviceFilterKind interface {
	Kind() DeviceFilterKindCode
	Clone() DeviceFilterKind
}

type DeviceFilterClass struct {
	Model  string
	Vendor string
}

func (c DeviceFilterClass) Kind() DeviceFilterKindCode {
	return DeviceFilterKindClass
}

func (c DeviceFilterClass) Clone() DeviceFilterKind {
	return DeviceFilterClass{
		Model:  c.Model,
		Vendor: c.Vendor,
	}
}

type DeviceFilterIndividual struct {
	ID string
}

func (i DeviceFilterIndividual) Kind() DeviceFilterKindCode {
	return DeviceFilterKindID
}

func (i DeviceFilterIndividual) Clone() DeviceFilterKind {
	return DeviceFilterIndividual{
		ID: i.ID,
	}
}
