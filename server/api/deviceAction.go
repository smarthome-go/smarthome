package api

import (
	"encoding/json"
	"net/http"

	"github.com/smarthome-go/smarthome/core/drivers"
)

var allowedActions = []drivers.DeviceActionType{
	drivers.DeviceActionTypePower,
}

func DeviceAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request drivers.DeviceActionRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	res, found, err := drivers.DeviceAction(request)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to execute device action", Error: err.Error()})
		return
	}
	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to execute device action", Error: "no device with id exists"})
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(err.Error())
	}
}
