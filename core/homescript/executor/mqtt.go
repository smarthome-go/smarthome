package executor

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

		// BUG: this is definitely unsafe
		fn := args[1].(value.ValueVMFunction).Ident

		hmsExecutor := executor.(InterpreterExecutor)

		id, err := dispatcher.Instance.Register(
			dispatcherTypes.RegisterInfo{
				ProgramID: hmsExecutor.ProgramID,
				Function: &dispatcherTypes.CalledFunction{
					Ident:          fn,
					IdentIsLiteral: true,
					CallMode: dispatcherTypes.CallModeAttaching{
						HMSJobID: hmsExecutor.jobID,
					},
				},
				Trigger: dispatcherTypes.CallBackTriggerMqtt{
					Topics: topicsActual,
				},
			},
			dispatcherTypes.NoTolerance,
		)

		if err != nil {
			return nil, value.NewVMFatalException(fmt.Sprintf("Dispatcher registration failed: %s", err.Error()), value.Vm_HostErrorKind, span)
		}

		(*hmsExecutor.registrations) = append((*hmsExecutor.registrations), id)

		return value.NewValueNull(), nil
	})
}

func mqttPublish() value.Value {
	return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
		topic := args[0].(value.ValueString).Inner
		payload := args[1].(value.ValueString).Inner

		if err := dispatcher.Instance.Mqtt.Publish(topic, payload); err != nil {
			return nil, value.NewVMFatalException(fmt.Sprintf("Dispatcher registration failed: %s", err.Error()), value.Vm_HostErrorKind, span)
		}

		return value.NewValueNull(), nil
	})
}
