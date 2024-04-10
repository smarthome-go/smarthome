package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smarthome-go/smarthome/core/automation"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/homescript/types"
	"github.com/smarthome-go/smarthome/core/user"
	"github.com/smarthome-go/smarthome/server/api"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type userLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenLoginRequest struct {
	Token string `json:"token"`
}

type TokenLoginResponse struct {
	Username   string `json:"username"`
	TokenLabel string `json:"tokenLabel"`
}

// Accepts a json request like `{"token": "foo"}`
// If the credentials are valid, a new session is created and the user is saved, otherwise a 401 is returned
func tokenLoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var request tokenLoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		api.Res(w, api.Response{Success: false, Message: "login failed", Error: "malformed request"})
		return
	}
	// Check the token against the database
	tokenData, tokenValid, err := database.GetUserTokenByToken(request.Token)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		api.Res(w, api.Response{Success: false, Message: "login failed", Error: "could not validate login: internal error: database failure"})
		return
	}
	if !tokenValid {
		w.WriteHeader(http.StatusUnauthorized)
		api.Res(w, api.Response{Success: false, Message: "login failed", Error: "invalid authentication token"})
		event.Warn("Failed Login Attempt", fmt.Sprintf("Failed login attempt using token `%s`", request.Token))
		return
	}
	// Once the token is validated, save the user's session
	session, _ := middleware.Store.Get(r, "session")
	session.Values["valid"] = true
	session.Values["username"] = tokenData.User
	if err := session.Save(r, w); err != nil {
		log.Error("Failed to save session: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		api.Res(w, api.Response{Success: false, Message: "failed to authenticate", Error: "could not save session after successful authentication"})
	}
	// Send back the username for clients which rely on this information (SDK)
	if err := json.NewEncoder(w).Encode(TokenLoginResponse{
		Username:   tokenData.User,
		TokenLabel: tokenData.Data.Label,
	}); err != nil {
		log.Error("Failed to send response to client: ", err.Error())
		return
	}
	log.Debug(fmt.Sprintf("User `%s` logged in successfully using an access-token", tokenData.User))
	go event.Info("Successful login", fmt.Sprintf("User `%s` logged in using an access-token", tokenData.User))

	// Run any login hooks
	go automation.Manager.RunAllAutomationsWithTrigger(tokenData.User, database.TriggerOnLogin, types.AutomationContext{})
}

// Accepts a json request like `{"username": "user", "password":"password"}`
// If the credentials are valid, a new session is created and the user is saved, otherwise a 401 is returned
func userLoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var loginRequest userLoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&loginRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		api.Res(w, api.Response{Success: false, Message: "login failed", Error: "malformed request"})
		return
	}
	loginValid, err := user.ValidateCredentials(loginRequest.Username, loginRequest.Password)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		api.Res(w, api.Response{Success: false, Message: "login failed", Error: "could not validate login: internal error: database failure"})
		return
	}
	if !loginValid {
		w.WriteHeader(http.StatusUnauthorized)
		api.Res(w, api.Response{Success: false, Message: "login failed", Error: "invalid credentials"})
		event.Warn("Failed Login Attempt", fmt.Sprintf("Failed login attempt of user account `%s`", loginRequest.Username))
		return
	}
	// Once the credentials are validated, save the user in a session
	session, _ := middleware.Store.Get(r, "session")
	session.Values["valid"] = true
	session.Values["username"] = loginRequest.Username
	if err := session.Save(r, w); err != nil {
		log.Error("Failed to save session: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		api.Res(w, api.Response{Success: false, Message: "failed to authenticate", Error: "could not save session after successful authentication"})
	}
	w.WriteHeader(http.StatusNoContent)
	log.Debug(fmt.Sprintf("User `%s` logged in successfully", loginRequest.Username))
	go event.Info("Successful login", fmt.Sprintf("User %s logged in", loginRequest.Username))

	// Run any login hooks
	go automation.Manager.RunAllAutomationsWithTrigger(loginRequest.Username, database.TriggerOnLogin, types.AutomationContext{})
}

// invalidates the user session and then redirects back to the login page
func logoutGetHandler(w http.ResponseWriter, r *http.Request) {
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		// No user is logged in
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	session, err := middleware.Store.Get(r, "session")
	if err != nil {
		// No user is logged in
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	session.Values["valid"] = false
	session.Values["username"] = ""
	if err := session.Save(r, w); err != nil {
		log.Error("Failed to save session: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		api.Res(w, api.Response{Success: false, Message: "failed to authenticate", Error: "could not save session after successful authentication"})
	}
	log.Trace("A user logged out")
	http.Redirect(w, r, "/login", http.StatusFound)

	// Run any logout hooks
	go automation.Manager.RunAllAutomationsWithTrigger(username, database.TriggerOnLogout, types.AutomationContext{})
}
