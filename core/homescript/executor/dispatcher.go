package executor

import (
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/homescript/dispatcher"
	"github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
	hmsTypes "github.com/smarthome-go/smarthome/core/homescript/types"
)

func (self InterpreterExecutor) RegisterTrigger(
	callbackFunctionIdentMangled string,
	eventTriggerIdent string,
	span errors.Span,
	args []value.Value,
) error {
	var registrationID types.RegistrationID
	var err error
	switch eventTriggerIdent {
	case hmsTypes.TriggerMqttMessageIdent:
		registrationID, err = registerTriggerMessage(
			callbackFunctionIdentMangled,
			nil,
			self.ProgramID,
			args,
			self.context,
		)
	case hmsTypes.TriggerMinuteIdent:
		registrationID, err = registerTriggerMinute(
			callbackFunctionIdentMangled,
			self.ProgramID,
			self.jobID,
			args,
			self.context,
		)
	case hmsTypes.TriggerKillIdent:
		self.registerTriggerKill(callbackFunctionIdentMangled)
	case hmsTypes.TriggerDeviceEvent:
		registrationID, err = registerTriggerSingleDevice(
			callbackFunctionIdentMangled,
			self.ProgramID,
			self.jobID,
			args,
			self.context,
		)
	case hmsTypes.TriggerDeviceClassEvent:
		panic("HALLO")
		// registerTriggerDevice()
	default:
		panic("Encountered unimplemented trigger function")
	}

	if err != nil {
		return err
	}

	*self.registrations = append(*self.registrations, registrationID)

	return nil
}

type DeviceFilter struct {
	// TODO: add actual code here
}

func registerTriggerSingleDevice(
	callbackFunctionIdentMangled string,
	programID string,
	// This is required as this trigger does not make sense in annotations.
	// Therefore, this trigger only works with the attaching calling mode.
	jobID uint64,
	args []value.Value,
	context hmsTypes.ExecutionContext,
) (types.RegistrationID, error) {
	deviceID := args[0].(value.ValueString).Inner

	topicsStrList := make([]string, 0)
	topicList := args[1].(value.ValueOption).Inner
	topicWildcard := true

	if topicList != nil {
		topicWildcard = false

		for _, item := range *(*topicList).(value.ValueList).Values {
			topicsStrList = append(topicsStrList, (*item).(value.ValueString).Inner)
		}
	}

	callMode := types.CallMode(types.CallModeAdaptive{
		AllocatingFallback: types.CallModeAllocating{
			Context: context,
		},
	})

	id, err := dispatcher.Instance.Register(
		types.RegisterInfo{
			ProgramID: programID,
			Function: &types.CalledFunction{
				Ident:          callbackFunctionIdentMangled,
				IdentIsLiteral: true,
				CallMode:       callMode,
			},
			Trigger: types.CallbackTriggerDeviceAction{
				FilterKind: types.DeviceFilterKind(types.DeviceFilterIndividual{
					ID: deviceID,
				}),
				Topics:        topicsStrList,
				TopicWildcard: topicWildcard,
			},
		},
		types.ToleranceRetry,
	)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func registerTriggerDeviceClass(
	callbackFunctionIdentMangled string,
	programID string,
	// This is required as this trigger does not make sense in annotations.
	// Therefore, this trigger only works with the attaching calling mode.
	jobID uint64,
	args []value.Value,
	context hmsTypes.ExecutionContext,
) (types.RegistrationID, error) {
	vendor := args[0].(value.ValueString).Inner
	model := args[1].(value.ValueString).Inner

	topicsStrList := make([]string, 0)
	topicList := args[2].(value.ValueOption).Inner
	topicWildCard := true

	if topicList != nil {
		topicWildCard = false

		for _, item := range *(*topicList).(value.ValueList).Values {
			topicsStrList = append(topicsStrList, (*item).(value.ValueString).Inner)
		}
	}

	callMode := types.CallMode(types.CallModeAdaptive{
		AllocatingFallback: types.CallModeAllocating{
			Context: context,
		},
	})

	id, err := dispatcher.Instance.Register(
		types.RegisterInfo{
			ProgramID: programID,
			Function: &types.CalledFunction{
				Ident:          callbackFunctionIdentMangled,
				IdentIsLiteral: true,
				CallMode:       callMode,
			},
			Trigger: types.CallbackTriggerDeviceAction{
				FilterKind: types.DeviceFilterKind(types.DeviceFilterClass{
					Model:  model,
					Vendor: vendor,
				}),
				Topics:        topicsStrList,
				TopicWildcard: topicWildCard,
			},
		},
		types.ToleranceRetry,
	)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (self *InterpreterExecutor) registerTriggerKill(callbackFunctionMangled string) {
	*self.OnKillCallbackFuncs = append(*self.OnKillCallbackFuncs, callbackFunctionMangled)
	spew.Dump(self.OnKillCallbackFuncs)
}

func registerTriggerMessage(
	callbackFunctionIdentMangled string,
	callmodeOverride *types.CallMode,
	programID string,
	args []value.Value,
	context hmsTypes.ExecutionContext,
) (types.RegistrationID, error) {
	topicsStrList := make([]string, 0)

	topicList := args[0].(value.ValueList).Values
	for _, item := range *topicList {
		topicsStrList = append(topicsStrList, (*item).(value.ValueString).Inner)
	}

	callMode := types.CallMode(types.CallModeAdaptive{
		AllocatingFallback: types.CallModeAllocating{
			Context: context,
		},
	})

	if callmodeOverride != nil {
		callMode = *callmodeOverride
	}

	id, err := dispatcher.Instance.Register(
		types.RegisterInfo{
			ProgramID: programID,
			Function: &types.CalledFunction{
				Ident:          callbackFunctionIdentMangled,
				IdentIsLiteral: true,
				CallMode:       callMode,
				// TODO: this is broken?
			},
			Trigger: nil,
		},
		// TODO: maybe make this a `toleranceFunc` to only retry on specific failures
		types.ToleranceRetry,
	)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func registerTriggerMinute(
	callbackFunctionIdentMangled string,
	programID string,

	// This is required as this trigger does not make sense in annotations.
	// Therefore, this trigger only works with the attaching calling mode.
	jobID uint64,

	args []value.Value,
	context hmsTypes.ExecutionContext,
) (types.RegistrationID, error) {
	stringArgs := make([]string, len(args))
	for idx, arg := range args {
		argVString, err := arg.Display()
		if err != nil {
			panic((*err).Message())
		}
		stringArgs[idx] = argVString
	}

	// logger.Tracef(
	// 	"Registered trigger `minute` with callback fn `%s` and args `[%s]`",
	// 	callbackFunctionIdentMangled,
	// 	strings.Join(stringArgs, ", "),
	// )

	minutes := args[0].(value.ValueInt).Inner
	now := time.Now()
	then := now.Add(time.Minute * time.Duration(minutes))

	callmode := types.CallMode(types.CallModeAttaching{
		HMSJobID: jobID,
	})

	id, err := dispatcher.Instance.Register(
		types.RegisterInfo{
			ProgramID: programID,
			Function: &types.CalledFunction{
				Ident:          callbackFunctionIdentMangled,
				IdentIsLiteral: true,
				CallMode:       callmode,
			},
			Trigger: types.CallBackTriggerAtTime{
				Hour:         uint8(then.Hour()),
				Minute:       uint8(then.Minute()),
				Second:       uint8(then.Second()),
				Mode:         types.OnlyOnceTriggerTimeMode,
				RegisteredAt: time.Now(),
			},
		},
		// NOTE: annotations should never fail silently.
		types.ToleranceRetry,
	)

	if err != nil {
		return 0, err
	}

	return id, nil
}
