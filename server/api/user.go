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

// Creates a new user and gives him a provided password
// Request: `{"username": "x", "password": "y"}`
func AddUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request AddUserRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	userAlreadyExists, err := database.DoesUserExist(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add user", Error: "database failure"})
		return
	}
	if userAlreadyExists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add user", Error: "user already exists"})
		return
	}
	if err = database.AddUser(database.FullUser{Username: request.Username, Password: request.Password}); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add user", Error: "database failure"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully created new user"})
}

// Deletes a user given a valid username
// This also needs to delete any data that depends on this user in terms of a foreign key
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request RemoveUserRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	userDoesExist, err := database.DoesUserExist(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to remove user", Error: "database failure"})
		return
	}
	if !userDoesExist {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to delete user", Error: "no user exists with given username"})
		return
	}
	if err := user.DeleteUser(request.Username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to remove user", Error: "backend failure"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully deleted user"})
}

// Returns the user's personal data
func GetUserDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	userData, err := database.GetUserByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed get user data", Error: "database failure"})
		return
	}
	json.NewEncoder(w).Encode(userData)
}
