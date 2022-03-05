package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/core/user"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

// Accepts a json request like `"username": "user",  "password":"password"`
// If the credentials are valid, a new session is created and the user is saved, otherwise a 401 is returned
func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var loginRequest LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&loginRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "login failed", Error: "malformed request"})
	}
	loginValid, err := user.ValidateLogin(loginRequest.Username, loginRequest.Password)
	if err != nil {
		log.Error("User failed to login: database failure.")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "login failed", Error: "could not validate login: internal error: database failure"})
		return
	}
	if loginValid {
		session, _ := middleware.Store.Get(r, "session")
		session.Values["valid"] = true
		session.Values["username"] = loginRequest.Username
		session.Save(r, w)
		w.WriteHeader(http.StatusNoContent)
		log.Debug(fmt.Sprintf("User %s logged in successfully", loginRequest.Username))
		go event.Info("Successful login", fmt.Sprintf("%s logged in", loginRequest.Username))
		return
	}
	log.Debug("Login failed: invalid credentials")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(Response{Success: false, Message: "login failed", Error: "invalid credentials"})
	go event.Warn("Failed Login", fmt.Sprintf("Someone is tying to login to the account %s", loginRequest.Username))
}

// TODO: implement following function
func logoutGetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "not implemented")
}
