package routes

import (
	"encoding/json"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/hardware"
)

type PowerRequest struct {
	SwitchName string `json:"switch"`
	PowerOn    bool   `json:"powerOn"`
}

func powerPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request PowerRequest
	err := decoder.Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	err = hardware.SetPower(request.SwitchName, request.PowerOn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "hardware error", Error: "failed to communicate with hardware"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "power action successful", Error: ""})
}
