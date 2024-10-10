package types

import (
	"fmt"
	"slices"
	"time"

	hmsTypes "github.com/smarthome-go/smarthome/core/homescript/types"
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
	Clone() CallMode
}

//
// Allocating.
//

type CallModeAllocating struct {
	Context hmsTypes.ExecutionContext
}

func (c CallModeAllocating) Kind() CallModeKind { return CallModeKindAllocating }
func (c CallModeAllocating) Clone() CallMode {
	return CallModeAllocating{
		Context: c.Context.Clone(),
	}
}

//
// Adaptive.
//

type CallModeAdaptive struct {
	// The execution context is required for spawning a new HMS process.
	AllocatingFallback CallModeAllocating
}

func (c CallModeAdaptive) Kind() CallModeKind { return CallModeKindAdaptive }
func (c CallModeAdaptive) Clone() CallMode {
	return CallModeAdaptive{
		AllocatingFallback: c.AllocatingFallback.Clone().(CallModeAllocating),
	}
}

//
// Attaching.
//

type CallModeAttaching struct {
	// This is required so that the dispatcher can dispatch the call to the correct VM instance.
	HMSJobID uint64
}

func (c CallModeAttaching) Kind() CallModeKind { return CallModeKindAttaching }
func (c CallModeAttaching) Clone() CallMode {
	return CallModeAttaching{
		HMSJobID: c.HMSJobID,
	}
}

type CalledFunction struct {
	Ident          string
	IdentIsLiteral bool
	CallMode       CallMode
}

func (f CalledFunction) Clone() CalledFunction {
	return CalledFunction{
		Ident:          f.Ident,
		IdentIsLiteral: f.IdentIsLiteral,
		CallMode:       f.CallMode.Clone(),
	}
}

//
// Triggers.
//

type CallBackTriggerKind uint8

const (
	OnMqttCallBackTriggerKind CallBackTriggerKind = iota
	AtTimeCallBackTriggerKind
	OnDeviceActionTriggerKind
)

type CallBackTrigger interface {
	Kind() CallBackTriggerKind
	Eq(other CallBackTrigger) bool
	Clone() CallBackTrigger
}

// MQTT Trigger.

type CallBackTriggerMqtt struct {
	Topics []string
}

func (self CallBackTriggerMqtt) Kind() CallBackTriggerKind { return OnMqttCallBackTriggerKind }
func (self CallBackTriggerMqtt) Eq(other CallBackTrigger) bool {
	if other.Kind() != OnMqttCallBackTriggerKind {
		return false
	}

	otherM := other.(CallBackTriggerMqtt)

	// Other is subset of self.
	for _, topic := range otherM.Topics {
		if !slices.Contains(self.Topics, topic) {
			return false
		}
	}

	// Self is subset of other.
	for _, topic := range self.Topics {
		if !slices.Contains(otherM.Topics, topic) {
			return false
		}
	}

	return true
}

func (self CallBackTriggerMqtt) Clone() CallBackTrigger {
	return CallBackTriggerMqtt{
		Topics: slices.Clone(self.Topics),
	}
}

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
func (self CallBackTriggerAtTime) Eq(other CallBackTrigger) bool {
	if other.Kind() != AtTimeCallBackTriggerKind {
		return false
	}

	otherT := other.(CallBackTriggerAtTime)

	if otherT.Hour == self.Hour && otherT.Minute == self.Minute && otherT.Second == self.Second {
		return true
	}

	return false
}
func (self CallBackTriggerAtTime) Clone() CallBackTrigger {
	return CallBackTriggerAtTime{
		Hour:         self.Hour,
		Minute:       self.Minute,
		Second:       self.Second,
		Mode:         self.Mode,
		RegisteredAt: self.RegisteredAt,
	}
}

//
// Dispatcher.
//

type RegisterInfo struct {
	ProgramID string
	Function  *CalledFunction
	Trigger   CallBackTrigger
}

func (i RegisterInfo) Clone() RegisterInfo {
	var fn *CalledFunction
	if i.Function != nil {
		fnT := i.Function.Clone()
		fn = &fnT
	}
	tr := i.Trigger.Clone()

	return RegisterInfo{
		ProgramID: i.ProgramID,
		Function:  fn,
		Trigger:   tr,
	}
}

type Dispatcher interface {
	Register(RegisterInfo) error
	UnRegister(RegisterInfo) error
}

const mqttProtocol = "tcp"

func MakeBrokerURI(host string, port uint16) string {
	return fmt.Sprintf("%s://%s:%d", mqttProtocol, host, port)
}
