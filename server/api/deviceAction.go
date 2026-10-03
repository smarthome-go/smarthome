package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/device/driver"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type DeviceActionrequestBody struct {
	DeviceID string `json:"deviceId"`

	// TODO: use dynamic typing here?
	// Or use separate API endpoint for each intent?
	Power *driver.DriverSetPowerInput `json:"power"`
	Dim   *driver.DriverDimInput      `json:"dim"`
	Color *driver.DriverColorInput    `json:"color"`
}

func DeviceActionHandlerFactory(action driver.DriverActionKind) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		username, err := middleware.GetUserFromCurrentSession(w, r)
		if err != nil {
			// `GetUserFromCurrentSession` already answered the request
			return
		}

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var request DeviceActionrequestBody
		if err := decoder.Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
			return
		}

		// Validate that the user is allowed to interact with this device.
		// Note: users with the `modifyRooms` permission implicitly have
		// access to every device, see `database.UserHasDevicePermission`.
		hasDevicePermission, err := database.UserHasDevicePermission(username, request.DeviceID)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			Res(w, Response{Success: false, Message: "failed to execute device action", Error: "database failure"})
			return
		}
		if !hasDevicePermission {
			w.WriteHeader(http.StatusForbidden)
			Res(w, Response{Success: false, Message: "failed to execute device action", Error: fmt.Sprintf("you lack permission to interact with the device `%s`", request.DeviceID)})
			return
		}

		res, found, validationErr, backendErr := driver.Manager.DeviceAction(
			action,
			request.DeviceID,
			request.Power,
			request.Dim,
			request.Color,
		)

		if backendErr != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			Res(w, Response{Success: false, Message: "failed to execute device action", Error: backendErr.Error()})
			return
		}

		if !found {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Res(w, Response{Success: false, Message: "failed to execute device action", Error: fmt.Sprintf("no device with id `%s` exists", request.DeviceID)})
			return
		}

		if validationErr != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Res(w, Response{Success: false, Message: "failed to execute device action", Error: fmt.Sprintf("validation error: %s", validationErr.Error())})
			return
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			panic(err.Error())
		}
	}
}
