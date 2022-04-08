package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

type UserPermissionRequest struct {
	Username   string `json:"username"`
	Permission string `json:"permission"`
}
type UserSwitchPermissionRequest struct {
	Username string `json:"username"`
	Switch   string `json:"switch"`
}

// Returns a list of strings which resemble permissions of the currently logged in user, authentication required
// Request: empty | Response: `["a", "b", "c"]`
func GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	permissions, err := database.GetUserPermissions(username)
	if err != nil {
		log.Error("Exception in getUserPermissions: database failure: ", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(permissions); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "could not get user permissions", Error: "failed to encode response"})
	}
}

// Adds a given permission to a given user, admin authentication required
// If the permission is invalid, a `422` is returned
// Request: `{"username": "", "permission": ""}` | Response: Response
func AddUserPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserPermissionRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	validPermission := database.DoesPermissionExist(request.Permission)
	if !validPermission {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not add permission to user", Error: "invalid permission type"})
		return
	}
	_, userExists, err := database.GetUserByUsername(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add permission", Error: "database failure"})
		return
	}
	if !userExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not add permission to user", Error: "invalid user"})
		return
	}
	modified, err := database.AddUserPermission(request.Username, database.PermissionType(request.Permission))
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusConflict)
		Res(w, Response{Success: true, Message: "user is already in possession of this permission"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	go event.Info("Added User Permission", fmt.Sprintf("Added permission %s to user %s.", request.Permission, request.Username))
	Res(w, Response{Success: true, Message: fmt.Sprintf("successfully added permission `%s` to user `%s`", request.Permission, request.Username)})
}

// Todo: unit tests for some of the subfunction
// Removes a given permission from a user, admin authentication required
// Request: `{"username": "x", "permission": "y"}` | Response: Response
func RemoveUserPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserPermissionRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	validPermission := database.DoesPermissionExist(request.Permission)
	if !validPermission {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not remove permission from user", Error: "invalid permission type"})
		return
	}
	_, userExists, err := database.GetUserByUsername(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to remove permission", Error: "database failure"})
		return
	}
	if !userExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not remove permission from user", Error: "invalid user"})
		return
	}
	modified, err := database.RemoveUserPermission(request.Username, database.PermissionType(request.Permission))
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to remove permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusConflict)
		Res(w, Response{Success: true, Message: "user does not have this permission"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	go event.Info("Removed User Permission", fmt.Sprintf("Removed permission %s from user %s.", request.Permission, request.Username))
	Res(w, Response{Success: true, Message: "successfully removed permission from user"})
}

// Add a switchPermission to a given user, admin authentication required
// Request: `{"username": "x", "switch": "y"}` | Response: Response
func AddSwitchPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserSwitchPermissionRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	switchExists, err := database.DoesSwitchExist(request.Switch)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add switch permission", Error: "database failure"})
		return
	}
	if !switchExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not add switch permission to user", Error: "invalid switch permission type: not found"})
		return
	}
	_, userExists, err := database.GetUserByUsername(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add switch permission", Error: "database failure"})
		return
	}
	if !userExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not add switch permission from user", Error: "invalid user"})
		return
	}
	modified, err := database.AddUserSwitchPermission(request.Username, request.Switch)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add switch permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusConflict)
		Res(w, Response{Success: true, Message: "user is already in possession of this switch permission"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	go event.Info("Added Switch Permission", fmt.Sprintf("Added switch permission %s to user %s.", request.Switch, request.Username))
	Res(w, Response{Success: true, Message: "successfully added switch permission to user"})
}

// Removes a given switch permission from a given user, admin authentication required
// Request: `{"username": "x", "switch": "y"}` | Response: Response
func RemoveSwitchPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserSwitchPermissionRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	switchExists, err := database.DoesSwitchExist(request.Switch)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to remove switch permission", Error: "database failure"})
		return
	}
	if !switchExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not remove switch permission from user", Error: "invalid switch permission type: not found"})
		return
	}
	_, userExists, err := database.GetUserByUsername(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to remove switch permission", Error: "database failure"})
		return
	}
	if !userExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not remove switch permission from user", Error: "invalid user"})
		return
	}
	modified, err := database.RemoveUserSwitchPermission(request.Username, request.Switch)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to remove switch permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusConflict)
		Res(w, Response{Success: true, Message: "user does not have this switch permission"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	go event.Info("Removed Switch Permission", fmt.Sprintf("Removed switch permission %s from user %s.", request.Switch, request.Username))
	Res(w, Response{Success: true, Message: "successfully removed switch permission from user"})
}
