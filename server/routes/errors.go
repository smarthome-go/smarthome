package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MikMuellerDev/smarthome/server/api"
	"github.com/MikMuellerDev/smarthome/server/templates"
)

// If a `404 - not found` error occurs, this page is served, no authentication required
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	path := strings.Split(r.URL.Path, "/")
	if len(path) >= 1 {
		if path[1] == "api" {
			// Any subpath under `/api` which was not found which means that a response struct should be returned instead if html
			api.Res(w, api.Response{Success: false, Message: "not found", Error: fmt.Sprintf("The url `%s` could not be found", r.URL.Path)})
			return
		}
	}
	templates.ExecuteTemplate(w, "404.html", http.StatusNotFound)
}

// If a `405 - method not allowed` error occurs, this endpoint will return the error in a JSON format, no authentication required
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	api.Res(w, api.Response{Success: false, Message: fmt.Sprintf("The method `%s` is invalid: method not allowed", r.Method), Error: "method not allowed"})
}
