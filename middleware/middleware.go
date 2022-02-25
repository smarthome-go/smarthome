package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/utils"
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
		loginValid, err := utils.ValidateLogin(username, password)
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

		if ok && okParse && valid {
			log.Trace(fmt.Sprintf("Valid Session, serving %s", r.URL.Path))
			handler.ServeHTTP(w, r)
			return
		}
		loginValid, err := utils.ValidateLogin(username, password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if loginValid {
			session, _ := Store.Get(r, "session")
			session.Values["valid"] = true
			session.Values["username"] = username
			session.Save(r, w)
			log.Trace(fmt.Sprintf("Invalid Session, but authenticated with query: Session Saved. Serving %s", r.URL.Path))
			handler.ServeHTTP(w, r)
			return
		}
		log.Trace(fmt.Sprintf("Invalid Session, redirecting %s to /login", r.URL.Path))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ResponseStruct{false, 401, "Access denied", "You must be authenticated."})
	}
}
