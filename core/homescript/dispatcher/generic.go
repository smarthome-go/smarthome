package dispatcher

type CallMode uint8

const (
	// For every `callback`, a new VM is spawned.
	CallModeAllocating CallMode = iota
	// If there is already a VM running, hijack it and spawn the function in that VM instance.
	CallModeAdaptive
	// ONLY runs in existing VMs, if a suitable target does not exist, this is considered an error.
	CallModeAttaching
)

type CalledFunction struct {
	Ident    string
	CallMode CallMode
}

//
// Triggers.
//

type CallBackTrigger interface {
}

type CallBackTriggerMqtt struct {
	Topics []string
}

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
	// UnRegister(RegisterInfo) error
}
