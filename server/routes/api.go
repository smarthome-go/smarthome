package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/core/hardware"
	"github.com/MikMuellerDev/smarthome/core/user"
	"github.com/MikMuellerDev/smarthome/core/utils"
	"github.com/MikMuellerDev/smarthome/server/middleware"
	"github.com/MikMuellerDev/smarthome/services/camera"
)

type PowerRequest struct {
	Switch  string `json:"switch"`
	PowerOn bool   `json:"powerOn"`
}

type UserPermissionRequest struct {
	Username   string `json:"username"`
	Permission string `json:"permission"`
}

// TODO: check if `Switch` is called somewhat else in other places
type UserSwitchPermissionRequest struct {
	Username string `json:"username"`
	Switch   string `json:"switch"`
}

type AddUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RemoveUserRequest struct {
	Username string `json:"username"`
}

type NotificationCountResponse struct {
	NotificationCount uint16 `json:"count"`
}

// API endpoint for manipulating power states and (de) activating sockets, authentication required
// Permission and switch permission is needed to interact with this endpoint
func powerPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request PowerRequest
	if err := decoder.Decode(&request); err != nil {
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
func getSwitches(w http.ResponseWriter, r *http.Request) {
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
func getUserSwitches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
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

// Returns a list of strings which resemble permissions of the currently logged in user, authentication required
// Request: empty | Response: `["a", "b", "c"]`
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
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	json.NewEncoder(w).Encode(permissions)
}

// Returns a list of power states, no authentication required
// Request: empty | Response: `[{"switchId": "x", power: false}, {...}]`
func getPowerStates(w http.ResponseWriter, r *http.Request) {
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

// Triggers deletion of internal server logs which are older than 30 days, admin authentication required
// Request: empty | Response: Response
func flushOldLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := database.FlushOldLogs(); err != nil {
		log.Error("Exception in flushOldLogs: database failure: ", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "failed to flush logs: database failure"})
		return
	}
	go event.Info("Flushed Old Logs", "Logs which are older than 30 days were deleted.")
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully flushed logs older than 30 days"})
}

// Triggers deletion of ALL internal server logs, admin authentication required
// Request: empty | Response: Response
func flushAllLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := database.FlushAllLogs(); err != nil {
		log.Error("Exception in flushOldLogs: database failure: ", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
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
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "failed to get logs: database failure"})
		return
	}
	json.NewEncoder(w).Encode(logs)
}

// Adds a given permission to a given user, admin authentication required
// If the permission is invalid, a `422` is returned
// Request: `{"username": "", "permission": ""}` | Response: Response
func addUserPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserPermissionRequest
	if err := decoder.Decode(&request); err != nil {
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
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Response{Success: true, Message: "user is already in possession of this permission"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	go event.Info("Added User Permission", fmt.Sprintf("Added permission %s to user %s.", request.Permission, request.Username))
	json.NewEncoder(w).Encode(Response{Success: true, Message: fmt.Sprintf("successfully added permission `%s` to user `%s`", request.Permission, request.Username)})
}

// Todo: unit tests for some of the subfunction
// Removes a given permission from a user, admin authentication required
// Request: `{"username": "x", "permission": "y"}` | Response: Response
func removeUserPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserPermissionRequest
	if err := decoder.Decode(&request); err != nil {
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
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to remove permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Response{Success: true, Message: "user does not have this permission"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	go event.Info("Removed User Permission", fmt.Sprintf("Removed permission %s from user %s.", request.Permission, request.Username))
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully removed permission from user"})
}

// TODO: add all request body docs
// Add a switchPermission to a given user, admin authentication required
// Request: `{"username": "x", "switch": "y"}` | Response: Response
func addSwitchPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserSwitchPermissionRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	switchExists, err := database.DoesSwitchExist(request.Switch)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add switch permission", Error: "database failure"})
		return
	}
	if !switchExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not add switch permission to user", Error: "invalid switch permission type: not found"})
		return
	}
	modified, err := database.AddUserSwitchPermission(request.Username, request.Switch)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add switch permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Response{Success: true, Message: "user is already in possession of this switch permission"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	go event.Info("Added Switch Permission", fmt.Sprintf("Added switch permission %s to user %s.", request.Switch, request.Username))
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully added switch permission to user"})
}

// TODO: split into either a submodlue or separate files, should be in a subfolder in server/api or server/routes/api and /server/routes/ui
// Removes a given switch permission from a given user, admin authentication required
// Request: `{"username": "x", "switch": "y"}` | Response: Response
func removeSwitchPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserSwitchPermissionRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	switchExists, err := database.DoesSwitchExist(request.Switch)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to remove switch permission", Error: "database failure"})
		return
	}
	if !switchExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not remove switch permission from user", Error: "invalid switch permission type: not found"})
		return
	}
	modified, err := database.RemoveUserSwitchPermission(request.Username, request.Switch)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to remove switch permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Response{Success: true, Message: "user does not have this switch permission"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	go event.Info("Removed Switch Permission", fmt.Sprintf("Removed switch permission %s from user %s.", request.Switch, request.Username))
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully removed switch permission from user"})
}

// Creates a new user and gives him a provided password
// Request: `{"username": "x", "password": "y"}`
func addUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request AddUserRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	userAlreadyExists, err := database.DoesUserExist(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add user", Error: "database failure"})
		return
	}
	if userAlreadyExists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add user", Error: "user already exists"})
		return
	}
	if err = database.AddUser(database.FullUser{Username: request.Username, Password: request.Password}); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add user", Error: "database failure"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully created new user"})
}

// Deletes a user given a valid username
// This also needs to delete any data that depends on this user in terms of a foreign key
func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request RemoveUserRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	userDoesExist, err := database.DoesUserExist(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to remove user", Error: "database failure"})
		return
	}
	if !userDoesExist {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to delete user", Error: "no user exists with given username"})
		return
	}
	if err := user.DeleteUser(request.Username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to remove user", Error: "backend failure"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully deleted user"})
}

// Runs a healthcheck of most systems on which the appplication relies on, will be used by e.g `Uptime Kuma`, no authentication required
// TODO: also check the hardware nodes
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := database.CheckDatabase(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.Error("Healthcheck failed: ", err.Error())
		json.NewEncoder(w).Encode(Response{Success: false, Message: "healthcheck failed: database downtime", Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Reading system debug information, admin authentication required
// Todo: read raspberry pi information
func debugInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.SysInfo())
}

// Returns a uin16 that indicates the number of notifications the current user has, no authentication required
func getNotificationCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	notificationCount, err := database.GetUserNotificationCount(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed get notification count", Error: "database failure"})
		return
	}
	json.NewEncoder(w).Encode(NotificationCountResponse{NotificationCount: notificationCount})
}

// Returns a list containing notifications of the current user
func getNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	notifications, err := database.GetUserNotifications(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed get notification count", Error: "database failure"})
		return
	}
	json.NewEncoder(w).Encode(notifications)
}

// TEST IMAGE FETCHING MODULE
func TestImageProxy(w http.ResponseWriter, r *http.Request) {
	imageData, err := camera.TestReturn()
	if err != nil {
		log.Error("Failed to test proxy: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", http.DetectContentType(imageData))
	if _, err := w.Write(imageData); err != nil {
		log.Error(err.Error())
	}
}

// Returns the user's personal data
func getUserDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	userData, err := database.GetUserByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed get user data", Error: "database failure"})
		return
	}
	json.NewEncoder(w).Encode(userData)
}
