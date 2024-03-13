package homescript

import (
	"context"
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/homescript/dispatcher"
	dispatcherTypes "github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
)

func mqttSubscribe() value.Value {
	return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
		topicsRaw := *args[0].(value.ValueList).Values

		topicsActual := make([]string, len(topicsRaw))
		for idx, v := range topicsRaw {
			topicsActual[idx] = (*v).(value.ValueString).Inner
		}

		// BUG: this is definitly unsafe
		fnName := args[1].(value.ValueString).Inner

		hmsExecutor := executor.(interpreterExecutor)

		id, err := dispatcher.Instance.Register(dispatcherTypes.RegisterInfo{
			ProgramID: hmsExecutor.programID,
			Function: &dispatcherTypes.CalledFunction{
				Ident: fnName,
				CallMode: dispatcherTypes.CallModeAttaching{
					HMSJobID: hmsExecutor.jobID,
				},
			},
			Trigger: dispatcherTypes.CallBackTriggerMqtt{
				Topics: topicsActual,
			},
		})

		if err != nil {
			return nil, value.NewVMFatalException(fmt.Sprintf("Dispatcher registration failed: %s", err.Error()), value.Vm_HostErrorKind, span)
		}

		(*hmsExecutor.registrations) = append((*hmsExecutor.registrations), id)

		return value.NewValueNull(), nil
	})
}
