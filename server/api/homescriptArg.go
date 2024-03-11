package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"unicode/utf8"

	"github.com/gorilla/mux"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/server/middleware"
)

// Is returned when a new Hms argument was created successfully
type AddedHomescriptArgResponse struct {
	NewId    uint     `json:"id"`
	Response Response `json:"response"`
}

// Is used for deleting a Homescript argument
type DeleteHomescriptArgumentRequest struct {
	Id uint `json:"id"`
}

// Is used as a request for editing a Homescript argument
type ModifyHomescriptArgumentRequest struct {
	Id        uint                     `json:"id"`
	ArgKey    string                   `json:"argKey"`
	Prompt    string                   `json:"prompt"`
	MDIcon    string                   `json:"mdIcon"`
	InputType database.HmsArgInputType `json:"inputType"`
	Display   database.HmsArgDisplay   `json:"display"`
}

// Returns all Homescript arguments of all Homescripts which the current user owns
func ListUserHomescriptArgs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	arguments, err := database.ListAllHomescriptArgsOfUser(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to list personal Homescript arguments", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(arguments); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "could not encode response", Error: "could not encode response"})
	}
}

// Returns the arguments of a Homescript given its id
func GetHomescriptArgsByHmsId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	vars := mux.Vars(r)
	homescriptId, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to get arguments of Homescript by its id", Error: "no Homescript id provided"})
		return
	}
	_, exists, err := homescript.HmsManager.GetPersonalScriptById(homescriptId, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to get arguments of Homescript by its id", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to get arguments of Homescript by its id", Error: "referenced Homescript does not exist"})
		return
	}
	arguments, err := database.ListArgsOfHomescript(homescriptId)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to get arguments of Homescript by its id", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(arguments); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "could not encode response", Error: "could not encode response"})
	}
}

// Adds a new argument to a Homescript
func CreateNewHomescriptArg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request database.HomescriptArgData
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Validate the existence of the mentioned Homescript
	_, exists, err := homescript.HmsManager.GetPersonalScriptById(request.HomescriptId, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create new Homescript argument: checks failed", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create new Homescript argument", Error: "referenced Homescript does not exist"})
		return
	}
	// Validate that the length of the key does not exceed 100 chars
	if utf8.RuneCountInString(request.ArgKey) > 100 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create new Homescript argument", Error: "the key must not exceed 100 characters"})
		return
	}
	// Validate that the length of the MdIcon does not exceed 100 chars
	if utf8.RuneCountInString(request.MDIcon) > 100 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add Homescript", Error: fmt.Sprintf("the mdIcon: '%s' must not exceed 100 characters", request.MDIcon)})
		return
	}
	// Creates the argument
	newId, err := database.AddHomescriptArg(request)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create new Homescript argument", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(
		AddedHomescriptArgResponse{
			NewId: newId,
			Response: Response{
				Success: true,
				Message: "successfully created new Homescript argument",
			},
		}); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "could not encode response", Error: "could not encode response"})
	}
}

// Modifies a Homescript argument
func ModifyHomescriptArgument(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request ModifyHomescriptArgumentRequest
	if err := decoder.Decode(&request); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Validate the existence of the requested argument
	_, exists, err := database.GetUserHomescriptArgById(request.Id, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify Homescript argument: checks failed", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify Homescript argument", Error: "Homescript argument does not exist"})
		return
	}
	// Validate that the length of the new key does not exceed 100 chars
	if utf8.RuneCountInString(request.ArgKey) > 100 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify Homescript argument", Error: "the key must not exceed 100 characters"})
		return
	}
	// Validate that the length of the MdIcon does not exceed 100 chars
	if utf8.RuneCountInString(request.MDIcon) > 100 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add Homescript", Error: fmt.Sprintf("the mdIcon: '%s' must not exceed 100 characters", request.MDIcon)})
		return
	}
	// Modifies the argument
	if err := database.ModifyHomescriptArg(request.Id, database.HomescriptArgData{
		ArgKey:    request.ArgKey,
		Prompt:    request.Prompt,
		MDIcon:    request.MDIcon,
		InputType: request.InputType,
		Display:   request.Display,
	}); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify Homescript argument", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully modified Homescript argument", Error: ""})
}

// Deletes a Homescript argument
func DeleteHomescriptArgument(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request DeleteHomescriptArgumentRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Validate the existence of the requested argument
	_, exists, err := database.GetUserHomescriptArgById(request.Id, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete Homescript argument: checks failed", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete Homescript argument", Error: "Homescript argument does not exist"})
		return
	}
	if err := database.DeleteHomescriptArg(request.Id); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete Homescript argument", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted Homescript argument", Error: ""})
}
