package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smarthome-go/smarthome/core/drivers"
)

func DeviceActionHandlerFactory(action drivers.DriverActionKind) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var request drivers.DeviceActionrequestBody
		if err := decoder.Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
			return
		}

		res, found, validationErr, backendErr := drivers.DeviceAction(
			action,
			request,
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
