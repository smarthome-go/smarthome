package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"

	"golang.org/x/exp/utf8string"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type AddSwitchRequest struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	RoomId string `json:"roomId"`
	Watts  uint16 `json:"watts"`
	//TargetNode     *string `json:"targetNode"`
	DriverVendorId string `json:"driverVendorId"`
	DriverModelId  string `json:"driverModelId"`
}

type ModifySwitchRequest struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Watts uint16 `json:"watts"`
	//TargetNode *string `json:"targetNode"`

	DriverVendorId string `json:"driverVendorId"`
	DriverModelId  string `json:"driverModelId"`
}

type DeleteSwitchRequest struct {
	Id string `json:"id"`
}

// Returns a list of available switches as JSON to the user, no authentication required
func GetAllDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switches, err := database.ListSwitches()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(switches); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to get switches", Error: "could not encode content"})
		return
	}
}

// Only returns switches which the user has access to, authentication required
func GetUserDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	switches, err := database.ListUserSwitches(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(switches); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to get personal switches", Error: "could not encode content"})
	}
}

// Creates a switch in the database
func CreateDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request AddSwitchRequest
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
	if utf8.RuneCountInString(request.Id) > 20 || utf8.RuneCountInString(r.RequestURI) > 30 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "maximum lengths for id and name are 20 and 30"})
		return
	}
	// Validate that no conflicts are present
	_, alreadyExists, err := database.GetSwitchById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create switch", Error: "database failure"})
		return
	}
	if alreadyExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create switch", Error: "id already exists"})
		return
	}
	// Validate that the room exists
	_, roomExists, err := database.GetRoomDataById(request.RoomId)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create switch", Error: "database failure"})
		return
	}
	if !roomExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create switch", Error: "invalid room id"})
		return
	}

	// if request.TargetNode != nil {
	// 	_, found, err := database.GetHardwareNodeByUrl(*request.TargetNode)
	// 	if err != nil {
	// 		w.WriteHeader(http.StatusServiceUnavailable)
	// 		Res(w, Response{Success: false, Message: "failed to modify switch", Error: "database failure"})
	// 		return
	// 	}
	// 	if !found {
	// 		w.WriteHeader(http.StatusUnprocessableEntity)
	// 		Res(w, Response{Success: false, Message: "failed to modify switch", Error: "no hardware node with URL exists"})
	// 		return
	// 	}
	// }

	// TODO: Validate drivers

	if err := database.CreateDevice(
		request.Id,
		request.Name,
		request.RoomId,
		request.Watts,
		request.TargetNode,
	); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create switch", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully created switch"})
}

func ModifyDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request ModifySwitchRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, found, err := database.GetSwitchById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify switch", Error: "database failure"})
		return
	}
	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify switch", Error: "no switch with id exists"})
		return
	}

	if request.TargetNode != nil {
		_, found, err := database.GetHardwareNodeByUrl(*request.TargetNode)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			Res(w, Response{Success: false, Message: "failed to modify switch", Error: "database failure"})
			return
		}
		if !found {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Res(w, Response{Success: false, Message: "failed to modify switch", Error: "no hardware node with URL exists"})
			return
		}
	}

	// Validate length
	if len(request.Name) > 30 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "maximum name length of 30 chars. was exceeded"})
		return
	}

	if err := database.ModifySwitch(request.Id, request.Name, request.Watts, request.TargetNode); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify switch", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully modified switch"})
}

func DeleteDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request DeleteSwitchRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, found, err := database.GetSwitchById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete switch", Error: "database failure"})
		return
	}
	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete switch", Error: "no switch with id exists"})
		return
	}
	if err := database.DeleteSwitch(request.Id); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete switch", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted switch"})
}
