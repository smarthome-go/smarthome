package routes

import (
	"net/http"

	// `mdl`` is shorter than `middleware`
	mdl "github.com/MikMuellerDev/smarthome/server/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	// Auth: middleware that checks if the user is logged in, will redirect to `/login` if the user is not logged in
	// ApiAuth: middleware that checks if the user is logged in for API request, will return JSON errors if the user is not logged in
	r.HandleFunc("/", mdl.Auth(indexGetHandler)).Methods("GET")
	r.HandleFunc("/dash", mdl.Auth(dashGetHandler)).Methods("GET")

	// Healthcheck for uptime monitoring
	r.HandleFunc("/health", healthCheck).Methods("GET")

	// User profile (settings)
	r.HandleFunc("/profile", mdl.Auth(userProfileGetHandler)).Methods("GET")

	r.HandleFunc("/login", loginGetHandler).Methods("GET")
	r.HandleFunc("/logout", logoutGetHandler).Methods("GET")
	r.HandleFunc("/api/login", loginPostHandler).Methods("POST")

	//// API ////
	r.HandleFunc("/api/power/list", getSwitches).Methods("GET")
	r.HandleFunc("/api/power/states", getPowerStates).Methods("GET")
	r.HandleFunc("/api/power/set", mdl.ApiAuth(mdl.Perm(powerPostHandler, "setPower"))).Methods("POST")
	r.HandleFunc("/api/power/list/personal", mdl.ApiAuth(mdl.Perm(getUserSwitches, "getUserSwitches"))).Methods("GET")

	// Logs for the admin user
	r.HandleFunc("/api/logs/delete/old", mdl.ApiAuth(mdl.Perm(flushOldLogs, "deleteOldLogs"))).Methods("DELETE")
	r.HandleFunc("/api/logs/delete/all", mdl.ApiAuth(mdl.Perm(flushAllLogs, "deleteAllLogs"))).Methods("DELETE")
	r.HandleFunc("/api/logs/get", mdl.ApiAuth(mdl.Perm(listLogs, "listLogs"))).Methods("GET")

	// Customization for the user
	// Profile picture upload test
	r.HandleFunc("/api/user/avatar", mdl.ApiAuth(getAvatar)).Methods("GET")
	r.HandleFunc("/api/user/avatar/upload", mdl.ApiAuth(mdl.Perm(handleAvatarUpload, "uploadAvatar"))).Methods("POST")
	r.HandleFunc("/api/user/avatar/delete", mdl.ApiAuth(mdl.Perm(deleteAvatar, "deleteAvatar"))).Methods("DELETE")

	// Permissions
	r.HandleFunc("/api/user/permissions/personal", mdl.ApiAuth(getUserPermissions))
	r.HandleFunc("/api/user/permissions/add", mdl.ApiAuth(mdl.Perm(addUserPermission, "addUserPermission"))).Methods("PUT")
	r.HandleFunc("/api/user/permissions/delete", mdl.ApiAuth(mdl.Perm(removeUserPermission, "removeUserPermission"))).Methods("DELETE")

	// Switch Permissions
	r.HandleFunc("/api/user/permissions/switch/add", mdl.ApiAuth(mdl.Perm(addSwitchPermission, "addSwitchPermission"))).Methods("PUT")
	r.HandleFunc("/api/user/permissions/switch/delete", mdl.ApiAuth(mdl.Perm(removeSwitchPermission, "removeSwitchPermission"))).Methods("DELETE")

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
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	log.Debug("Initialized Router.")
	return r
}
