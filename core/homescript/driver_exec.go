package homescript

import (
	"bytes"
	"context"
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"

	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
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
) (HmsRunResultContext, []HmsError, error) {
	driver, found, err := database.GetDeviceDriver(vendorId, modelId)
	if err != nil {
		return HmsRunResultContext{}, nil, err
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
	contextSingletons[DriverSingletonIdent] = driverSingleton

	// Load device singleton if possible.
	var deviceSingleton value.Value
	if driverCtx.DeviceId != nil {
		deviceId := *driverCtx.DeviceId
		deviceSingleton, found = DeviceStore[deviceId]
		if !found {
			panic(fmt.Sprintf("Device singleton of `%s` not found in store", deviceId))
		}
		contextSingletons[DriverDeviceSingletonIdent] = deviceSingleton
	}

	// TODO: load corresponding device singleton.
	hmsRes, resultContext, err := HmsManager.Run(
		HMS_PROGRAM_KIND_DEVICE_DRIVER,
		&AnalyzerDriverMetadata{
			VendorID: vendorId,
			ModelID:  modelId,
		},
		"", // TODO: fix username requirement.
		&filename,
		driver.HomescriptCode,
		InitiatorAPI,
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
		return HmsRunResultContext{}, nil, err
	}

	if !hmsRes.Success {
		errors := make([]HmsError, 0)

		// Filter out any non-error messages.
		for _, d := range hmsRes.Errors {
			if d.DiagnosticError != nil && d.DiagnosticError.Level != diagnostic.DiagnosticLevelError {
				continue
			}
			errors = append(errors, d)
		}

		return HmsRunResultContext{}, errors, nil
	}

	// Get driver and device singleton.
	driverSingletonAfter, found := resultContext.Singletons[DriverSingletonIdent]
	if !found {
		panic(fmt.Sprintf("Driver singleton (`%s`) not found after driver execution", DriverSingletonIdent))
	}

	// Save driver singleton state after VM has terminated.
	driverMarshaled, _ := value.MarshalValue(driverSingletonAfter, false)
	if err := StoreDriverSingletonConfigUpdate(driver.VendorId, driver.ModelId, driverMarshaled); err != nil {
		return HmsRunResultContext{}, nil, err
	}

	// Save device singleton state after VM has terminated (if device was even loaded).
	if driverCtx.DeviceId != nil {
		deviceMarshaled, _ := value.MarshalValue(deviceSingleton, false)
		if err := StoreDeviceSingletonConfigUpdate(*driverCtx.DeviceId, deviceMarshaled); err != nil {
			return HmsRunResultContext{}, nil, err
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
) (HmsRunResultContext, []HmsError, error) {
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
		return HmsRunResultContext{}, hmsErrs, dbErr
	}

	return runResult, nil, nil
}

func InvokeDriverReportPowerState(
	ids DriverInvocationIDs,
) (DriverActionGetPowerStateOutput, []HmsError, error) {
	ret, hmsErrs, err := InvokeDriverFunc(
		ids,
		FunctionCall{
			Invocation: runtime.FunctionInvocation{
				Function: DeviceFunctionReportPowerState,
				Args:     []value.Value{},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					DeviceReportPowerStateSignature(errors.Span{}).Signature,
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
) (DriverActionGetPowerDrawOutput, []HmsError, error) {
	ret, hmsErrs, err := InvokeDriverFunc(
		ids,
		FunctionCall{
			Invocation: runtime.FunctionInvocation{
				Function: DeviceFunctionReportPowerDraw,
				Args:     []value.Value{},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					DeviceReportPowerDrawSignature(errors.Span{}).Signature,
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
			[]HmsError{
				{
					SyntaxError:     nil,
					DiagnosticError: nil,
					RuntimeInterrupt: &HmsRuntimeInterrupt{
						Kind: "driver",
						Message: fmt.Sprintf(
							"Device function `%s` should return positive power consumption but returned %d",
							DeviceFunctionReportPowerDraw,
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
) (DriverActionPowerOutput, []HmsError, error) {
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
				Function: DeviceFuncionSetPower,
				Args: []value.Value{
					*value.NewValueBool(powerAction.State), // TODO: test this by providing an int for instance.
				},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					DeviceSetPowerSignature(errors.Span{}).Signature,
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

func normalizeRange(input value.ValueRange) (lower, upper int64) {
	start := (*input.Start).(value.ValueInt).Inner
	end := (*input.End).(value.ValueInt).Inner

	if end < start {
		if !input.EndIsInclusive {
			end++
		}
		return end, start
	}

	if !input.EndIsInclusive {
		end--
	}
	return start, end
}

func InvokeDriverReportDimmable(
	ids DriverInvocationIDs,
) ([]DriverActionReportDimOutput, []HmsError, error) {
	ret, hmsErrs, err := InvokeDriverFunc(
		ids,
		FunctionCall{
			Invocation: runtime.FunctionInvocation{
				Function: DeviceFunctionReportDim,
				Args:     []value.Value{},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					DeviceReportDimSignature(errors.Span{}).Signature,
				),
			},
		},
	)

	if err != nil || hmsErrs != nil {
		return nil, hmsErrs, err
	}

	values := *ret.ReturnValue.(value.ValueList).Values
	dimmables := make([]DriverActionReportDimOutput, len(values))

	for idx, currentListElement := range values {
		fields := (*currentListElement).(value.ValueObject).FieldsInternal

		value_ := (*fields[ReportDimTypeValueIdent]).(value.ValueInt).Inner
		label := (*fields[ReportDimTypeLabelIdent]).(value.ValueString).Inner
		range_ := (*fields[ReportDimTypeRangeIdent]).(value.ValueRange)

		lower, upper := normalizeRange(range_)

		if value_ < lower || value_ > upper {
			return nil,
				[]HmsError{
					{
						SyntaxError:     nil,
						DiagnosticError: nil,
						RuntimeInterrupt: &HmsRuntimeInterrupt{
							Kind: "driver",
							Message: fmt.Sprintf(
								"Device function `%s` should return value in dimmable range(%d..%d) but returned %d for label `%s`",
								DeviceFunctionReportDim,
								lower,
								upper,
								value_,
								label,
							),
						},
						Span: ret.CalledFunctionSpan,
					},
				}, nil
		}

		dimmables[idx] = DriverActionReportDimOutput{
			Value: value_,
			Label: label,
			Range: DriverActionReportRange{
				Lower: lower,
				Upper: upper,
			},
		}
	}

	return dimmables, nil, nil
}

func InvokeDriverDim(
	deviceID,
	vendorID,
	modelID string,
	dimAction DriverActionDim,
) (DriverActionPowerOutput, []HmsError, error) {
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
				Function: DeviceFuncionSetDim,
				Args: []value.Value{
					*value.NewValueString(dimAction.Label), // TODO: test this by providing an int for instance.
					*value.NewValueInt(dimAction.Value),    // TODO: test this by providing an int for instance.
				},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					DeviceDimSignature(errors.Span{}).Signature,
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
