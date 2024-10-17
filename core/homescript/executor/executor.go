package executor

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/homescript/dispatcher"
	dispatcherT "github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

type InterpreterExecutor struct {
	// All `attaching` registrations in the dispatcher (these need to be revoked before the VM is deleted).
	registrations *[]dispatcherT.RegistrationID
	// Other.
	jobID     uint64
	ProgramID string

	// username          string
	ioWriter io.Writer
	// args              map[string]string
	// automationContext *types.AutomationContext
	// cancelCtxFunc     context.CancelFunc

	singletons map[string]value.Value

	context types.ExecutionContext

	cancelation types.Cancelation

	// Mangled names of the functions that are to be called when this program is killed.
	OnKillCallbackFuncs *[]string

	// Stdin buffer.
	stdin *types.StdinBuffer

	manager types.Manager
}

func (self InterpreterExecutor) Free() error {
	var errRes error

	for _, registration := range *self.registrations {
		// Return the first error that is found
		if err := dispatcher.Instance.Unregister(registration); err != nil && errRes == nil {
			errRes = err
		}
	}

	return errRes
}

func NewInterpreterExecutor(
	jobID uint64,
	programID string,
	// username string,
	writer io.Writer,
	// args map[string]string,
	// automationContext *types.AutomationContext,
	// cancelCtxFunc context.CancelFunc,
	cancelation types.Cancelation,
	singletons map[string]value.Value,
	context types.ExecutionContext,
	stdin *types.StdinBuffer,
	mananger types.Manager,
) InterpreterExecutor {
	registrations := make([]dispatcherT.RegistrationID, 0)
	onKillCallbackFuncs := make([]string, 0)

	if stdin == nil {
		stdin = types.NewStdinBuffer()
	}

	return InterpreterExecutor{
		registrations:       &registrations,
		jobID:               jobID,
		ProgramID:           programID,
		ioWriter:            writer,
		singletons:          singletons,
		context:             context,
		cancelation:         cancelation,
		OnKillCallbackFuncs: &onKillCallbackFuncs,
		stdin:               stdin,
		manager:             mananger,
	}
}

func parseDate(year, month, day int) (time.Time, bool) {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	y, m, d := t.Date()
	return t, y == year && int(m) == month && d == day
}

func (self InterpreterExecutor) LoadSingleton(singletonIdent, moduleName string) (val value.Value, valid bool, err error) {
	// logger.Tracef("Loading singleton `%s` from module `%s`", singletonIdent, moduleName)
	value, available := self.singletons[singletonIdent]

	if !available {
		panic(fmt.Sprintf("Singleton `%s` could not be loaded from: %v", singletonIdent, self.singletons))
	}

	// disp, e := value.Display()
	// if e != nil {
	// 	panic(e)
	// }

	// logger.Tracef("Successfully loaded singleton `%s` from module `%s`: %s", singletonIdent, moduleName, disp)

	// TODO: maybe load these on demand?
	return value, available, nil
}

func (self InterpreterExecutor) execHelper(
	username,
	programID string,
	arguments map[string]string,
	span errors.Span,
	manager types.Manager,
) (*value.Value, *value.VmInterrupt) {
	res, err := manager.RunUserScript(
		programID,
		username,
		nil,
		self.cancelation,
		self.ioWriter,
		nil,
		self.stdin,
		arguments,
	)

	if err != nil {
		return nil, value.NewVMThrowInterrupt(
			span,
			fmt.Sprintf("Failed to run program: `%s`", err.Error()),
		)
	}

	if !res.Errors.ContainsError {
		message := ""

		for _, err := range res.Errors.Diagnostics {
			if err.SyntaxError != nil {
				message = err.SyntaxError.Message
				break
			}
			if err.DiagnosticError != nil && err.DiagnosticError.Level == diagnostic.DiagnosticLevelError {
				message = err.DiagnosticError.Message
				break
			}
			if err.RuntimeInterrupt != nil {
				message = err.RuntimeInterrupt.Message
				break
			}
		}

		return nil, value.NewVMThrowInterrupt(
			span,
			fmt.Sprintf("Invoked program crashed: `%s`", message),
		)
	}

	return value.NewValueNull(), nil
}

func (self InterpreterExecutor) execBuiltin(usernameNeedsToBeSpecified bool, manager types.Manager) value.Value {
	return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
		var username *string
		// This will be 1 if the username is being read as the first argument.
		argumentIndexOffset := 0

		switch usernameNeedsToBeSpecified {
		case true:
			if self.context.Kind() == types.HMS_PROGRAM_KIND_USER {
				return nil, value.NewVMFatalException(
					fmt.Sprintf("The usage of the `%s` function in a user environment is not allowed", execUserFnIdent),
					value.Vm_HostErrorKind,
					span,
				)
			}

			usernameStr := args[0].(value.ValueString).Inner
			username = &usernameStr
			argumentIndexOffset = 1
		case false:
			if self.context.Username() == nil {
				return nil, value.NewVMFatalException(
					fmt.Sprintf("The usage of the `%s` function in a non-user environment is not possible", execFnIdent),
					value.Vm_HostErrorKind,
					span,
				)
			}
		}

		if username == nil {
			panic("Encountered internal bug: `username` cannot be <nil> at this point")
		}

		programID := args[argumentIndexOffset+0].(value.ValueString).Inner
		argOpt := args[argumentIndexOffset+1].(value.ValueOption)

		arguments := make(map[string]string)
		if argOpt.IsSome() {
			argFields := (*argOpt.Inner).(value.ValueAnyObject).FieldsInternal
			for key, value := range argFields {
				disp, i := (*value).Display()
				if i != nil {
					return nil, i
				}
				arguments[key] = disp
			}
		}

		// TODO: remove this once argument support is implemented.
		if len(arguments) != 0 {
			return nil, value.NewVMFatalException(
				"BUG: Argument support is not implemented yet",
				value.Vm_HostErrorKind,
				span,
			)
		}

		return self.execHelper(
			*username,
			programID,
			arguments,
			span,
			manager,
		)
	})
}
