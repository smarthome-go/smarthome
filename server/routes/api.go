package routes

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
	SwitchName string `json:"switch"`
	PowerOn    bool   `json:"powerOn"`
}

type UserPermissionRequest struct {
	Username   string `json:"username"`
	Permission string `json:"permission"`
}

// API endpoint for manipulating power states and (de) activating sockets
// Authentication, permission and switch permission is needed to interact with this endpoint
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
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	userHasPermission, err := database.UserHasSwitchPermission(username, request.SwitchName)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "failed to check permission for this switch"})
		return
	}
	if !userHasPermission {
		log.Debug("User requested to use a switch but lacks permission to use it")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "permission denied", Error: "missing permission to interact with this switch, contact your administrator"})
		return
	}
	err = hardware.SetPower(request.SwitchName, request.PowerOn)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "hardware error", Error: "failed to communicate with hardware"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "power action successful"})
	if request.PowerOn {
		go event.Info("User Activated Switch", fmt.Sprintf("%s activated switch %s", username, request.SwitchName))
	} else {
		go event.Info("User Deactivated Switch", fmt.Sprintf("%s deactivated switch %s", username, request.SwitchName))
	}
}

// Returns a list of available switches as JSON to the user, no authentication required
func getSwitches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switches, err := database.ListSwitches()
	if err != nil {
		log.Error("Exception in getSwitches: database failure: ", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	json.NewEncoder(w).Encode(switches)
}

// Only returns switches which the user has access to, authentication required
func getUserSwitches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	switches, err := database.ListUserSwitches(username)
	if err != nil {
		log.Error("Exception in getUserSwitches: database failure: ", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	json.NewEncoder(w).Encode(switches)
}

// Returns a list of strings which resemble permissions of the currently logged in user, authentication required
func getUserPermissions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	permissions, err := database.GetUserPermissions(username)
	if err != nil {
		log.Error("Exception in getUserPermissions: database failure: ", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	json.NewEncoder(w).Encode(permissions)
}

// Returns a list of power states, no authentication required
// {SwitchId: string, Power: bool}
func getPowerStates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	powerStates, err := database.GetPowerStates()
	if err != nil {
		log.Error("Could not list powerstates: database failure: ", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	json.NewEncoder(w).Encode(powerStates)
}

// Triggers deletion of internal server logs which are older than 30 days, admin authentication required
func flushOldLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := database.FlushOldLogs()
	if err != nil {
		log.Error("Exception in flushOldLogs: database failure: ", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "failed to flush logs: database failure"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully flushed logs older than 30 days"})
}

// Triggers deletion of ALL internal server logs, admin authentication required
func flushAllLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := database.FlushAllLogs()
	if err != nil {
		log.Error("Exception in flushOldLogs: database failure: ", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "failed to flush logs: database failure"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully flushed logs"})
}

// Returns a list of logging items in the logging table, admin authentication required
func listLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	logs, err := database.GetLogs()
	if err != nil {
		log.Error("Failed to list logs: database failure: ", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "failed to get logs: database failure"})
		return
	}
	json.NewEncoder(w).Encode(logs)
}

// Adds a given permission to a given user, admin authentication required
// Takes a JSON body as `input` `{"username": "", "permission": ""}`
// If the permission is invalid, a `422` is returned
// TODO: test following endpoint against all scenarios
func addUserPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserPermissionRequest
	err := decoder.Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	modified, err := database.AddUserPermission(request.Username, request.Permission)
	if err != nil {
		if err.Error() == "permission not found error: unknown permission type" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(Response{Success: false, Message: "could not add permission to user", Error: "invalid permission type"})
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusNotModified)
		json.NewEncoder(w).Encode(Response{Success: true, Message: "user is already in possession of this permission", Error: ""})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Success: true, Message: fmt.Sprintf("successfully added permission `%s` to user `%s`", request.Permission, request.Username)})
}

// TODO: test following endpoint against all scenarios
// TODO: implement a boolean for permission removal that can be checked here
func removeUserPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserPermissionRequest
	err := decoder.Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	modified, err := database.RemoveUserPermission(request.Username, request.Permission)
	if err != nil {
		if err.Error() == "permission not found error: unknown permission type" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(Response{Success: false, Message: "could not remove permission from user", Error: "invalid permission type"})
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to remove permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusNotModified)
		json.NewEncoder(w).Encode(Response{Success: true, Message: "user does not have this permission"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully removed permission from user"})
}
