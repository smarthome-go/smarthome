package api

import (
	"encoding/json"
	"net/http"

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
		w.WriteHeader(http.StatusConflict)
		Res(w, Response{Success: false, Message: "failed to add user", Error: "user already exists"})
		return
	}
	if err = database.AddUser(database.FullUser{Username: request.Username, Password: request.Password}); err != nil {
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
	userData, found, err := database.GetUserByUsername(username)
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
func SetColorTheme(w http.ResponseWriter, r *http.Request) {
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
