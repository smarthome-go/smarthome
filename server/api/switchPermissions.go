package api

import (
	"encoding/json"
	"net/http"

	"github.com/smarthome-go/smarthome/core/database"
)

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
	_, switchExists, err := database.GetSwitchById(request.Switch)
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
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add switch permission", Error: "user is already in possession of this switch permission"})
		return
	}
	w.WriteHeader(http.StatusCreated)
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
	_, switchExists, err := database.GetSwitchById(request.Switch)
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
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to remove switch permission", Error: "user does not have this switch permission"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	Res(w, Response{Success: true, Message: "successfully removed switch permission from user"})
}
