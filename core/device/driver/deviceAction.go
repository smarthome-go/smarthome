package driver

import (
	"errors"
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

//
// Action requests.
//

//
// Action-specific inputs.
//

type DriverSetPowerInput struct {
	State bool `json:"state"`
}

type DriverDimInput struct {
	Value int64  `json:"percent"`
	Label string `json:"label"`
}

//
// Action responses.
//

type ActionResponse struct {
	Success   bool                      `json:"success"`
	HmsErrors []types.HmsError          `json:"hmsErrors"`
	Output    DriverActionOutputPayload `json:"output"`
}

// TODO: make this function signature better
// Use interfaces here
func (d DriverManager) DeviceAction(action DriverActionKind, deviceID string, Power *DriverSetPowerInput, Dim *DriverDimInput) (
	res ActionResponse,
	deviceFound bool,
	httpErr error,
	err error,
) {
	device, found, err := database.GetDeviceById(deviceID)
	if !found || err != nil {
		return ActionResponse{}, false, nil, err
	}

	var out DriverActionOutputPayload
	var hmsErrs []types.HmsError

	// Invoke driver.
	switch action {
	case DriverActionKindHealthCheck:
		// TODO: implement this
		panic("TODO")
	case DriverActionKindReportPowerState:
		// TODO: implement this
		panic("TODO")
	case DriverActionKindReportPowerDraw:
		// TODO: implement this
		panic("TODO")
	case DriverActionKindDim:
		if Dim == nil {
			return ActionResponse{},
				true,
				errors.New("Dim action field is missing even though it is required"),
				nil
		}
		out, hmsErrs, err = d.InvokeDriverDim(
			device.ID,
			device.VendorID,
			device.ModelID,
			DriverActionDim{
				Value: Dim.Value,
				Label: Dim.Label,
			},
		)
	case DriverActionKindSetPower:
		if Power == nil {
			return ActionResponse{},
				true,
				errors.New("Power action field is missing even though it is required"),
				nil
		}
		out, hmsErrs, err = d.InvokeDriverSetPower(
			device.ID,
			device.VendorID,
			device.ModelID,
			DriverActionPower{State: Power.State},
		)
	default:
		panic(fmt.Sprintf("A new device action kind was added without updating this code: `%d`", action))
	}

	if err != nil {
		return ActionResponse{}, false, nil, err
	}

	return ActionResponse{
		Success:   len(hmsErrs) == 0,
		HmsErrors: hmsErrs,
		Output:    out,
	}, true, nil, err
}
