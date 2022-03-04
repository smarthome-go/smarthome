package routes

import (
	"encoding/json"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/hardware"
)

type PowerRequest struct {
	SwitchName string `json:"switch"`
	PowerOn    bool   `json:"powerOn"`
}

// API endpoint for manipulating power states and (de) activating sockets
// TODO: implement authentication middleware
// TODO: implement permission middleware
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

// Returns a list of available switches
func getSwitches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switches, err := database.ListSwitches()
	if err != nil {
		log.Error("Exception in getSwitches: database failure: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	json.NewEncoder(w).Encode(switches)
}
