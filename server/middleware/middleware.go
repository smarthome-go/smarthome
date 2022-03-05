package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/user"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func AuthRequired(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "session")
		value, ok := session.Values["valid"]
		valid, okParse := value.(bool)
		query := r.URL.Query()
		username := query.Get("username")
		password := query.Get("password")

		if ok && okParse && valid {
			handler.ServeHTTP(w, r)
			return
		}
		loginValid, err := user.ValidateLogin(username, password)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		if loginValid {
			// Saves session
			session, _ := Store.Get(r, "session")
			session.Values["valid"] = true
			session.Values["username"] = username
			session.Save(r, w)
			handler.ServeHTTP(w, r)
			return
		}
		log.Trace(fmt.Sprintf("Invalid Session, redirecting %s to /login", r.URL.Path))
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

//
func ApiAuthRequired(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "session")
		value, ok := session.Values["valid"]
		valid, okParse := value.(bool)

		query := r.URL.Query()
		username := query.Get("username")
		password := query.Get("password")

		// The last part checks if the user has the intention to authenticate again
		// This could be the case if another user wants to log in from the same connection
		if ok && okParse && valid && username == "" {
			log.Trace(fmt.Sprintf("Valid Session, serving %s", r.URL.Path))
			handler.ServeHTTP(w, r)
			return
		}
		if username == "" && password == "" {
			log.Trace("user session invalid, no query present")
			log.Trace("Invalid Session, not serving", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{false, "access denied, please authenticate", "authentication required"})
			return
		}
		loginValid, err := user.ValidateLogin(username, password)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		if loginValid {
			session, _ := Store.Get(r, "session")
			session.Values["valid"] = true
			session.Values["username"] = username
			session.Save(r, w)
			log.Trace(fmt.Sprintf("valid query: Session Saved. Serving %s", r.URL.Path))
			handler.ServeHTTP(w, r)
			return
		} else {
			log.Trace("bad credentials, invalid Session: not serving", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{false, "access denied, please authenticate", "invalid credentials"})
			return
		}
	}
}

// Parses the session and returns the currently logged in user
// If no user is logged in but is trying to authenticate with URL-queries,
// this function will call `getUserFromQuery` internally in order to get the username
func GetUserFromCurrentSession(r *http.Request) (string, error) {
	session, err := Store.Get(r, "session")
	if err != nil {
		username, validCredentials, err := getUserFromQuery(r)
		if err != nil {
			log.Error("Could not get session from request: ", err.Error())
			return "", err
		}
		if !validCredentials {
			// this should not happen
			// Either a session or query will be valid (middleware.AuthRequired checks for this requirement)
			log.Error("failed to get username from query")
			return "", errors.New("failed to get username from query")
		}
		return username, nil

	}
	usernameTemp, ok := session.Values["username"]
	if !ok {
		return "", errors.New("could not obtain username from session")
	}
	username, ok := usernameTemp.(string)
	if !ok {
		return "", errors.New("could not obtain username from session")
	}
	return username, nil
}

// Will be used by GetUserFromCurrentSession if GetUserFromCurrentSession fails
// Returns a string for the username,
// a boolean that indicates if the credentials are valid and an error for database failure
func getUserFromQuery(r *http.Request) (string, bool, error) {
	query := r.URL.Query()
	username := query.Get("username")
	password := query.Get("password")
	loginValid, err := user.ValidateLogin(username, password)
	if err != nil {
		log.Error("Could not use GetUserFromQuery: failed validate login credentials due to database failure", err.Error())
		return "", false, err
	}
	if loginValid {
		return username, true, nil
	} else {
		return "", false, nil
	}
}

//
func Permission(handler http.HandlerFunc, permissionToCheck string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := GetUserFromCurrentSession(r)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			log.Error("failed to get username from query")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{false, "access denied, invalid session", "clear your browser's cookies"})
			return
		}
		hasPermission, err := database.UserHasPermission(username, permissionToCheck)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(Response{Success: false, Message: "database error", Error: "failed to check permission to access this ressource"})
			return
		}
		if !hasPermission {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(Response{Success: false, Message: "permission denied", Error: "missing permission to access this ressource, contact your administrator"})
			return
		}
		handler.ServeHTTP(w, r)
	}
}
