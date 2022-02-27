package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/middleware"
	"github.com/MikMuellerDev/smarthome/utils"
)

// Accepts a json request like `"username": "user",  "password":"password"`
// If the credentials are valid, a new session is created and the user is saved, otherwise a 401 is returned
// Can return a `500` internal server error if the database fails
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
	log.Debug(fmt.Sprintf("User `%s` is trying to authenticate", loginRequest.Username))
	loginValid, err := utils.ValidateLogin(loginRequest.Username, loginRequest.Password)
	if err != nil {
		log.Error("User failed to login: database failure.")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "login failed", Error: "could not validate login: internal error: database failure"})
		return
	}
	if loginValid {
		session, _ := middleware.Store.Get(r, "session")
		session.Values["valid"] = true
		session.Values["username"] = loginRequest.Username
		session.Save(r, w)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	log.Warn("Login failed: invalid credentials")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(Response{Success: false, Message: "login failed", Error: "invalid credentials"})
}

func logoutGetHandler()
