package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/user"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type UserPermissionRequest struct {
	Username   string `json:"username"`
	Permission string `json:"permission"`
}
type UserSwitchPermissionRequest struct {
	Username string `json:"username"`
	Switch   string `json:"switch"`
}

// Returns a list of strings which represent permissions of the currently logged in user, admin authentication required
// Request: empty | Response: `["a", "b", "c"]`
func GetCurrentUserPermissions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	permissions, err := database.GetUserPermissions(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(permissions); err != nil {
		log.Error("Failed to encode response: ", err.Error())
		Res(w, Response{Success: false, Message: "could not get user permissions", Error: "failed to encode response"})
	}
}

// Returns a list of strings which represent the permissions of an arbitrary user, admin authentication required
func GetForeignUserPermissions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	username, ok := vars["username"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to get user permissions", Error: "no username provided"})
		return
	}
	_, exists, err := database.GetUserByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to get user permissions", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		Res(w, Response{Success: false, Message: "failed to get user permissions", Error: "invalid username"})
		return
	}
	permissions, err := database.GetUserPermissions(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(permissions); err != nil {
		log.Error("Failed to encode response: ", err.Error())
		Res(w, Response{Success: false, Message: "could not get user permissions", Error: "failed to encode response"})
	}
}

// Returns a list of strings which represent the switch permissions of an arbitrary user, admin authentication required
func GetForeignUserSwitchPermissions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	username, ok := vars["username"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "no username provided", Error: "no username provided"})
		return
	}
	_, exists, err := database.GetUserByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to get user permissions", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		Res(w, Response{Success: false, Message: "failed to get user permissions", Error: "invalid username"})
		return
	}
	permissions, err := database.GetUserSwitchPermissions(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(permissions); err != nil {
		log.Error("Failed to encode response: ", err.Error())
		Res(w, Response{Success: false, Message: "could not get user switch permissions", Error: "failed to encode response"})
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
		Res(w, Response{Success: false, Message: "failed to add permission", Error: "invalid permission type"})
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
		Res(w, Response{Success: false, Message: "failed to add permission", Error: "invalid user"})
		return
	}
	modified, err := user.AddPermission(request.Username, database.PermissionType(request.Permission))
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add permission", Error: "user is already in possession of this permission"})
		return
	}
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
	// If the `manageUsers` permission should be removed, assure that it will not break the system
	// Otherwise, a complete reset would be required because no new users could be created or permission edited
	// If other users with this permission exist, it is safe to remove
	if request.Permission == string(database.PermissionManageUsers) || request.Permission == string(database.PermissionWildCard) {
		isAlone, err := user.IsStandaloneUserAdmin(request.Username)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			Res(w, Response{Success: false, Message: "failed to remove permission", Error: "database failure"})
			return
		}
		if isAlone {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Res(w, Response{Success: false, Message: "failed to remove permission", Error: "permission is not safe to remove: removal would lock system because the requested user is the only user-administrator"})
			return
		}
	}
	modified, err := user.RemovePermission(request.Username, database.PermissionType(request.Permission))
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to remove permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "did remove permission", Error: "user does not have this permission"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully removed permission from user"})
}

// Returns the list of all available permissions
func ListPermissions(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(database.Permissions); err != nil {
		Res(w, Response{Success: false, Message: "failed to list permissions", Error: "could not encode content"})
		return
	}
}
