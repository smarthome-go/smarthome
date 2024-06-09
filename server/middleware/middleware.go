package middleware

import (
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/user"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
	Time    string `json:"time"`
}

// Parses the session and returns the currently logged in user
// If no user is logged in but is trying to authenticate with URL-queries,
// this function will call `getUserFromQuery` internally in order to get the username
func GetUserFromCurrentSession(w http.ResponseWriter, r *http.Request) (string, error) {
	session, err := Store.Get(r, "session")
	if err != nil {
		username, validCredentials, err := getUserFromQuery(r)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			log.Error("Could not get session from request: ", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
			return "", err
		}
		if !validCredentials {
			// This should not happen.
			// Either a session or query will be valid (middleware.AuthRequired checks for this requirement).
			log.Error("failed to get username from query")
			return "", errors.New("failed to get username from query")
		}
		return username, nil
	}
	usernameTemp, ok := session.Values["username"]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return "", errors.New("could not obtain username from session")
	}
	username, ok := usernameTemp.(string)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
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
	loginValid, err := user.ValidateCredentials(username, password)
	if err != nil {
		log.Error("Could not use GetUserFromQuery: failed to validate login credentials due to database failure", err.Error())
		return "", false, err
	}
	if loginValid {
		return username, true, nil
	}
	// If the conventional way of authentication failed, check if a authentication token is present
	token := query.Get("token")
	data, found, err := database.GetUserTokenByToken(token)
	if err != nil {
		log.Error("Could not use GetUserFromQuery: failed to validate authentication token due to database failure", err.Error())
		return "", false, err
	}
	if !found {
		return "", false, nil
	}
	return data.User, true, nil
}
