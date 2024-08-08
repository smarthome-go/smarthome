package types

import (
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
)

const TriggerMqttMessageIdent = "message"
const TriggerMinuteIdent = "minute"
const TriggerKillIdent = "kill"

// When a Homescript is executed in user mode, the user cannot use these triggers.
var ForbiddenUserTriggers = []string{TriggerMqttMessageIdent, TriggerMinuteIdent}

type TriggerAnnotation struct {
	// Callback function name (mangled)
	CalledFnIdentMangled string
	Trigger              string
	Args                 []value.Value
	Span                 errors.Span
}
