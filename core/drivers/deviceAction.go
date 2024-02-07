package drivers

import (
	"errors"
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

//
// Action requests.
//

type DeviceActionrequestBody struct {
	DeviceID string `json:"deviceId"`

	// TODO: use dynamic typing here?
	// Or use separate API endpoint for each intent?
	Power *DriverSetPowerInput `json:"power"`
}

//
// Action-specific inputs.
//

type DriverSetPowerInput struct {
	State bool `json:"state"`
}

//
// Action responses.
//

type ActionResponse struct {
	Success   bool                      `json:"success"`
	HmsErrors []homescript.HmsError     `json:"hmsErrors"`
	Output    DriverActionOutputPayload `json:"output"`
}

func DeviceAction(action DriverActionKind, body DeviceActionrequestBody) (
	res ActionResponse,
	deviceFound bool,
	httpErr error,
	err error,
) {
	device, found, err := database.GetDeviceById(body.DeviceID)
	if !found || err != nil {
		return ActionResponse{}, false, nil, err
	}

	var out DriverActionOutputPayload
	var hmsErrs []homescript.HmsError

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
		// TODO: implement this
		panic("TODO")
	case DriverActionKindSetPower:
		if body.Power == nil {
			return ActionResponse{}, true, errors.New("Power action field is missing even though it is required"), nil
		}
		out, hmsErrs, err = InvokeDriverSetPower(
			device.Id,
			device.VendorId,
			device.ModelId,
			DriverActionPower{State: body.Power.State},
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
