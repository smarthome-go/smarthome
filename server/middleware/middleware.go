package middleware

import (
	"errors"
	"fmt"
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

// Checks if a user is already logged in (session)
// If not, it checks for a url query `username=x&password=y` in order to authenticate the user
// If both methods fail, the user is redirected to `/login`
func Auth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := Store.Get(r, "session")
		if err != nil {
			log.Debug("Session exists but could not be decoded: ", err.Error())
		}

		loginValidTemp, loginValidOkTemp := session.Values["valid"]
		loginValid, loginValidOk := loginValidTemp.(bool)

		usernameTemp, usernameTempOk := session.Values["username"]
		usernameSession, usernameSessionOk := usernameTemp.(string)

		query := r.URL.Query()
		username := query.Get("username")
		password := query.Get("password")

		if loginValidOkTemp && loginValidOk && loginValid && username == "" {
			// Check if user exists
			// TODO: implement check in Redis or use other caching
			if usernameTempOk && usernameSessionOk && usernameSession != "" {
				_, exists, err := database.GetUserByUsername(usernameSession)
				if err != nil {
					w.WriteHeader(http.StatusServiceUnavailable)
					Res(w, Response{Success: false, Message: "Could not check user validity", Error: "database failure"})
					return
				}
				if exists { // Do not return an error if the does not exists to allow correction via url queries
					// The session is valid: allow access
					handler.ServeHTTP(w, r)
					return
				}
			}
		}

		// If the provided url query username is blank, redirect to login
		if username == "" {
			log.Trace(fmt.Sprintf("Invalid Session, redirecting %s to /login", r.URL.Path))
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Check potential credentials if the session is invalid
		validCredentials, err := user.ValidateCredentials(username, password)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			Res(w, Response{Success: false, Message: "could not validate credentials", Error: "database failure"})
			return
		}

		if validCredentials {
			session.Values["valid"] = true
			session.Values["username"] = username
			if err := session.Save(r, w); err != nil {
				log.Error("Failed to save session: ", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				Res(w, Response{Success: false, Message: "failed to authenticate", Error: "could not save session after successful authentication"})
			}
			handler.ServeHTTP(w, r)
			return
		}

		log.Trace(fmt.Sprintf("Invalid Session, redirecting %s to /login", r.URL.Path))
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func ApiAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := Store.Get(r, "session")
		if err != nil {
			log.Debug("Session exists but could not be decoded: ", err.Error())
		}

		loginValidTemp, loginValidOkTemp := session.Values["valid"]
		loginValid, loginValidOk := loginValidTemp.(bool)

		usernameTemp, usernameTempOk := session.Values["username"]
		usernameSession, usernameSessionOk := usernameTemp.(string)

		query := r.URL.Query()
		username := query.Get("username")
		password := query.Get("password")

		if loginValidOkTemp && loginValidOk && loginValid && username == "" {
			// The last part (`username == ""`) checks if the user has the intention to authenticate again
			// This could be the case if another user wants to log in from the same connection
			if usernameTempOk && usernameSessionOk && usernameSession != "" {
				_, exists, err := database.GetUserByUsername(usernameSession)
				if err != nil {
					w.WriteHeader(http.StatusServiceUnavailable)
					Res(w, Response{Success: false, Message: "Could not check user validity", Error: "database failure"})
					return
				}
				if exists { // Do not return an error if the does not exists to allow correction via url queries
					// The session is valid: allow access
					log.Trace(fmt.Sprintf("Valid Session, serving %s", r.URL.Path))
					handler.ServeHTTP(w, r)
					return
				}
			}
		}
		if username == "" {
			// Session is invalid and no authentication query is present
			log.Trace("user session invalid, no query present")
			log.Trace("Invalid Session, not serving", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			Res(w, Response{Success: false, Message: "access denied, please authenticate", Error: "authentication required"})
			return
		}
		validCredentials, err := user.ValidateCredentials(username, password)
		if err != nil {
			// The database could not verify the given credentials
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			Res(w, Response{Success: false, Message: "could not authenticate: failed to validate credentials", Error: "database failure"})
			return
		}
		if validCredentials {
			// Supplied credentials are valid and the session should be saved
			session, _ := Store.Get(r, "session")
			session.Values["valid"] = true
			session.Values["username"] = username
			if err := session.Save(r, w); err != nil {
				log.Error("Failed to save session: ", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				Res(w, Response{Success: false, Message: "failed to authenticate", Error: "could not save session after successful authentication"})
			}
			log.Trace(fmt.Sprintf("valid query: serving %s", r.URL.Path))
			handler.ServeHTTP(w, r)
			return
		} else {
			// The database could validate the credentials but they were invalid
			log.Trace("bad credentials, invalid Session: not serving", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			Res(w, Response{Success: false, Message: "access denied, wrong username or password", Error: "invalid credentials"})
			return
		}
	}
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
			// this should not happen
			// Either a session or query will be valid (middleware.AuthRequired checks for this requirement)
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
		log.Error("Could not use GetUserFromQuery: failed validate login credentials due to database failure", err.Error())
		return "", false, err
	}
	if loginValid {
		return username, true, nil
	} else {
		return "", false, nil
	}
}

// Middleware for checking if a user has permission to access a given route
// The permission to check is given as a second argument as a string
// Make sure that the permission to check exists before checking it here
func Perm(handler http.HandlerFunc, permissionToCheck database.PermissionType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := GetUserFromCurrentSession(w, r)
		log.Trace(fmt.Sprintf("Checking permission `%s` for user `%s`", permissionToCheck, username))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			Res(w, Response{Success: false, Message: "access denied, invalid session", Error: "clear your browser's cookies"})
			return
		}
		hasPermission, err := database.UserHasPermission(username, permissionToCheck)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			Res(w, Response{Success: false, Message: "database error", Error: "failed to check permission to access this ressource"})
			return
		}
		if !hasPermission {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			Res(w, Response{Success: false, Message: "permission denied", Error: "missing permission to access this ressource, contact your administrator"})
			return
		}
		handler.ServeHTTP(w, r)
	}
}
