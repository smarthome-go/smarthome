package routes

import (
	"net/http"

	"github.com/MikMuellerDev/smarthome/templates"
)

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	templates.ExecuteTemplate(w, "404.html", http.StatusNotFound)
}
