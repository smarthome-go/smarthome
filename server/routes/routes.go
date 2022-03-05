package routes

import (
	"net/http"

	"github.com/MikMuellerDev/smarthome/server/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", middleware.AuthRequired(indexGetHandler)).Methods("GET")
	r.HandleFunc("/dash", middleware.AuthRequired(dashGetHandler)).Methods("GET")

	// TODO: modify loginGethandler to redirect to the dashboard if the user is alerady logged in
	r.HandleFunc("/login", loginGetHandler).Methods("GET")
	r.HandleFunc("/logout", logoutGetHandler).Methods("GET")

	//// API ////
	r.HandleFunc("/api/login", loginPostHandler).Methods("POST")

	// Public endpoints without authentication
	r.HandleFunc("/api/power/list", getSwitches).Methods("GET")
	r.HandleFunc("/api/power/states", getPowerStates).Methods("GET")

	// Api / Power (with authentication)
	r.HandleFunc("/api/power/set", middleware.ApiAuthRequired(middleware.Permission(powerPostHandler, "setPower"))).Methods("POST")
	r.HandleFunc("/api/power/list/personal", middleware.ApiAuthRequired(middleware.Permission(getUserSwitches, "getUserSwitches"))).Methods("GET")

	// Logs for the admin user
	r.HandleFunc("/api/logs/delete/old", middleware.ApiAuthRequired(middleware.Permission(flushOldLogs, "deleteOldLogs"))).Methods("DELETE")
	r.HandleFunc("/api/logs/delete/all", middleware.ApiAuthRequired(middleware.Permission(flushAllLogs, "deleteAllLogs"))).Methods("DELETE")
	r.HandleFunc("/api/logs/get", middleware.ApiAuthRequired(middleware.Permission(listLogs, "listLogs"))).Methods("GET")

	// Get personal permissions
	r.HandleFunc("/api/user/permissions/personal", middleware.ApiAuthRequired(getUserPermissions))

	// Profile picture upload test
	r.HandleFunc("/profile/upload", handleProfileUpload).Methods("POST")

	/// Static files ///
	// For JS and CSS components
	outFilepath := "./web/out/"
	staticPathPrefix := "/static"
	outFileserver := http.FileServer(http.Dir(outFilepath))
	r.PathPrefix(staticPathPrefix).Handler(http.StripPrefix(staticPathPrefix, outFileserver))

	// Other assets, such as PNG or JPEG
	assetsFilepath := "./web/assets/"
	assetsPathPrefix := "/assets"
	assetsFileserver := http.FileServer(http.Dir(assetsFilepath))
	r.PathPrefix(assetsPathPrefix).Handler(http.StripPrefix(assetsPathPrefix, assetsFileserver))

	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	log.Debug("Initialized Router.")
	return r
}
