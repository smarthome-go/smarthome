package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smarthome-go/smarthome/core/device/driver"
)

type DeviceActionrequestBody struct {
	DeviceID string `json:"deviceId"`

	// TODO: use dynamic typing here?
	// Or use separate API endpoint for each intent?
	Power *driver.DriverSetPowerInput `json:"power"`
	Dim   *driver.DriverDimInput      `json:"dim"`
}

func DeviceActionHandlerFactory(action driver.DriverActionKind) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var request DeviceActionrequestBody
		if err := decoder.Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
			return
		}

		res, found, validationErr, backendErr := driver.Manager.DeviceAction(
			action,
			request.DeviceID,
			request.Power,
			request.Dim,
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
