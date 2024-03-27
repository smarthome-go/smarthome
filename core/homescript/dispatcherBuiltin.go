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
	switch eventTriggerIdent {
	case "message":
		topicsStrList := make([]string, 0)

		topicList := args[0].(value.ValueList).Values
		for _, item := range *topicList {
			topicsStrList = append(topicsStrList, (*item).(value.ValueString).Inner)
		}

		id, err := dispatcher.Instance.Register(
			types.RegisterInfo{
				ProgramID: self.programID,
				Function: &types.CalledFunction{
					Ident:          callbackFunctionIdent,
					IdentIsLiteral: false,
					CallMode: types.CallModeAttaching{
						HMSJobID: self.jobID,
					},
				},
				Trigger: types.CallBackTriggerMqtt{
					Topics: topicsStrList,
				},
			},
		)

		if err != nil {
			return err
		}

		*self.registrations = append(*self.registrations, id)
	case "minute":
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

		id, err := dispatcher.Instance.Register(
			types.RegisterInfo{
				ProgramID: self.programID,
				Function: &types.CalledFunction{
					Ident:          callbackFunctionIdent,
					IdentIsLiteral: false,
					CallMode: types.CallModeAttaching{
						HMSJobID: self.jobID,
					},
				},
				Trigger: types.CallBackTriggerAtTime{
					Hour:         uint8(then.Hour()),
					Minute:       uint8(then.Minute()),
					Second:       uint8(then.Second()),
					Mode:         types.OnlyOnceTriggerTimeMode,
					RegisteredAt: time.Now(),
				},
			},
		)

		if err != nil {
			return err
		}

		*self.registrations = append(*self.registrations, id)
	default:
		panic(fmt.Sprintf("Unknown event trigger ident: `%s`", eventTriggerIdent))
	}

	return nil
}
