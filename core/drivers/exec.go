package drivers

import (
	"bytes"
	"context"
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"

	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type DriverContext struct {
	DeviceId *string
}

type DriverGenericOutput struct {
	ReturnValue value.Value
}

type FunctionCall struct {
	Invocation runtime.FunctionInvocation
}

// NOTE: EXAMPLE: invocation
// &runtime.FunctionInvocation{
// 	Function: POWER_DRIVER_FUNCTION,
// 	Args: []value.Value{
// 		// Use the power state as an argument.
// 		*value.NewValueBool(action.State),
// 	},
// },

// if !hmsRes.Success {
// 	// Return first error message that is found.
// 	for _, diagnosticMsg := range hmsRes.Errors {
// 		if diagnosticMsg.DiagnosticError != nil && diagnosticMsg.DiagnosticError.Level != diagnostic.DiagnosticLevelError {
// 			continue
// 		}
//
// 		return homescript.HmsRunResultContext{}, &diagnosticMsg, nil
// 	}
//
// 	panic("Unreachable, there is at least one error if `Success` was `false`")
// }

func invokeDriverGeneric(
	// Termination handling
	cancelCtx context.Context,
	cancelFunc context.CancelFunc,
	// Call data
	driverCtx DriverContext,
	vendorId,
	modelId string,
	functionInvocation FunctionCall,
) (homescript.HmsRunResultContext, []homescript.HmsError, error) {
	driver, found, err := database.GetDeviceDriver(vendorId, modelId)
	if err != nil {
		return homescript.HmsRunResultContext{}, nil, err
	}

	if !found {
		panic(fmt.Sprintf("Driver `%s:%s` not found in the database", vendorId, modelId))
	}

	var outputBuffer bytes.Buffer

	// TODO: do not hardcode stuff like this
	filename := fmt.Sprintf("@driver:%s:%s", vendorId, modelId)

	contextSingletons := make(map[string]value.Value)

	// Load driver singleton.
	driverSingleton, found := DriverStore[DriverTuple{
		VendorID: vendorId,
		ModelID:  modelId,
	}]
	if !found {
		panic(fmt.Sprintf("Driver singleton of driver `%s:%s` not found in store", vendorId, modelId))
	}
	contextSingletons[homescript.DriverSingletonIdent] = driverSingleton

	// Load device singleton if possible.
	var deviceSingleton value.Value
	if driverCtx.DeviceId != nil {
		deviceId := *driverCtx.DeviceId
		deviceSingleton, found = DeviceStore[deviceId]
		if !found {
			panic(fmt.Sprintf("Device singleton of `%s` not found in store", deviceId))
		}
		contextSingletons[homescript.DriverDeviceSingletonIdent] = deviceSingleton
	}

	// TODO: load corresponding device singleton.
	hmsRes, resultContext, err := homescript.HmsManager.Run(
		homescript.HMS_PROGRAM_KIND_DEVICE_DRIVER,
		&homescript.AnalyzerDriverMetadata{
			VendorID: vendorId,
			ModelID:  modelId,
		},
		"", // TODO: fix username requirement.
		&filename,
		driver.HomescriptCode,
		homescript.InitiatorAPI,
		cancelCtx,
		cancelFunc,
		nil,
		make(map[string]string),
		&outputBuffer,
		nil,
		&functionInvocation.Invocation,
		contextSingletons,
	)

	if err != nil {
		return homescript.HmsRunResultContext{}, nil, err
	}

	if !hmsRes.Success {
		errors := make([]homescript.HmsError, 0)

		// Filter out any non-error messages.
		for _, d := range hmsRes.Errors {
			if d.DiagnosticError != nil && d.DiagnosticError.Level != diagnostic.DiagnosticLevelError {
				continue
			}
			errors = append(errors, d)
		}

		return homescript.HmsRunResultContext{}, errors, nil
	}

	// Get driver and device singleton.
	driverSingletonAfter, found := resultContext.Singletons[homescript.DriverSingletonIdent]
	if !found {
		panic(fmt.Sprintf("Driver singleton (`%s`) not found after driver execution", homescript.DriverSingletonIdent))
	}

	// Save driver singleton state after VM has terminated.
	driverMarshaled, _ := value.MarshalValue(driverSingletonAfter, false)
	if err := StoreDriverSingletonConfigUpdate(driver.VendorId, driver.ModelId, driverMarshaled); err != nil {
		return homescript.HmsRunResultContext{}, nil, err
	}

	// Save device singleton state after VM has terminated (if device was even loaded).
	if driverCtx.DeviceId != nil {
		deviceMarshaled, _ := value.MarshalValue(deviceSingleton, false)
		if err := StoreDeviceSingletonConfigUpdate(*driverCtx.DeviceId, deviceMarshaled); err != nil {
			return homescript.HmsRunResultContext{}, nil, err
		}
	}

	return resultContext, nil, nil
}

//
//
//
// Specialized driver action invocations.
//
//
//

type DriverInvocationIDs struct {
	deviceID string
	vendorID string
	modelID  string
}

//
// TDOO: maybe implement a function factory to create those almost identical functions more ideomatically.
//

func InvokeDriverFunc(
	ids DriverInvocationIDs,
	call FunctionCall,
) (homescript.HmsRunResultContext, []homescript.HmsError, error) {
	// TODO: add context support
	ctx, cancel := context.WithCancel(context.Background())

	runResult, hmsErrs, dbErr := invokeDriverGeneric(
		ctx,
		cancel,
		DriverContext{
			DeviceId: &ids.deviceID,
		},
		ids.vendorID,
		ids.modelID,
		call,
	)

	if dbErr != nil || hmsErrs != nil {
		return homescript.HmsRunResultContext{}, hmsErrs, dbErr
	}

	return runResult, nil, nil
}

func InvokeDriverReportPowerState(
	ids DriverInvocationIDs,
) (DriverActionGetPowerStateOutput, []homescript.HmsError, error) {
	ret, hmsErrs, err := InvokeDriverFunc(
		ids,
		FunctionCall{
			Invocation: runtime.FunctionInvocation{
				Function: homescript.DeviceFunctionReportPowerState,
				Args:     []value.Value{},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					homescript.DeviceReportPowerStateSignature(errors.Span{}).Signature,
				),
			},
		},
	)

	if err != nil || hmsErrs != nil {
		return DriverActionGetPowerStateOutput{}, hmsErrs, err
	}

	return DriverActionGetPowerStateOutput{
		State: ret.ReturnValue.(value.ValueBool).Inner,
	}, nil, nil
}

func InvokeDriverReportPowerDraw(
	ids DriverInvocationIDs,
) (DriverActionGetPowerDrawOutput, []homescript.HmsError, error) {
	ret, hmsErrs, err := InvokeDriverFunc(
		ids,
		FunctionCall{
			Invocation: runtime.FunctionInvocation{
				Function: homescript.DeviceFunctionReportPowerDraw,
				Args:     []value.Value{},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					homescript.DeviceReportPowerDrawSignature(errors.Span{}).Signature,
				),
			},
		},
	)

	if err != nil || hmsErrs != nil {
		return DriverActionGetPowerDrawOutput{}, hmsErrs, err
	}

	wattsRaw := ret.ReturnValue.(value.ValueInt).Inner
	if wattsRaw < 0 {
		return DriverActionGetPowerDrawOutput{Watts: 0},
			[]homescript.HmsError{
				{
					SyntaxError:     nil,
					DiagnosticError: nil,
					RuntimeInterrupt: &homescript.HmsRuntimeInterrupt{
						Kind: "driver",
						Message: fmt.Sprintf(
							"Device function `%s` should return positive power consumption but returned %d",
							homescript.DeviceFunctionReportPowerDraw,
							wattsRaw,
						),
					},
					Span: ret.CalledFunctionSpan,
				},
			}, nil
	}

	return DriverActionGetPowerDrawOutput{
		Watts: uint(wattsRaw),
	}, nil, nil
}

func InvokeDriverSetPower(
	deviceID,
	vendorID,
	modelID string,
	powerAction DriverActionPower,
) (DriverActionPowerOutput, []homescript.HmsError, error) {
	// TODO: add context support
	ctx, cancel := context.WithCancel(context.Background())

	runResult, hmsErrs, dbErr := invokeDriverGeneric(
		ctx,
		cancel,
		DriverContext{
			DeviceId: &deviceID,
		},
		vendorID,
		modelID,
		FunctionCall{
			Invocation: runtime.FunctionInvocation{
				Function: homescript.DeviceFuncionSetPower,
				Args: []value.Value{
					*value.NewValueBool(powerAction.State), // TODO: test this by providing an int for instance.
				},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					homescript.DeviceSetPowerSignature(errors.Span{}).Signature,
				),
			},
		},
	)

	if dbErr != nil || hmsErrs != nil {
		return DriverActionPowerOutput{}, hmsErrs, dbErr
	}

	return DriverActionPowerOutput{
		Changed: runResult.ReturnValue.(value.ValueBool).Inner,
	}, nil, nil
}

//
// Report dimmable percent
//

func InvokeDriverReportDimmable(
	ids DriverInvocationIDs,
) (DeviceDimmableInformation, []homescript.HmsError, error) {
	ret, hmsErrs, err := InvokeDriverFunc(
		ids,
		FunctionCall{
			Invocation: runtime.FunctionInvocation{
				Function: homescript.DeviceFunctionReportPowerDraw,
				Args:     []value.Value{},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					homescript.DeviceReportPowerDrawSignature(errors.Span{}).Signature,
				),
			},
		},
	)

	if err != nil || hmsErrs != nil {
		return DeviceDimmableInformation{}, hmsErrs, err
	}

	percent := ret.ReturnValue.(value.ValueInt).Inner
	if percent < 0 || percent > 100 {
		return DeviceDimmableInformation{Percent: 0},
			[]homescript.HmsError{
				{
					SyntaxError:     nil,
					DiagnosticError: nil,
					RuntimeInterrupt: &homescript.HmsRuntimeInterrupt{
						Kind: "driver",
						Message: fmt.Sprintf(
							"Device function `%s` should return positive dim percent in range (0 <= x <= 100) but returned %d",
							homescript.DeviceFunctionReportPowerDraw,
							percent,
						),
					},
					Span: ret.CalledFunctionSpan,
				},
			}, nil
	}

	return DeviceDimmableInformation{
		Percent: uint8(percent),
	}, nil, nil
}

func InvokeDriverDim(
	deviceID,
	vendorID,
	modelID string,
	dimAction DriverActionDim,
) (DriverActionPowerOutput, []homescript.HmsError, error) {
	// TODO: add context support
	ctx, cancel := context.WithCancel(context.Background())

	runResult, hmsErrs, dbErr := invokeDriverGeneric(
		ctx,
		cancel,
		DriverContext{
			DeviceId: &deviceID,
		},
		vendorID,
		modelID,
		FunctionCall{
			Invocation: runtime.FunctionInvocation{
				Function: homescript.DeviceFuncionSetPower,
				Args: []value.Value{
					*value.NewValueInt(dimAction.Percent), // TODO: test this by providing an int for instance.
				},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					homescript.DeviceDimSignature(errors.Span{}).Signature,
				),
			},
		},
	)

	if dbErr != nil || hmsErrs != nil {
		return DriverActionPowerOutput{}, hmsErrs, dbErr
	}

	return DriverActionPowerOutput{
		Changed: runResult.ReturnValue.(value.ValueBool).Inner,
	}, nil, nil
}
