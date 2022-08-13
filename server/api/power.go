package api

import (
	"encoding/json"
	"net/http"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type PowerRequest struct {
	Switch  string `json:"switch"`
	PowerOn bool   `json:"powerOn"`
}

// API endpoint for manipulating power states and (de) activating sockets, authentication required
// Permission and switch permission is needed to interact with this endpoint
func PowerPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request PowerRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	_, switchExists, err := database.GetSwitchById(request.Switch)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to check existence of this switch", Error: "database error"})
		return
	}
	if !switchExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to set power: invalid switch id", Error: "switch not found"})
		return
	}
	userHasPermission, err := database.UserHasSwitchPermission(username, request.Switch)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to check permission for this switch", Error: "database error"})
		return
	}
	if !userHasPermission {
		w.WriteHeader(http.StatusForbidden)
		Res(w, Response{Success: false, Message: "permission denied", Error: "missing permission to interact with this switch, contact your administrator"})
		return
	}
	if err := hardware.SetPower(request.Switch, request.PowerOn); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "hardware error", Error: "failed to communicate with hardware"})
		return
	}
	Res(w, Response{Success: true, Message: "power action successful"})
}

// Returns a list of power states, no authentication required
// Request: empty | Response: `[{"switchId": "x", power: false}, {...}]`
func GetPowerStates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	powerStates, err := database.GetPowerStates()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(powerStates); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to get power states", Error: "could not encode content"})
	}
}

// Returns the power draw points from the last 24 hours
func GetPowerDrawFrom24Hours(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	powerUsageData, err := hardware.GetPowerUsageRecordsUnixMillis(24)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "could not get power usage data from the last 24 hours", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(powerUsageData); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to get power usage data from the last 24 hours", Error: "could not encode content"})
	}
}
