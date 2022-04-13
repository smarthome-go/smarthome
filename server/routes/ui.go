package routes

import (
	"net/http"

	"github.com/MikMuellerDev/smarthome/server/middleware"
	"github.com/MikMuellerDev/smarthome/server/templates"
)

// Redirects to the dashboard
func indexGetHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dash", http.StatusSeeOther)
}

// Serves HTML for the dashboard
func dashGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "dash.html", http.StatusOK)
}

// Serves HTML for rooms
func roomsGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "rooms.html", http.StatusOK)
}

// Serves HTML for reminders
func reminderGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "reminders.html", http.StatusOK)
}

// Serves HTML for user management
func usersGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "users.html", http.StatusOK)
}

// If not user is logged in, it serves the HTML for the login page
// Otherwise the user is redirected to the dashboard
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

// Serves HTML for profile settings
func userProfileGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "profile.html", http.StatusOK)
}
