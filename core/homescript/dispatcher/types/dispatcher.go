package types

import "sync"

type RegistrationID uint64

type Registrations struct {
	Lock              sync.RWMutex
	Set               map[RegistrationID]RegisterInfo
	MqttRegistrations map[string][]RegistrationID
}
