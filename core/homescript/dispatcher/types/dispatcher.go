package types

import "sync"

type RegistrationID uint64

type Registrations struct {
	Lock                   sync.RWMutex
	Set                    map[RegistrationID]RegisterInfo
	MqttRegistrations      map[string][]RegistrationID
	SchedulerRegistrations map[string]RegistrationID
	// NOTE: Kind of inefficient.
	Device []DeviceRegistration
}

type DeviceRegistration struct {
	ID     RegistrationID
	Action CallbackTriggerDeviceAction
}
