package middleware

import (
	"fmt"
	"net/http"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/user"
)

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
