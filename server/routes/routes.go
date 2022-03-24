package routes

import (
	"net/http"

	// `mdl`` is shorter than `middleware`
	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/server/api"
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
	r.HandleFunc("/rooms", mdl.Auth(roomsGetHandler)).Methods("GET")

	// Healthcheck for uptime monitoring
	r.HandleFunc("/health", api.HealthCheck).Methods("GET")

	// Debug information about the system
	r.HandleFunc("/api/debug", mdl.ApiAuth(mdl.Perm(api.DebugInfo, database.PermissionGetDebugInfo))).Methods("GET")

	// User profile (settings)
	r.HandleFunc("/profile", mdl.Auth(userProfileGetHandler)).Methods("GET")

	r.HandleFunc("/login", loginGetHandler).Methods("GET")
	r.HandleFunc("/logout", logoutGetHandler).Methods("GET")
	r.HandleFunc("/api/login", loginPostHandler).Methods("POST")

	//// API ////
	// Power
	r.HandleFunc("/api/switch/list", api.GetSwitches).Methods("GET")
	r.HandleFunc("/api/switch/list/personal", mdl.ApiAuth(api.GetUserSwitches)).Methods("GET")
	r.HandleFunc("/api/power/states", api.GetPowerStates).Methods("GET")
	r.HandleFunc("/api/power/set", mdl.ApiAuth(mdl.Perm(api.PowerPostHandler, database.PermissionSetPower))).Methods("POST")

	// Rooms
	r.HandleFunc("/api/room/list/personal", mdl.ApiAuth(api.GetUserRoomsWithSwitches)).Methods("GET")

	// Logs for the admin user
	r.HandleFunc("/api/logs/delete/old", mdl.ApiAuth(mdl.Perm(api.FlushOldLogs, database.PermissionDeleteLogs))).Methods("DELETE")
	r.HandleFunc("/api/logs/delete/all", mdl.ApiAuth(mdl.Perm(api.FlushAllLogs, database.PermissionDeleteLogs))).Methods("DELETE")
	r.HandleFunc("/api/logs", mdl.ApiAuth(mdl.Perm(api.ListLogs, database.PermissionListLogs))).Methods("GET")

	// Customization for the user
	// Profile picture upload test
	r.HandleFunc("/api/user/avatar", mdl.ApiAuth(getAvatar)).Methods("GET")
	r.HandleFunc("/api/user/avatar/upload", mdl.ApiAuth(mdl.Perm(handleAvatarUpload, database.PermissionChangeAvatar))).Methods("POST")
	r.HandleFunc("/api/user/avatar/delete", mdl.ApiAuth(mdl.Perm(deleteAvatar, database.PermissionChangeAvatar))).Methods("DELETE")

	// Permissions
	r.HandleFunc("/api/user/permissions/personal", mdl.ApiAuth(api.GetUserPermissions))
	r.HandleFunc("/api/user/permissions/add", mdl.ApiAuth(mdl.Perm(api.AddUserPermission, database.PermissionChangeUserPermissions))).Methods("PUT")
	r.HandleFunc("/api/user/permissions/delete", mdl.ApiAuth(mdl.Perm(api.RemoveUserPermission, database.PermissionChangeUserPermissions))).Methods("DELETE")

	// Switch Permissions
	r.HandleFunc("/api/user/permissions/switch/add", mdl.ApiAuth(mdl.Perm(api.AddSwitchPermission, database.PermissionChangeSwitchPermissions))).Methods("PUT")
	r.HandleFunc("/api/user/permissions/switch/delete", mdl.ApiAuth(mdl.Perm(api.RemoveSwitchPermission, database.PermissionChangeSwitchPermissions))).Methods("DELETE")

	// Creating and removing users
	r.HandleFunc("/api/user/list", mdl.ApiAuth(mdl.Perm(api.ListUsers, database.PermissionListUsers))).Methods("GET")
	r.HandleFunc("/api/user/add", mdl.ApiAuth(mdl.Perm(api.AddUser, database.PermissionChangeUsers))).Methods("POST")
	r.HandleFunc("/api/user/delete", mdl.ApiAuth(mdl.Perm(api.DeleteUser, database.PermissionChangeUsers))).Methods("DELETE")

	// Get personal details
	r.HandleFunc("/api/user/data", mdl.ApiAuth(api.GetUserDetails)).Methods("GET")

	// Notification-related
	r.HandleFunc("/api/user/notification/count", mdl.ApiAuth(api.GetNotificationCount)).Methods("GET")
	r.HandleFunc("/api/user/notification/delete", mdl.ApiAuth(api.DeleteUserNotificationById)).Methods("DELETE")
	r.HandleFunc("/api/user/notification/delete/all", mdl.ApiAuth(api.DeleteAllUserNotifications)).Methods("DELETE")
	r.HandleFunc("/api/user/notification/list", mdl.ApiAuth(api.GetNotifications)).Methods("GET")

	// Homescript-related
	r.HandleFunc("/api/homescript/run/live", mdl.ApiAuth(mdl.Perm(api.RunHomescriptString, database.PermissionRunHomescript))).Methods("POST")
	r.HandleFunc("/api/homescript/list/personal", mdl.ApiAuth(api.ListPersonalHomescripts)).Methods("GET")

	// TODO: add removal functions

	// TODO: remove this one below
	// Test camera module here
	r.HandleFunc("/api/camera/test", api.TestImageProxy).Methods("GET")

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
