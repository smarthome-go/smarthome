package driver

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"

	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	driverTypes "github.com/smarthome-go/smarthome/core/device/driver/types"
	"github.com/smarthome-go/smarthome/core/homescript/types"
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

func (d *DriverManager) invokeDriverGeneric(
	// Termination handling
	cancelCtx context.Context,
	cancelFunc context.CancelFunc,
	// Call data
	driverCtx DriverContext,
	vendorId,
	modelId string,
	functionInvocation FunctionCall,
) (types.HmsRes, error) {
	driver, found, err := database.GetDeviceDriver(vendorId, modelId)
	if err != nil {
		return types.HmsRes{}, err
	}

	if !found {
		panic(fmt.Sprintf("Driver `%s:%s` not found in the database", vendorId, modelId))
	}

	// TODO: attaching to existing VM or syncing state accross instances

	var outputBuffer bytes.Buffer

	// TODO: do not hardcode stuff like this
	// filename := fmt.Sprintf("@driver:%s:%s", vendorId, modelId)

	contextSingletons := make(map[string]value.Value)

	// Load driver singleton.
	driverSingleton, found := DriverStore[database.DriverTuple{
		VendorID: vendorId,
		ModelID:  modelId,
	}]
	if !found {
		panic(fmt.Sprintf("Driver singleton of driver `%s:%s` not found in store", vendorId, modelId))
	}
	contextSingletons[DriverSingletonIdent] = driverSingleton

	if driverCtx.DeviceId != nil && driver.SingletonJSON != nil {
		disp, err := driverSingleton.DisplayFlat()
		if err != nil {
			panic(err)
		}
		log.Debugf("Loaded driver singleton for device ID `%s` from database: `%s`", *driverCtx.DeviceId, disp)
	}

	// Load device singleton if possible.
	var deviceSingleton value.ValueObject
	if driverCtx.DeviceId != nil {
		deviceId := *driverCtx.DeviceId
		deviceSingleton, found = DeviceStore[deviceId]
		if !found {
			panic(fmt.Sprintf("Device singleton of `%s` not found in store", deviceId))
		}
		contextSingletons[DriverDeviceSingletonIdent] = deviceSingleton

		disp, err := deviceSingleton.DisplayFlat()
		if err != nil {
			panic(err)
		}

		log.Debugf("Loaded driver singleton for device ID `%s` from database: %s", *driverCtx.DeviceId, disp)
	}

	var deviceId string
	if driverCtx.DeviceId != nil {
		deviceId = *driverCtx.DeviceId
	}

	res, err := d.Hms.RunDriverScript(
		driverTypes.DriverInvocationIDs{
			DeviceID: &deviceId,
			VendorID: vendorId,
			ModelID:  modelId,
		},
		functionInvocation.Invocation,
		types.Cancelation{
			Context:    cancelCtx,
			CancelFunc: cancelFunc,
		},
		&outputBuffer,
	)

	// hmsRes, resultContext, err := d.Hms.Run(
	// 	types.HMS_PROGRAM_KIND_DEVICE_DRIVER,
	// 	&driverTypes.DriverInvocationIDs{
	// 		DeviceID: deviceId,
	// 		VendorID: vendorId,
	// 		ModelID:  modelId,
	// 	},
	// 	"", // TODO: fix username requirement.
	// 	&filename,
	// 	driver.HomescriptCode,
	// 	types.InitiatorAPI,
	// 	cancelCtx,
	// 	cancelFunc,
	// 	nil,
	// 	make(map[string]string),
	// 	&outputBuffer,
	// 	nil,
	// 	&functionInvocation.Invocation,
	// 	contextSingletons,
	// )

	if err != nil {
		return types.HmsRes{}, err
	}

	if res.Errors.ContainsError {
		errorList := make([]types.HmsError, 0)

		// Filter out any non-error messages.
		for _, d := range res.Errors.Diagnostics {
			if d.DiagnosticError != nil && d.DiagnosticError.Level != diagnostic.DiagnosticLevelError {
				continue
			}
			errorList = append(errorList, d)
		}

		return types.HmsRes{
			Errors: types.HmsDiagnosticsContainer{
				ContainsError: true,
				Diagnostics:   errorList,
				FileContents:  res.Errors.FileContents,
			},
			Singletons:  nil,
			ReturnValue: nil,
			// TODO: called function span
			// CalledFunctionSpan: errors.Span{},
		}, nil
	}

	// Get driver and device singleton.
	driverSingletonAfter, found := res.Singletons[DriverSingletonIdent]
	if !found {
		panic(fmt.Sprintf("Driver singleton (`%s`) not found after driver execution", DriverSingletonIdent))
	}

	// Save driver singleton state after VM has terminated.
	driverMarshaled, _ := value.MarshalValue(driverSingletonAfter, false)
	if err := d.StoreDriverSingletonConfigUpdate(driver.VendorID, driver.ModelID, driverMarshaled); err != nil {
		return types.HmsRes{}, err
	}

	// Save device singleton state after VM has terminated (if device was even loaded).
	if driverCtx.DeviceId != nil {
		deviceMarshaled, _ := value.MarshalValue(deviceSingleton, false)
		if err := d.StoreDeviceSingletonConfigUpdate(*driverCtx.DeviceId, deviceMarshaled); err != nil {
			return types.HmsRes{}, err
		}
	}

	return res, nil
}

//
//
//
// Specialized driver action invocations.
//
//
//

//
// TDOO: maybe implement a function factory to create those almost identical functions more ideomatically.
//

func (d DriverManager) InvokeDriverFunc(
	ids driverTypes.DriverInvocationIDs,
	call FunctionCall,
) (types.HmsRes, error) {
	if ids.VendorID == "" || ids.ModelID == "" || ids.DeviceID == nil {
		panic("One or more ids in the driver-device triplet were empty or <nil>")
	}

	// TODO: add context support
	ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel = context.WithTimeout(ctx, time.Second*10)

	runResult, dbErr := d.invokeDriverGeneric(
		ctx,
		cancel,
		DriverContext{
			DeviceId: ids.DeviceID,
		},
		ids.VendorID,
		ids.ModelID,
		call,
	)

	if dbErr != nil {
		return types.HmsRes{}, dbErr
	}

	return runResult, nil
}

func (d DriverManager) InvokeValidateCheckDriver(ids driverTypes.DriverInvocationIDs) ([]types.HmsError, error) {
	res, err := d.InvokeDriverFunc(
		ids,
		FunctionCall{
			Invocation: runtime.FunctionInvocation{
				Function: DeviceFunctionValidateDriver,
				Args:     []value.Value{},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					deviceValidateDeviceOrDriverSignature(errors.Span{}).Signature,
				),
			},
		},
	)

	if err != nil || res.Errors.ContainsError {
		return res.Errors.Diagnostics, err
	}

	return nil, nil
}

func (d DriverManager) InvokeDriverReportSensors(
	ids driverTypes.DriverInvocationIDs,
) ([]DriverActionReportSensorReadingsOutput, []types.HmsError, error) {
	ret, err := d.InvokeDriverFunc(
		ids,
		FunctionCall{
			Invocation: runtime.FunctionInvocation{
				Function: DeviceFunctionReportSensorReadings,
				Args:     []value.Value{},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					DeviceReportSensorReadingsSignature(errors.Span{}).Signature,
				),
			},
		},
	)

	if err != nil || ret.Errors.ContainsError {
		return nil, ret.Errors.Diagnostics, err
	}

	values := *ret.ReturnValue.(value.ValueList).Values
	readings := make([]DriverActionReportSensorReadingsOutput, len(values))

	for idx, currentListElement := range values {
		fields := (*currentListElement).(value.ValueObject).FieldsInternal

		value_ := (*fields[ReportDimTypeValueIdent])
		label := (*fields[ReportSensorTypeLabelIdent]).(value.ValueString).Inner
		unit := (*fields[ReportSensorTypeUnitIdent]).(value.ValueString).Inner

		if !value_.Kind().TypeKind().IsPrimitive() {
			return nil,
				[]types.HmsError{
					{
						SyntaxError:     nil,
						DiagnosticError: nil,
						RuntimeInterrupt: &types.HmsRuntimeInterrupt{
							Kind: "driver",
							Message: fmt.Sprintf(
								"Device function `%s` should return values of primitive data type but returned value `%s` of type `%s`",
								DeviceFunctionReportSensorReadings,
								label,
								value_.Kind(),
							),
						},
						Span: ret.CalledFunctionSpan,
					},
				}, nil
		}

		valueMarshaled, _ := value.MarshalValue(value_, false)
		readings[idx] = DriverActionReportSensorReadingsOutput{
			Label:       label,
			Value:       valueMarshaled,
			HmsTypeKind: value_.Kind().TypeKind().String(),
			Unit:        unit,
		}
	}

	return readings, nil, nil
}

func (d DriverManager) InvokeDriverReportPowerState(
	ids driverTypes.DriverInvocationIDs,
) (DriverActionGetPowerStateOutput, []types.HmsError, error) {
	ret, err := d.InvokeDriverFunc(
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

	if err != nil || ret.Errors.ContainsError {
		return DriverActionGetPowerStateOutput{}, ret.Errors.Diagnostics, err
	}

	return DriverActionGetPowerStateOutput{
		State: ret.ReturnValue.(value.ValueBool).Inner,
	}, nil, nil
}

func (d DriverManager) InvokeDriverReportPowerDraw(
	ids driverTypes.DriverInvocationIDs,
) (DriverActionGetPowerDrawOutput, []types.HmsError, error) {
	ret, err := d.InvokeDriverFunc(
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

	if err != nil || ret.Errors.ContainsError {
		return DriverActionGetPowerDrawOutput{}, ret.Errors.Diagnostics, err
	}

	wattsRaw := ret.ReturnValue.(value.ValueInt).Inner
	if wattsRaw < 0 {
		return DriverActionGetPowerDrawOutput{Watts: 0},
			[]types.HmsError{
				{
					SyntaxError:     nil,
					DiagnosticError: nil,
					RuntimeInterrupt: &types.HmsRuntimeInterrupt{
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

func (d DriverManager) InvokeDriverSetPower(
	deviceID,
	vendorID,
	modelID string,
	powerAction DriverActionPower,
) (DriverActionPowerOutput, []types.HmsError, error) {
	// TODO: add context support
	ctx, cancel := context.WithCancel(context.Background())

	runResult, dbErr := d.invokeDriverGeneric(
		ctx,
		cancel,
		DriverContext{
			DeviceId: &deviceID,
		},
		vendorID,
		modelID,
		FunctionCall{
			Invocation: runtime.FunctionInvocation{
				Function: DeviceFunctionSetPower,
				Args: []value.Value{
					*value.NewValueBool(powerAction.State), // TODO: test this by providing an int for instance.
				},
				FunctionSignature: runtime.FunctionInvocationSignatureFromType(
					DeviceSetPowerSignature(errors.Span{}).Signature,
				),
			},
		},
	)

	if dbErr != nil || runResult.Errors.ContainsError {
		return DriverActionPowerOutput{}, runResult.Errors.Diagnostics, dbErr
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

func (d DriverManager) InvokeDriverReportDimmable(
	ids driverTypes.DriverInvocationIDs,
) ([]DriverActionReportDimOutput, []types.HmsError, error) {
	res, err := d.InvokeDriverFunc(
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

	if err != nil || res.Errors.ContainsError {
		return nil, res.Errors.Diagnostics, err
	}

	values := *res.ReturnValue.(value.ValueList).Values
	dimmables := make([]DriverActionReportDimOutput, len(values))

	for idx, currentListElement := range values {
		fields := (*currentListElement).(value.ValueObject).FieldsInternal

		value_ := (*fields[ReportDimTypeValueIdent]).(value.ValueInt).Inner
		label := (*fields[ReportDimTypeLabelIdent]).(value.ValueString).Inner
		range_ := (*fields[ReportDimTypeRangeIdent]).(value.ValueRange)

		lower, upper := normalizeRange(range_)

		if value_ < lower || value_ > upper {
			return nil,
				[]types.HmsError{
					{
						SyntaxError:     nil,
						DiagnosticError: nil,
						RuntimeInterrupt: &types.HmsRuntimeInterrupt{
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
						Span: res.CalledFunctionSpan,
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

func (d DriverManager) InvokeDriverDim(
	deviceID,
	vendorID,
	modelID string,
	dimAction DriverActionDim,
) (DriverActionDimOutput, []types.HmsError, error) {
	// TODO: add context support
	ctx, cancel := context.WithCancel(context.Background())

	res, dbErr := d.invokeDriverGeneric(
		ctx,
		cancel,
		DriverContext{
			DeviceId: &deviceID,
		},
		vendorID,
		modelID,
		FunctionCall{
			Invocation: runtime.FunctionInvocation{
				Function: DeviceFunctionSetDim,
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

	if dbErr != nil || res.Errors.ContainsError {
		return DriverActionDimOutput{}, nil, dbErr
	}

	return DriverActionDimOutput{
		Changed: res.ReturnValue.(value.ValueBool).Inner,
	}, nil, nil
}
