package types

import (
	"fmt"
	"time"
)

type CallModeKind uint8

const (
	// For every `callback`, a new VM is spawned.
	CallModeKindAllocating CallModeKind = iota
	// If there is already a VM running, hijack it and spawn the function in that VM instance.
	CallModeKindAdaptive
	// ONLY runs in existing VMs, if a suitable target does not exist, this is considered an error.
	CallModeKindAttaching
)

type CallMode interface {
	Kind() CallModeKind
}

//
// Allocating.
//

type CallModeAllocating struct{}

func (c CallModeAllocating) Kind() CallModeKind { return CallModeKindAllocating }

//
// Adaptive.
//

type CallModeAdaptive struct{}

func (c CallModeAdaptive) Kind() CallModeKind { return CallModeKindAdaptive }

//
// Attaching.
//

type CallModeAttaching struct {
	// This is required so that the dispatcher can dispatch the call to the correct VM instance.
	HMSJobID uint64
}

func (c CallModeAttaching) Kind() CallModeKind { return CallModeKindAttaching }

type CalledFunction struct {
	Ident          string
	IdentIsLiteral bool
	CallMode       CallMode
}

//
// Triggers.
//

type CallBackTriggerKind uint8

const (
	OnMqttCallBackTriggerKind CallBackTriggerKind = iota
	AtTimeCallBackTriggerKind
)

type CallBackTrigger interface {
	Kind() CallBackTriggerKind
}

// MQTT Trigger.

type CallBackTriggerMqtt struct {
	Topics []string
}

func (self CallBackTriggerMqtt) Kind() CallBackTriggerKind { return OnMqttCallBackTriggerKind }

// AtTime Trigger.

type TriggerTimeMode uint8

const (
	OnlyOnceTriggerTimeMode TriggerTimeMode = iota
	RepeatingTriggerTimeMode
)

type CallBackTriggerAtTime struct {
	Hour         uint8
	Minute       uint8
	Second       uint8
	Mode         TriggerTimeMode
	RegisteredAt time.Time
}

func (self CallBackTriggerAtTime) Kind() CallBackTriggerKind { return AtTimeCallBackTriggerKind }

//
// Dispatcher.
//

type RegisterInfo struct {
	ProgramID string
	Function  *CalledFunction
	Trigger   CallBackTrigger
}

type Dispatcher interface {
	Register(RegisterInfo) error
	UnRegister(RegisterInfo) error
}

const mQTTProtocol = "tcp"

func MakeBrokerURI(host string, port uint16) string {
	return fmt.Sprintf("%s://%s:%d", mQTTProtocol, host, port)
}
