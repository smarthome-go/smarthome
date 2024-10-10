package types

import "sync"

type RegistrationID uint64

type Registrations struct {
	Lock                   sync.RWMutex
	Set                    map[RegistrationID]RegisterInfo
	MqttRegistrations      map[string][]RegistrationID
	SchedulerRegistrations map[string]RegistrationID
	Device                 []DeviceRegistration
}

func (self *Registrations) Copy() map[RegistrationID]RegisterInfo {
	clone := make(map[RegistrationID]RegisterInfo)

	self.Lock.RLock()
	defer self.Lock.RUnlock()

	for k, v := range self.Set {
		clone[k] = v.Clone()
	}

	return clone
}

type DeviceRegistration struct {
	ID     RegistrationID
	Action CallbackTriggerDeviceAction
}
