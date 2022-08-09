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
