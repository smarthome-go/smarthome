package drivers

import (
	"bytes"
	"context"
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type DeviceActionType string

const (
	DeviceActionTypePower DeviceActionType = "power"
)

type PowerAction struct {
	State bool `json:"state"`
}

type DeviceActionRequest struct {
	Type     DeviceActionType `json:"type"`
	DeviceID string           `json:"deviceId"`

	// TODO: use dynamic typing here?
	// Or use separate API endpoint for each intent?
	Power *PowerAction `json:"power"`
}

type ActionResponse struct {
	Success   bool                      `json:"success"`
	HmsErrors []homescript.HmsError     `json:"hmsErrors"`
	Output    DriverActionOutputPayload `json:"output"`
}

func DeviceAction(action DeviceActionRequest) (res ActionResponse, deviceFound bool, err error) {
	device, found, err := database.GetDeviceById(action.DeviceID)
	if !found || err != nil {
		return ActionResponse{}, false, err
	}

	// Invoke driver.
	switch action.Type {
	case DeviceActionTypePower:
		if action.Power == nil {
			panic("Power action field is `nil` even though it is required")
		}

		out, hmsErrs, err := InvokeDriverPower(
			device.Id,
			device.VendorId,
			device.ModelId,
			DriverActionPower{
				State: action.Power.State,
			},
		)

		if err != nil {
			return ActionResponse{}, false, err
		}

		return ActionResponse{
			Success:   len(hmsErrs) == 0,
			HmsErrors: hmsErrs,
			Output:    out,
		}, true, err
	default:
		panic(fmt.Sprintf("A new device action kind was added without updating this code: `%s`", action.Type))
	}
}

//
//
//
// Specialized driver action invocations.
//
//
//

func InvokeDriverPower(
	deviceId,
	vendorId,
	modelId string,
	action DriverActionPower,
) (DriverActionPowerOutput, []homescript.HmsError, error) {
	driver, found, err := database.GetDeviceDriver(vendorId, modelId)
	if err != nil {
		return DriverActionPowerOutput{}, nil, err
	}

	if !found {
		panic(fmt.Sprintf("Driver `%s:%s` not found in the database", vendorId, modelId))
	}

	ctx, cancel := context.WithCancel(context.Background())

	var outputBuffer bytes.Buffer

	filename := fmt.Sprintf("@driver:%s:%s", vendorId, modelId)

	const POWER_DRIVER_FUNCTION = "set_power"

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

	// Load device singleton.
	deviceSingleton, found := DeviceStore[deviceId]
	if !found {
		panic(fmt.Sprintf("Device singleton of `%s` not found in store", deviceId))
	}
	contextSingletons[homescript.DriverDeviceSingletonIdent] = deviceSingleton

	// TODO: load corresponding device singleton.
	hmsRes, finalContext, err := homescript.HmsManager.Run(
		homescript.HMS_PROGRAM_KIND_DEVICE_DRIVER,
		&homescript.AnalyzerDriverMetadata{
			VendorID: vendorId,
			ModelID:  modelId,
		},
		"", // TODO: fix username requirement.
		&filename,
		driver.HomescriptCode,
		homescript.InitiatorAPI,
		ctx,
		cancel,
		nil,
		make(map[string]string),
		&outputBuffer,
		nil,
		&runtime.FunctionInvocation{
			Function: POWER_DRIVER_FUNCTION,
			Args: []value.Value{
				// Use the power state as an argument.
				*value.NewValueBool(action.State),
			},
		},
		contextSingletons,
	)

	if err != nil {
		return DriverActionPowerOutput{}, nil, err
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

		return DriverActionPowerOutput{
			Changed: false,
		}, errors, nil
	}

	// Get driver and device singleton.
	driverSingletonAfter, found := finalContext.Singletons[homescript.DriverSingletonIdent]
	if !found {
		panic(fmt.Sprintf("Driver singleton (`%s`) not found after driver execution", homescript.DriverSingletonIdent))
	}

	// Save driver & device singleton state after VM has terminated.
	driverMarshaled, _ := value.MarshalValue(driverSingletonAfter, false)
	if err := StoreDriverSingletonConfigUpdate(driver.VendorId, driver.ModelId, driverMarshaled); err != nil {
		return DriverActionPowerOutput{}, nil, err
	}

	// Save device singleton state after VM has terminated.
	deviceMarshaled, _ := value.MarshalValue(deviceSingleton, false)
	if err := StoreDeviceSingletonConfigUpdate(deviceId, deviceMarshaled); err != nil {
		return DriverActionPowerOutput{}, nil, err
	}

	//
	// TODO: create API for spawning HMS functions.
	//

	return DriverActionPowerOutput{
		Changed: false, // TODO: determine this!
	}, nil, nil
}
