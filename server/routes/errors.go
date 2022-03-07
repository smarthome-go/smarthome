package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/server/templates"
)

// If a `404 - not found` error occurs, this page is served, no authentication required
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	templates.ExecuteTemplate(w, "404.html", http.StatusNotFound)
}

// If a `405 - method not allowed` error occurs, this endpoint will return the error in a JSON format, no authentication required
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(Response{Success: false, Message: fmt.Sprintf("The method `%s` is invalid: method not allowed", r.Method), Error: "method not allowed"})
}
