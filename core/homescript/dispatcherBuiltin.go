package homescript

import (
	"fmt"
	"strings"
	"time"

	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/homescript/dispatcher"
	"github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
)

func (self interpreterExecutor) RegisterTrigger(
	callbackFunctionIdent string,
	eventTriggerIdent string,
	span errors.Span,
	args []value.Value,
) error {
	return self.registerTriggerOverride(callbackFunctionIdent, eventTriggerIdent, span, args, nil, nil)
}

func (self interpreterExecutor) registerTriggerOverride(
	callbackFunctionIdent string,
	eventTriggerIdent string,
	span errors.Span,
	args []value.Value,
	callmodeOverride *types.CallMode,
	// driverTiplet *driverTypes.DriverInvocationIDs,
) error {
	// TODO: also refactor this
	id, err := registerTriggerOverride(
		callbackFunctionIdent,
		eventTriggerIdent,
		span,
		args,
		callmodeOverride,
		self.username,
		self.programID,
		&self.jobID,
		driverTiplet,
	)

	if err != nil {
		return err
	}

	*self.registrations = append(*self.registrations, id)

	return nil
}

// TODO: think about a better argument structure here.
func registerTriggerOverride(
	// callbackFunctionIdent string,
	// eventTriggerIdent string,
	// span errors.Span,
	// args []value.Value,
	// callmodeOverride *types.CallMode,
	// username string,
	// programID string,
	// jobID *uint64,
	// driverTiplet *driverTypes.DriverInvocationIDs,
	foo int,
) (types.RegistrationID, error) {
	switch eventTriggerIdent {
	case "message":
		topicsStrList := make([]string, 0)

		topicList := args[0].(value.ValueList).Values
		for _, item := range *topicList {
			topicsStrList = append(topicsStrList, (*item).(value.ValueString).Inner)
		}

		callmode := types.CallMode(types.CallModeAdaptive{
			Username: username,
		})

		if callmodeOverride != nil {
			callmode = *callmodeOverride
		}

		id, err := dispatcher.Instance.Register(
			types.RegisterInfo{
				ProgramID: programID,
				Function: &types.CalledFunction{
					Ident:          callbackFunctionIdent,
					IdentIsLiteral: false,
					CallMode:       callmode,
				},
				Trigger: types.CallBackTriggerMqtt{
					Topics: topicsStrList,
				},
				DriverTriplet: driverTiplet,
			},
			// NOTE: Annotations should never fail silently.
			types.ToleranceRetry,
		)

		if err != nil {
			return 0, err
		}

		return id, nil
	case "minute":
		panic("Why does this even exist?")

		stringArgs := make([]string, len(args))
		for idx, arg := range args {
			argVString, err := arg.Display()
			if err != nil {
				panic((*err).Message())
			}
			stringArgs[idx] = argVString
		}

		logger.Tracef(
			"Registered trigger `minute` with callback fn `%s` and args `[%s]`",
			callbackFunctionIdent,
			strings.Join(stringArgs, ", "),
		)

		minutes := args[0].(value.ValueInt).Inner
		now := time.Now()
		then := now.Add(time.Minute * time.Duration(minutes))

		callmode := types.CallMode(types.CallModeAttaching{
			HMSJobID: *jobID,
		})

		if callmodeOverride != nil {
			callmode = *callmodeOverride
		}

		id, err := dispatcher.Instance.Register(
			types.RegisterInfo{
				ProgramID: programID,
				Function: &types.CalledFunction{
					Ident:          callbackFunctionIdent,
					IdentIsLiteral: false,
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
	default:
		panic(fmt.Sprintf("Unknown event trigger ident: `%s`", eventTriggerIdent))
	}
}
