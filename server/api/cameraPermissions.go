package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/smarthome-go/smarthome/core/database"
)

// Returns the camera-permissions of an arbitrary user as a slice of string
// Returns a list of cameras to which an arbitrary user has access to, admin authentication required
func GetForeignUserCameraPermission(w http.ResponseWriter, r *http.Request) {
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
		Res(w, Response{Success: false, Message: "failed to get user camera-permissions", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		Res(w, Response{Success: false, Message: "failed to get user camera-permissions", Error: "invalid username"})
		return
	}
	permissions, err := database.GetUserCameraPermissions(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(permissions); err != nil {
		log.Error("Failed to encode response: ", err.Error())
		Res(w, Response{Success: false, Message: "could not get user camera-permissions", Error: "failed to encode response"})
	}
}

// Add a camera-permission to a given user, admin authentication required
// Request: `{"username": "x", "cameraId": "y"}` | Response: Response
func AddCameraPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request CameraPermissionRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, camExists, err := database.GetCameraById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add camera permission", Error: "database failure"})
		return
	}
	if !camExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not add camera permission to user", Error: "invalid camera permission type: not found"})
		return
	}
	_, userExists, err := database.GetUserByUsername(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add camera permission", Error: "database failure"})
		return
	}
	if !userExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not add camera permission from user", Error: "invalid user"})
		return
	}
	modified, err := database.AddUserCameraPermission(request.Username, request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add camera permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add camera permission", Error: "user is already in possession of this camera permission"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	Res(w, Response{Success: true, Message: "successfully added camera permission to user"})
}

// Removes a given camera permission from a given user, admin authentication required
// Request: `{"username": "x", "id": "y"}` | Response: Response
func RemoveCameraPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request CameraPermissionRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, camExists, err := database.GetCameraById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to remove camera permission", Error: "database failure"})
		return
	}
	if !camExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not remove camera permission from user", Error: "invalid camera permission type: not found"})
		return
	}
	_, userExists, err := database.GetUserByUsername(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to remove camera permission", Error: "database failure"})
		return
	}
	if !userExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not remove camera permission from user", Error: "invalid user"})
		return
	}
	modified, err := database.RemoveUserCameraPermission(request.Username, request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to remove camera permission", Error: "database failure"})
		return
	}
	if !modified {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to remove camera permission", Error: "user does not have this camera permission"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	Res(w, Response{Success: true, Message: "successfully removed camera permission from user"})
}
