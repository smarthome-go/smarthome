package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/user"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

type ResponseStruct struct {
	Success   bool
	ErrorCode int
	Title     string
	Message   string
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
			w.WriteHeader(http.StatusInternalServerError)
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
			json.NewEncoder(w).Encode(ResponseStruct{false, 401, "access denied, please authenticate", "authentication required"})
			return
		}
		loginValid, err := user.ValidateLogin(username, password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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
			json.NewEncoder(w).Encode(ResponseStruct{false, 401, "access denied, please authenticate", "invalid credentials"})
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
	log.Debug("Using `getUserFromQuery`, this is likely a new session")
	query := r.URL.Query()
	username := query.Get("username")
	password := query.Get("password")
	loginValid, err := user.ValidateLogin(username, password)
	if err != nil {
		log.Error("Could not use GetUserFromQuery: failed to login", err.Error())
		return "", false, err
	}
	if loginValid {
		return username, true, nil
	} else {
		return "", false, nil
	}
}
