package middleware

import (
	"fmt"
	"net/http"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/user"
)

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
		token := query.Get("token")

		if loginValidOkTemp && loginValidOk && loginValid && username == "" && token == "" {
			// Check if the user exists
			// TODO: maybe implement a check
			if usernameTempOk && usernameSessionOk && usernameSession != "" {
				_, exists, err := database.GetUserByUsername(usernameSession)
				if err != nil {
					w.WriteHeader(http.StatusServiceUnavailable)
					Res(w, Response{Success: false, Message: "Could not check user validity", Error: "database failure"})
					return
				}
				if exists {
					// Do not return an error if the does not exists to allow correction via URL queries
					// The session is valid: allow access
					handler.ServeHTTP(w, r)
					return
				}
			}
		}

		// If the provided url query username and token are blank, redirect to login
		if username == "" && token == "" {
			log.Trace(fmt.Sprintf("Invalid Session, redirecting %s to /login", r.URL.Path))
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Check potential credentials or the token if the session is invalid
		var validCredentials bool
		validCredentials, err = user.ValidateCredentials(username, password)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			Res(w, Response{Success: false, Message: "could not validate credentials", Error: "database failure"})
			return
		}
		// Check the token if everything else fails
		if !validCredentials {
			data, found, err := database.GetUserTokenByToken(token)
			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				Res(w, Response{Success: false, Message: "could not validate authentication token", Error: "database failure"})
				return
			}
			if found {
				validCredentials = true
				username = data.User
			}
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
