package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/core/hardware"
	"github.com/MikMuellerDev/smarthome/server/middleware"
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
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	switchExists, err := database.DoesSwitchExist(request.Switch)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to check existence of this switch", Error: "database error"})
		return
	}
	if !switchExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to set power: invalid switch id", Error: "switch not found"})
		return
	}
	userHasPermission, err := database.UserHasSwitchPermission(username, request.Switch)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to check permission for this switch", Error: "database error"})
		return
	}
	if !userHasPermission {
		log.Debug("User requested to use a switch but lacks permission to use it")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "permission denied", Error: "missing permission to interact with this switch, contact your administrator"})
		return
	}
	if err := hardware.SetPower(request.Switch, request.PowerOn); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "hardware error", Error: "failed to communicate with hardware"})
		go event.Warn("Hardware Error", fmt.Sprintf("The hardware failed while %s tried to interact with switch %s.", username, request.Switch))
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "power action successful"})
	if request.PowerOn {
		go event.Info("User Activated Switch", fmt.Sprintf("%s activated switch %s", username, request.Switch))
	} else {
		go event.Info("User Deactivated Switch", fmt.Sprintf("%s deactivated switch %s", username, request.Switch))
	}
}

// Returns a list of available switches as JSON to the user, no authentication required
func GetSwitches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switches, err := database.ListSwitches()
	if err != nil {
		log.Error("Exception in getSwitches: database failure: ", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	json.NewEncoder(w).Encode(switches)
}

// Only returns switches which the user has access to, authentication required
func GetUserSwitches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	switches, err := database.ListUserSwitches(username)
	if err != nil {
		log.Error("Exception in getUserSwitches: database failure: ", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	json.NewEncoder(w).Encode(switches)
}

// Returns a list of power states, no authentication required
// Request: empty | Response: `[{"switchId": "x", power: false}, {...}]`
func GetPowerStates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	powerStates, err := database.GetPowerStates()
	if err != nil {
		log.Error("Could not list powerstates: database failure: ", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	json.NewEncoder(w).Encode(powerStates)
}

// Returns the wanted response for the frontend
func GetUserRoomsWithSwitches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	rooms, err := database.ListPersonalRoomsAll(username)
	if err != nil {
		log.Error("Could not list user rooms: database failure: ", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	json.NewEncoder(w).Encode(rooms)
}
