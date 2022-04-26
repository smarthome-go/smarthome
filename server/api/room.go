package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/exp/utf8string"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

type RoomRequest struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Returns list of rooms which contain switches that the user is allowed to use
func GetUserRoomsWithSwitches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	rooms, err := database.ListPersonalRooms(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "could not list personal rooms", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(rooms); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to get user rooms", Error: "could not encode content"})
	}
}

func AddRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request RoomRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	if strings.Contains(request.Id, " ") || !utf8string.NewString(request.Id).IsASCII() {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "id should only include ASCII characterst and must not have whitespaces"})
		return
	}
	_, alreadyExists, err := database.GetRoomDataById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create new room: could not check for conflicts", Error: "database failure"})
		return
	}
	if alreadyExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create new room", Error: "a room with the same room-id already exists"})
		return
	}
	if err := database.CreateRoom(database.RoomData(request)); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create new room", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully created new room"})
}

// Modifies the room's name and description
func ModifyRoomData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request RoomRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	room, found, err := database.GetRoomDataById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify room", Error: "database failure"})
		return
	}
	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify room", Error: "invalid room-id"})
		return
	}
	if room.Name == request.Name && room.Description == request.Description {
		Res(w, Response{Success: true, Message: "data unchanged"})
		return
	}
	if err := database.ModifyRoomData(request.Id, request.Name, request.Description); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify room", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully modified room"})
}

// Deletes a room and all its dependencies
func DeleteRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete room", Error: "invalid room-id"})
		return
	}
	_, exists, err := database.GetRoomDataById(id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete room", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete room", Error: "invalid room-id"})
		return
	}
	if err := database.DeleteRoom(id); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete room", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted room"})
}
