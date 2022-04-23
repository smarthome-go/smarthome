package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/user"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

type AddUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RemoveUserRequest struct {
	Username string `json:"username"`
}

type SetColorThemeRequest struct {
	DarkTheme bool `json:"darkTheme"`
}

type SetForeignUserColorThemeRequest struct {
	Username  string `json:"username"`
	DarkTheme bool   `json:"darkTheme"`
}

type ModifyUserMetadataRequest struct {
	Forename          string `json:"forename"`
	Surname           string `json:"surname"`
	PrimaryColorDark  string `json:"primaryColorDark"`
	PrimaryColorLight string `json:"primaryColorLight"`
}

type ModifyForeignUserMetadataRequest struct {
	Username string                    `json:"username"`
	Data     ModifyUserMetadataRequest `json:"data"`
}

// Creates a new user and gives him a provided password
// Request: `{"username": "x", "password": "y"}`, admin auth required
func AddUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request AddUserRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, userAlreadyExists, err := database.GetUserByUsername(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add user", Error: "database failure"})
		return
	}
	if userAlreadyExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add user", Error: "user already exists"})
		return
	}
	if len(request.Username) == 0 || strings.Contains(request.Username, " ") {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add user", Error: "bad format: username should not be blank and may not contain whitespaces"})
		return
	}
	if err = database.AddUser(
		database.FullUser{
			Username:          strings.ToLower(request.Username),
			Password:          request.Password,
			Forename:          "Forename",
			Surname:           "Surname",
			PrimaryColorDark:  "#88FF70",
			PrimaryColorLight: "#2E7D32",
		}); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add user", Error: "database failure"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	Res(w, Response{Success: true, Message: "successfully created new user"})
}

// Deletes a user given a valid username
// This also needs to delete any data that depends on this user in terms of a foreign key
// Admin auth required
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request RemoveUserRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, userDoesExist, err := database.GetUserByUsername(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to remove user", Error: "database failure"})
		return
	}
	if !userDoesExist {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete user", Error: "no user exists with given username"})
		return
	}
	isAlone, err := user.IsStandaloneUserAdmin(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to remove permission", Error: "database failure"})
		return
	}
	if isAlone {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to remove permission", Error: "user is the only user with permission to create other users"})
		return
	}
	if err := user.DeleteUser(request.Username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to remove user", Error: "backend failure"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	Res(w, Response{Success: true, Message: "successfully deleted user"})
}

// Returns the user's personal data, auth required
func GetUserDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	userData, found, err := database.GetUserDetails(username)
	if err != nil || !found {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to get user data", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(userData); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to --", Error: "failed to encode response"})
	}
}

// Returns a list of users and their metadata, admin auth required
func ListUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users, err := database.ListUsers()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to list users", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to --", Error: "failed to encode response"})
	}
}

// Allows the user to change whether they want to use the light or dark theme
func SetCurrentUserColorTheme(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request SetColorThemeRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	if err := database.SetUserDarkThemeEnabled(username, request.DarkTheme); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to update color theme", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully updated color theme"})
}

func SetUserColorTheme(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request SetForeignUserColorThemeRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	user, found, err := database.GetUserByUsername(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to update color theme", Error: "database failure"})
		return
	}
	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to update color theme", Error: "not found"})
		return
	}
	if user.DarkTheme == request.DarkTheme {
		w.WriteHeader(http.StatusOK)
		Res(w, Response{Success: true, Message: "theme unchanged"})
		return
	}
	if err := database.SetUserDarkThemeEnabled(request.Username, request.DarkTheme); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to update color theme", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully updated color theme"})
}

// Modifies the metadata of the current user
func ModifyCurrentUserMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request ModifyUserMetadataRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	if len(request.PrimaryColorDark) != 7 || len(request.PrimaryColorLight) != 7 || request.PrimaryColorDark[0] != '#' || request.PrimaryColorLight[0] != '#' {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify user metadata", Error: "invalid color format: a valid color would be `#ffffff`"})
		return
	}
	if len(request.Forename) > 20 || len(request.Surname) > 20 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify user metadata", Error: "data too long for forename or surname. max 20 chars allowed."})
		return
	}
	if err := database.UpdateUserMetadata(username, request.Forename, request.Surname, request.PrimaryColorDark, request.PrimaryColorLight); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		Res(w, Response{Success: false, Message: "failed to modify user metadata", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully updated user metadata"})
}

// Modifies the metadata of a given user
func ModifyUserMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request ModifyForeignUserMetadataRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, exists, err := database.GetUserByUsername(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		Res(w, Response{Success: false, Message: "failed to modify user metadata", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify user metadata", Error: "invalid username"})
		return
	}
	if len(request.Data.PrimaryColorDark) != 7 || len(request.Data.PrimaryColorLight) != 7 || request.Data.PrimaryColorDark[0] != '#' || request.Data.PrimaryColorLight[0] != '#' {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify user metadata", Error: "invalid color format: a valid color would be `#ffffff`"})
		return
	}
	if len(request.Data.Forename) > 20 || len(request.Data.Surname) > 20 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify user metadata", Error: "data too long for forename or surname. max 20 chars allowed."})
		return
	}
	if err := database.UpdateUserMetadata(request.Username, request.Data.Forename, request.Data.Surname, request.Data.PrimaryColorDark, request.Data.PrimaryColorLight); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		Res(w, Response{Success: false, Message: "failed to modify user metadata", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully updated user metadata"})
}