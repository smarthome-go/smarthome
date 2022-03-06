package routes

import (
	"net/http"

	"github.com/MikMuellerDev/smarthome/server/middleware"
	"github.com/MikMuellerDev/smarthome/server/templates"
)

func indexGetHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dash", http.StatusSeeOther)
}

func dashGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "dash.html", http.StatusOK)
}

func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := middleware.Store.Get(r, "session")
	loginValidTemp, loginValidOkTemp := session.Values["valid"]
	loginValid, loginValidOk := loginValidTemp.(bool)
	if loginValidOkTemp && loginValidOk && loginValid {
		http.Redirect(w, r, "/dash", http.StatusFound)
		return
	}
	templates.ExecuteTemplate(w, "login.html", http.StatusOK)
}
