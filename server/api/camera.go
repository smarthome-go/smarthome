package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/server/middleware"
	"golang.org/x/exp/utf8string"
)

type AddCameraRequest struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Url    string `json:"url"`
	RoomId string `json:"roomId"`
}

type ModifyCameraRequest struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type DeleteCameraRequest struct {
	Id string `json:"id"`
}

// Returns a list of available cameras as JSON to the user,
// admin authentication is required because such information is confidential
func GetAllCameras(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cameras, err := database.ListCameras()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to get cameras", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(cameras); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to get cameras", Error: "could not encode contents"})
		return
	}
}

// Only returns cameras to which the user has access to, authentication required
func GetUserCameras(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	cameras, err := database.ListUserCameras(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to get personal cameras", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(cameras); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to get personal cameras", Error: "could not encode contents"})
		return
	}
}

// Creates a new camera with the provided metadata
func CreateCamera(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request AddCameraRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Validate length and encoding
	if strings.Contains(request.Id, " ") || !utf8string.NewString(request.Id).IsASCII() {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "id should only include ASCII characters and must not have whitespaces"})
		return
	}
	if len(request.Id) > 50 || len(request.Name) > 50 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "maximum lengths for id and name are 50 and 50"})
		return
	}
	// Validate that no conflicts are present
	_, alreadyExists, err := database.GetCameraById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create camera", Error: "database failure"})
		return
	}
	if alreadyExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create camera", Error: "id already exists"})
		return
	}
	// Validate that the room exists
	_, roomExists, err := database.GetRoomDataById(request.RoomId)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create camera", Error: "database failure"})
		return
	}
	if !roomExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create camera", Error: "invalid room id"})
		return
	}
	if err := database.CreateCamera(database.Camera(request)); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create camera", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully created camera"})
}

func ModifyCamera(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request ModifyCameraRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	camera, found, err := database.GetCameraById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify camera", Error: "database failure"})
		return
	}
	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify camera", Error: "no camera with id exists"})
		return
	}
	if camera.Name == request.Name && camera.Url == request.Url {
		Res(w, Response{Success: true, Message: "properties unchanged"})
		return
	}
	// Validate length
	if len(request.Name) > 50 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "maximum name length of 50 chars. was exceeded "})
		return
	}
	if err := database.ModifyCamera(request.Id, request.Name, request.Url); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify camera", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully modified camera"})
}

func DeleteCamera(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request DeleteCameraRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, found, err := database.GetCameraById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete camera", Error: "database failure"})
		return
	}
	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete camera", Error: "no camera with id exists"})
		return
	}
	if err := database.DeleteCamera(request.Id); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete camera", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "succesfully deleted camera"})
}
