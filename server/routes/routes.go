package routes

import (
	"net/http"
	// `mdl`` is shorter than `middleware`

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/server/api"
	mdl "github.com/smarthome-go/smarthome/server/middleware"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	// Auth: middleware that checks if the user is logged in, will redirect to `/login` if the user is not logged in
	// ApiAuth: middleware that checks if the user is logged in for API request, will return JSON errors if the user is not logged in

	// Dashboard
	r.HandleFunc("/", mdl.Auth(indexGetHandler)).Methods("GET")
	r.HandleFunc("/dash", mdl.Auth(dashGetHandler)).Methods("GET")
	r.HandleFunc("/rooms", mdl.Auth(roomsGetHandler)).Methods("GET")
	r.HandleFunc("/reminders", mdl.Auth(reminderGetHandler)).Methods("GET")
	r.HandleFunc("/profile", mdl.Auth(userProfileGetHandler)).Methods("GET")
	r.HandleFunc("/users", mdl.Auth(usersGetHandler)).Methods("GET")
	r.HandleFunc("/editor", mdl.Auth(editorGetHandler)).Methods("GET")
	r.HandleFunc("/automations", mdl.Auth(automationsGetHandler)).Methods("GET")

	// Healthcheck for uptime monitoring
	r.HandleFunc("/health", api.HealthCheck).Methods("GET")

	// Debug information about the system
	r.HandleFunc("/api/debug", mdl.ApiAuth(mdl.Perm(api.DebugInfo, database.PermissionDebug))).Methods("GET")

	r.HandleFunc("/login", loginGetHandler).Methods("GET")
	r.HandleFunc("/logout", logoutGetHandler).Methods("GET")
	r.HandleFunc("/api/login", loginPostHandler).Methods("POST")

	//// API ////
	// Power
	r.HandleFunc("/api/power/states", api.GetPowerStates).Methods("GET")
	r.HandleFunc("/api/power/set", mdl.ApiAuth(mdl.Perm(api.PowerPostHandler, database.PermissionPower))).Methods("POST")

	// Rooms
	r.HandleFunc("/api/room/list/all", api.ListAllRoomsWithSwitches).Methods("GET")
	r.HandleFunc("/api/room/list/personal", mdl.ApiAuth(api.ListUserRoomsWithSwitches)).Methods("GET")
	r.HandleFunc("/api/room/add", mdl.ApiAuth(mdl.Perm(api.AddRoom, database.PermissionModifyRooms))).Methods("POST")
	r.HandleFunc("/api/room/modify", mdl.ApiAuth(mdl.Perm(api.ModifyRoomData, database.PermissionModifyRooms))).Methods("PUT")
	r.HandleFunc("/api/room/delete", mdl.ApiAuth(mdl.Perm(api.DeleteRoom, database.PermissionModifyRooms))).Methods("DELETE")

	// Switches
	r.HandleFunc("/api/switch/list/all", api.GetAllSwitches).Methods("GET")
	r.HandleFunc("/api/switch/list/personal", mdl.ApiAuth(api.GetUserSwitches)).Methods("GET")
	r.HandleFunc("/api/switch/add", mdl.ApiAuth(mdl.Perm(api.CreateSwitch, database.PermissionModifyRooms))).Methods("POST")
	r.HandleFunc("/api/switch/modify", mdl.ApiAuth(mdl.Perm(api.ModifySwitch, database.PermissionModifyRooms))).Methods("PUT")
	r.HandleFunc("/api/switch/delete", mdl.ApiAuth(mdl.Perm(api.DeleteSwitch, database.PermissionModifyRooms))).Methods("DELETE")

	// Cameras
	r.HandleFunc("/api/camera/add", mdl.ApiAuth(mdl.Perm(api.CreateCamera, database.PermissionModifyRooms))).Methods("POST")
	r.HandleFunc("/api/camera/modify", mdl.ApiAuth(mdl.Perm(api.ModifyCamera, database.PermissionModifyRooms))).Methods("PUT")
	r.HandleFunc("/api/camera/delete", mdl.ApiAuth(mdl.Perm(api.DeleteCamera, database.PermissionModifyRooms))).Methods("DELETE")
	r.HandleFunc("/api/camera/list/all", mdl.ApiAuth(mdl.Perm(api.GetAllCameras, database.PermissionModifyRooms))).Methods("GET")
	r.HandleFunc("/api/camera/list/redacted", mdl.ApiAuth(api.GetAllRedactedCameras)).Methods("GET")
	r.HandleFunc("/api/camera/list/personal", mdl.ApiAuth(mdl.Perm(api.GetCurrentUserCameras, database.PermissionViewCameras))).Methods("GET")

	// Logs for the admin user
	r.HandleFunc("/api/logs/delete/old", mdl.ApiAuth(mdl.Perm(api.FlushOldLogs, database.PermissionLogs))).Methods("DELETE")
	r.HandleFunc("/api/logs/delete/all", mdl.ApiAuth(mdl.Perm(api.FlushAllLogs, database.PermissionLogs))).Methods("DELETE")
	r.HandleFunc("/api/logs", mdl.ApiAuth(mdl.Perm(api.ListLogs, database.PermissionLogs))).Methods("GET")

	// Customization for the user
	// Profile picture upload test
	r.HandleFunc("/api/user/avatar/personal", mdl.ApiAuth(api.GetAvatar)).Methods("GET")
	r.HandleFunc("/api/user/avatar/user/{username}", mdl.ApiAuth(api.GetForeignUserAvatar)).Methods("GET")
	r.HandleFunc("/api/user/avatar/upload", mdl.ApiAuth(api.HandleAvatarUpload)).Methods("POST")
	r.HandleFunc("/api/user/avatar/delete", mdl.ApiAuth(api.DeleteAvatar)).Methods("DELETE")

	/** Permissions */
	// Normal Permissions
	r.HandleFunc("/api/user/permissions/add", mdl.ApiAuth(mdl.Perm(api.AddUserPermission, database.PermissionManageUsers))).Methods("POST")
	r.HandleFunc("/api/user/permissions/delete", mdl.ApiAuth(mdl.Perm(api.RemoveUserPermission, database.PermissionManageUsers))).Methods("DELETE")
	r.HandleFunc("/api/permissions/list/all", api.ListPermissions).Methods("GET")
	r.HandleFunc("/api/user/permissions/list/personal", mdl.ApiAuth(api.GetCurrentUserPermissions)).Methods("GET")
	r.HandleFunc("/api/user/permissions/list/user/{username}", mdl.ApiAuth(mdl.Perm(api.GetForeignUserPermissions, database.PermissionManageUsers))).Methods("GET")

	// Switch Permissions
	r.HandleFunc("/api/user/permissions/switch/add", mdl.ApiAuth(mdl.Perm(api.AddSwitchPermission, database.PermissionManageUsers))).Methods("POST")
	r.HandleFunc("/api/user/permissions/switch/delete", mdl.ApiAuth(mdl.Perm(api.RemoveSwitchPermission, database.PermissionManageUsers))).Methods("DELETE")
	r.HandleFunc("/api/user/permissions/switch/list/user/{username}", mdl.ApiAuth(mdl.Perm(api.GetForeignUserSwitchPermissions, database.PermissionManageUsers))).Methods("GET")

	// Camera Permissions
	r.HandleFunc("/api/user/permissions/camera/add", mdl.ApiAuth(mdl.Perm(api.AddCameraPermission, database.PermissionManageUsers))).Methods("POST")
	r.HandleFunc("/api/user/permissions/camera/delete", mdl.ApiAuth(mdl.Perm(api.RemoveCameraPermission, database.PermissionManageUsers))).Methods("DELETE")
	r.HandleFunc("/api/user/permissions/camera/list/user/{username}", mdl.ApiAuth(mdl.Perm(api.GetForeignUserCameraPermission, database.PermissionManageUsers))).Methods("GET")

	// Creating and removing users
	r.HandleFunc("/api/user/manage/list", mdl.ApiAuth(mdl.Perm(api.ListUsers, database.PermissionManageUsers))).Methods("GET")
	r.HandleFunc("/api/user/manage/add", mdl.ApiAuth(mdl.Perm(api.AddUser, database.PermissionManageUsers))).Methods("POST")
	r.HandleFunc("/api/user/manage/modify", mdl.ApiAuth(mdl.Perm(api.AddUser, database.PermissionManageUsers))).Methods("PUT")
	r.HandleFunc("/api/user/manage/delete", mdl.ApiAuth(mdl.Perm(api.DeleteUser, database.PermissionManageUsers))).Methods("DELETE")
	r.HandleFunc("/api/user/manage/data/modify", mdl.ApiAuth(mdl.Perm(api.ModifyUserMetadata, database.PermissionManageUsers))).Methods("PUT")

	// Manage personal data
	r.HandleFunc("/api/user/data", mdl.ApiAuth(api.GetUserDetails)).Methods("GET")
	r.HandleFunc("/api/user/data/update", mdl.ApiAuth(api.ModifyCurrentUserMetadata)).Methods("PUT")

	// Notification-related
	r.HandleFunc("/api/user/notification/count", mdl.ApiAuth(api.GetNotificationCount)).Methods("GET")
	r.HandleFunc("/api/user/notification/delete", mdl.ApiAuth(api.DeleteUserNotificationById)).Methods("DELETE")
	r.HandleFunc("/api/user/notification/delete/all", mdl.ApiAuth(api.DeleteAllUserNotifications)).Methods("DELETE")
	r.HandleFunc("/api/user/notification/list", mdl.ApiAuth(api.GetNotifications)).Methods("GET")

	// Homescript-related
	r.HandleFunc("/api/homescript/add", mdl.ApiAuth(mdl.Perm(api.CreateNewHomescript, database.PermissionHomescript))).Methods("POST")
	r.HandleFunc("/api/homescript/modify", mdl.ApiAuth(mdl.Perm(api.ModifyHomescript, database.PermissionHomescript))).Methods("PUT")
	r.HandleFunc("/api/homescript/delete", mdl.ApiAuth(mdl.Perm(api.DeleteHomescriptById, database.PermissionHomescript))).Methods("DELETE")
	r.HandleFunc("/api/homescript/run/live", mdl.ApiAuth(mdl.Perm(api.RunHomescriptString, database.PermissionHomescript))).Methods("POST")
	r.HandleFunc("/api/homescript/get/{id}", mdl.ApiAuth(mdl.Perm(api.GetUserHomescriptById, database.PermissionHomescript))).Methods("GET")
	r.HandleFunc("/api/homescript/list/personal", mdl.ApiAuth(api.ListPersonalHomescripts)).Methods("GET")

	// Automations-related
	r.HandleFunc("/api/automation/list/personal", mdl.ApiAuth(mdl.Perm(api.GetUserAutomations, database.PermissionAutomation))).Methods("GET")
	r.HandleFunc("/api/automation/add", mdl.ApiAuth(mdl.Perm(api.CreateNewAutomation, database.PermissionAutomation))).Methods("POST")
	r.HandleFunc("/api/automation/delete", mdl.ApiAuth(mdl.Perm(api.RemoveAutomation, database.PermissionAutomation))).Methods("DELETE")
	r.HandleFunc("/api/automation/modify", mdl.ApiAuth(mdl.Perm(api.ModifyAutomation, database.PermissionAutomation))).Methods("PUT")
	r.HandleFunc("/api/automation/state/global", mdl.ApiAuth(mdl.Perm(api.ChangeActivationAutomation, database.PermissionModifyServerConfig))).Methods("PUT")

	// Schedule-related
	r.HandleFunc("/api/scheduler/list/personal", mdl.ApiAuth(mdl.Perm(api.GetUserSchedules, database.PermissionScheduler))).Methods("GET")
	r.HandleFunc("/api/scheduler/add", mdl.ApiAuth(mdl.Perm(api.CreateNewSchedule, database.PermissionScheduler))).Methods("POST")
	r.HandleFunc("/api/scheduler/delete", mdl.ApiAuth(mdl.Perm(api.RemoveSchedule, database.PermissionScheduler))).Methods("DELETE")
	r.HandleFunc("/api/scheduler/modify", mdl.ApiAuth(mdl.Perm(api.ModifySchedule, database.PermissionScheduler))).Methods("PUT")
	r.HandleFunc("/api/scheduler/state/personal", mdl.ApiAuth(mdl.Perm(api.SetCurrentUserSchedulerEnabled, database.PermissionScheduler))).Methods("PUT")
	r.HandleFunc("/api/scheduler/state/user", mdl.ApiAuth(mdl.Perm(api.SetUserSchedulerEnabled, database.PermissionManageUsers))).Methods("PUT")

	// Admin-specific
	r.HandleFunc("/api/config/location/modify", mdl.ApiAuth(mdl.Perm(api.UpdateLocation, database.PermissionModifyServerConfig))).Methods("PUT")

	// Customization
	r.HandleFunc("/api/user/settings/theme/personal", mdl.ApiAuth(api.SetCurrentUserColorTheme)).Methods("PUT")
	r.HandleFunc("/api/user/settings/theme/user", mdl.ApiAuth(api.SetUserColorTheme)).Methods("PUT")

	// Reminders
	r.HandleFunc("/api/reminder/add", mdl.ApiAuth(mdl.Perm(api.AddReminder, database.PermissionReminder))).Methods("POST")
	r.HandleFunc("/api/reminder/list", mdl.ApiAuth(mdl.Perm(api.GetReminders, database.PermissionReminder))).Methods("GET")
	r.HandleFunc("/api/reminder/modify", mdl.ApiAuth(mdl.Perm(api.ModifyReminder, database.PermissionReminder))).Methods("PUT")
	r.HandleFunc("/api/reminder/delete", mdl.ApiAuth(mdl.Perm(api.DeleteReminder, database.PermissionReminder))).Methods("DELETE")

	// TODO: remove this one below
	// Test camera module here
	r.HandleFunc("/api/camera/feed/{id}", api.GetCameraFeed).Methods("GET")

	/// Static files ///
	assetsFilepath := "./web/dist/assets/"
	assetsPathPrefix := "/assets"
	assetsFileserver := http.FileServer(http.Dir(assetsFilepath))
	r.PathPrefix(assetsPathPrefix).Handler(http.StripPrefix(assetsPathPrefix, assetsFileserver))

	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	log.Debug("Successfully initialized router")
	return r
}
