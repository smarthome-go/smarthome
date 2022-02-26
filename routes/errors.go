package routes

import (
	"net/http"

	"github.com/MikMuellerDev/smarthome/templates"
)

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "404.html", http.StatusNotFound)
}
