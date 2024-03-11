package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/device/driver"
	"github.com/smarthome-go/smarthome/server/api"

	// `mdl` is shorter than `middleware`
	mdl "github.com/smarthome-go/smarthome/server/middleware"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func NewRouter() *mux.Router {
	log.Trace("Initializing server router...")
	r := mux.NewRouter()
	/*
		Middleware explanation
		Auth: middleware that checks if the user is logged in, will redirect to `/login` if the user is not logged in
		ApiAuth: middleware that checks if the user is logged in for API request, will return JSON errors if the user is not logged in
	*/

	// Health check for uptime monitoring
	r.HandleFunc("/health", api.HealthCheck).Methods("GET")

	/*
		=== Pages ===
		All routes below belong to pages or HTML serving
	*/

	// HTML-serving endpoints
	r.HandleFunc("/", mdl.Auth(indexGetHandler)).Methods("GET")
	r.HandleFunc("/dash", mdl.Auth(dashGetHandler)).Methods("GET")
	r.HandleFunc("/rooms", mdl.Auth(roomsGetHandler)).Methods("GET")
	r.HandleFunc("/reminders", mdl.Auth(reminderGetHandler)).Methods("GET")
	r.HandleFunc("/scheduler", mdl.Auth(schedulerGetHandler)).Methods("GET")
	r.HandleFunc("/automations", mdl.Auth(automationsGetHandler)).Methods("GET")
	r.HandleFunc("/homescript", mdl.Auth(homescriptGetHandler)).Methods("GET")
	r.HandleFunc("/homescript/editor", mdl.Auth(hmsEditorGetHandler)).Methods("GET")
	r.HandleFunc("/profile", mdl.Auth(userProfileGetHandler)).Methods("GET")
	r.HandleFunc("/users", mdl.Auth(usersGetHandler)).Methods("GET")
	r.HandleFunc("/system", mdl.Auth(systemGetHandler)).Methods("GET")

	// Session management
	r.HandleFunc("/login", loginGetHandler).Methods("GET")
	r.HandleFunc("/logout", logoutGetHandler).Methods("GET")

	/*
		=== API ===
		All routes below belong to the API
	*/

	// Debug Information
	r.HandleFunc("/api/debug", mdl.ApiAuth(mdl.Perm(api.DebugInfo, database.PermissionDebug))).Methods("GET")

	// Version information
	r.HandleFunc("/api/version", api.GetVersionInfo).Methods("GET")

	// Login handler
	r.HandleFunc("/api/login", userLoginHandler).Methods("POST")
	r.HandleFunc("/api/login/token", tokenLoginHandler).Methods("POST")

	// Power
	// TODO: implement this using the power API that is implemented later
	// TODO: also implement a sensor / input API
	// r.HandleFunc("/api/power/states", api.GetPowerStates).Methods("GET")
	// r.HandleFunc("/api/power/usage/day", api.GetPowerDrawFrom24Hours).Methods("GET")
	// r.HandleFunc("/api/power/usage/all", mdl.ApiAuth(api.GetPowerDrawAll)).Methods("GET")
	// r.HandleFunc("/api/power/set", mdl.ApiAuth(mdl.Perm(api.PowerPostHandler, database.PermissionPower))).Methods("POST")

	// Rooms
	r.HandleFunc("/api/room/list/all", mdl.ApiAuth(api.ListAllRoomsWithData)).Methods("GET")
	r.HandleFunc("/api/room/list/personal", mdl.ApiAuth(api.ListUserRoomsWithData)).Methods("GET")
	r.HandleFunc("/api/room/add", mdl.ApiAuth(mdl.Perm(api.AddRoom, database.PermissionModifyRooms))).Methods("POST")
	r.HandleFunc("/api/room/modify", mdl.ApiAuth(mdl.Perm(api.ModifyRoomData, database.PermissionModifyRooms))).Methods("PUT")
	r.HandleFunc("/api/room/delete", mdl.ApiAuth(mdl.Perm(api.DeleteRoom, database.PermissionModifyRooms))).Methods("DELETE")

	// Devices
	r.HandleFunc("/api/devices/list/all", api.GetAllDevices).Methods("GET")
	r.HandleFunc("/api/devices/list/personal", mdl.ApiAuth(api.GetUserDevices)).Methods("GET")

	r.HandleFunc("/api/devices/list/all/rich", api.GetAllDevicesRich).Methods("GET")
	r.HandleFunc("/api/devices/list/personal/rich", mdl.ApiAuth(api.GetUserDevicesRich)).Methods("GET")

	r.HandleFunc("/api/devices/extract/{id}", mdl.ApiAuth(api.ExtractUserDevice)).Methods("GET")

	r.HandleFunc("/api/devices/add", mdl.ApiAuth(mdl.Perm(api.CreateDevice, database.PermissionModifyRooms))).Methods("POST")
	r.HandleFunc("/api/devices/modify", mdl.ApiAuth(mdl.Perm(api.ModifyDevice, database.PermissionModifyRooms))).Methods("PUT")
	r.HandleFunc("/api/devices/delete", mdl.ApiAuth(mdl.Perm(api.DeleteDevice, database.PermissionModifyRooms))).Methods("DELETE")
	r.HandleFunc("/api/devices/configure", mdl.ApiAuth(mdl.Perm(api.ConfigureDevice, database.PermissionModifyRooms))).Methods("PUT")

	// TODO: Device actions???
	r.HandleFunc("/api/devices/action/power",
		mdl.ApiAuth(mdl.Perm(api.DeviceActionHandlerFactory(driver.DriverActionKindSetPower), database.PermissionPower)),
	).Methods("POST")

	r.HandleFunc("/api/devices/action/dim",
		mdl.ApiAuth(mdl.Perm(api.DeviceActionHandlerFactory(driver.DriverActionKindDim), database.PermissionPower)),
	).Methods("POST")

	// Cameras
	r.HandleFunc("/api/camera/add", mdl.ApiAuth(mdl.Perm(api.CreateCamera, database.PermissionModifyRooms))).Methods("POST")
	r.HandleFunc("/api/camera/modify", mdl.ApiAuth(mdl.Perm(api.ModifyCamera, database.PermissionModifyRooms))).Methods("PUT")
	r.HandleFunc("/api/camera/delete", mdl.ApiAuth(mdl.Perm(api.DeleteCamera, database.PermissionModifyRooms))).Methods("DELETE")
	r.HandleFunc("/api/camera/list/all", mdl.ApiAuth(mdl.Perm(api.GetAllCameras, database.PermissionModifyRooms))).Methods("GET")
	r.HandleFunc("/api/camera/list/redacted", mdl.ApiAuth(api.GetAllRedactedCameras)).Methods("GET")
	r.HandleFunc("/api/camera/list/personal", mdl.ApiAuth(mdl.Perm(api.GetCurrentUserCameras, database.PermissionViewCameras))).Methods("GET")
	r.HandleFunc("/api/camera/feed/{id}", mdl.ApiAuth(mdl.Perm(api.GetCameraFeed, database.PermissionViewCameras))).Methods("GET")

	// Normal Permissions
	r.HandleFunc("/api/user/permissions/add", mdl.ApiAuth(mdl.Perm(api.AddUserPermission, database.PermissionManageUsers))).Methods("POST")
	r.HandleFunc("/api/user/permissions/delete", mdl.ApiAuth(mdl.Perm(api.RemoveUserPermission, database.PermissionManageUsers))).Methods("DELETE")
	r.HandleFunc("/api/permissions/list/all", api.ListPermissions).Methods("GET")
	r.HandleFunc("/api/user/permissions/list/personal", mdl.ApiAuth(api.GetCurrentUserPermissions)).Methods("GET")
	r.HandleFunc("/api/user/permissions/list/user/{username}", mdl.ApiAuth(mdl.Perm(api.GetForeignUserPermissions, database.PermissionManageUsers))).Methods("GET")

	// Device Permissions
	r.HandleFunc("/api/user/permissions/device/add", mdl.ApiAuth(mdl.Perm(api.AddDevicePermission, database.PermissionManageUsers))).Methods("POST")
	r.HandleFunc("/api/user/permissions/device/delete", mdl.ApiAuth(mdl.Perm(api.RemoveDevicePermission, database.PermissionManageUsers))).Methods("DELETE")
	r.HandleFunc("/api/user/permissions/device/list/user/{username}", mdl.ApiAuth(mdl.Perm(api.GetForeignUserDevicePermissions, database.PermissionManageUsers))).Methods("GET")

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

	// User Data
	r.HandleFunc("/api/user/data", mdl.ApiAuth(api.GetUserDetails)).Methods("GET")
	r.HandleFunc("/api/user/data/update", mdl.ApiAuth(api.ModifyCurrentUserMetadata)).Methods("PUT")
	r.HandleFunc("/api/user/password/modify", mdl.ApiAuth(api.ModifyCurrentUserPassword)).Methods("PUT")
	r.HandleFunc("/api/user/manage/delete/self", mdl.ApiAuth(api.DeleteCurrentUser)).Methods("DELETE")

	// User Customization
	r.HandleFunc("/api/user/settings/theme/personal", mdl.ApiAuth(api.SetCurrentUserColorTheme)).Methods("PUT")
	r.HandleFunc("/api/user/settings/theme/user", mdl.ApiAuth(api.SetUserColorTheme)).Methods("PUT")

	// Customization for the user
	r.HandleFunc("/api/user/avatar/personal", mdl.ApiAuth(api.GetAvatar)).Methods("GET")
	r.HandleFunc("/api/user/avatar/user/{username}", mdl.ApiAuth(api.GetForeignUserAvatar)).Methods("GET")

	// Personal avatar manipulation
	r.HandleFunc("/api/user/avatar/upload", mdl.ApiAuth(api.HandleAvatarUpload)).Methods("POST")
	r.HandleFunc("/api/user/avatar/delete", mdl.ApiAuth(api.DeleteAvatar)).Methods("DELETE")

	// Authentication Tokens
	r.HandleFunc("/api/user/token/generate", mdl.ApiAuth(api.GenerateUserToken)).Methods("POST")
	r.HandleFunc("/api/user/token/delete", mdl.ApiAuth(api.DeleteUserToken)).Methods("DELETE")
	r.HandleFunc("/api/user/token/list/personal", mdl.ApiAuth(api.ListUserTokens)).Methods("GET")

	// Notifications
	r.HandleFunc("/api/user/notification/notify", mdl.ApiAuth(api.NotifyUser)).Methods("POST")
	r.HandleFunc("/api/user/notification/count", mdl.ApiAuth(api.GetNotificationCount)).Methods("GET")
	r.HandleFunc("/api/user/notification/delete", mdl.ApiAuth(api.DeleteUserNotificationById)).Methods("DELETE")
	r.HandleFunc("/api/user/notification/delete/all", mdl.ApiAuth(api.DeleteAllUserNotifications)).Methods("DELETE")
	r.HandleFunc("/api/user/notification/list", mdl.ApiAuth(api.GetNotifications)).Methods("GET")

	// Homescript
	r.HandleFunc("/api/homescript/add", mdl.ApiAuth(mdl.Perm(api.CreateNewHomescript, database.PermissionHomescript))).Methods("POST")
	r.HandleFunc("/api/homescript/modify", mdl.ApiAuth(mdl.Perm(api.ModifyHomescript, database.PermissionHomescript))).Methods("PUT")
	r.HandleFunc("/api/homescript/modify/code", mdl.ApiAuth(mdl.Perm(api.ModifyHomescriptCode, database.PermissionHomescript))).Methods("PUT")
	r.HandleFunc("/api/homescript/delete", mdl.ApiAuth(mdl.Perm(api.DeleteHomescriptById, database.PermissionHomescript))).Methods("DELETE")
	r.HandleFunc("/api/homescript/get/{id}", mdl.ApiAuth(mdl.Perm(api.GetUserHomescriptById, database.PermissionHomescript))).Methods("GET")
	r.HandleFunc("/api/homescript/list/personal", mdl.ApiAuth(mdl.Perm(api.ListPersonalHomescripts, database.PermissionHomescript))).Methods("GET")
	r.HandleFunc("/api/homescript/list/personal/complete", mdl.ApiAuth(mdl.Perm(api.ListPersonalHomescriptsWithArgs, database.PermissionHomescript))).Methods("GET")
	r.HandleFunc("/api/homescript/list/personal/complete", mdl.ApiAuth(mdl.Perm(api.ListPersonalHomescriptsWithArgs, database.PermissionHomescript))).Methods("GET")
	r.HandleFunc("/api/homescript/sources", mdl.ApiAuth(mdl.Perm(api.ListHomescriptSources, database.PermissionHomescript))).Methods("PUT")

	// Homescript Execution And Linting
	r.HandleFunc("/api/homescript/lint", mdl.ApiAuth(mdl.Perm(api.LintHomescriptId, database.PermissionHomescript))).Methods("POST")
	r.HandleFunc("/api/homescript/lint/live", mdl.ApiAuth(mdl.Perm(api.LintHomescriptString, database.PermissionHomescript))).Methods("POST")
	r.HandleFunc("/api/homescript/run", mdl.ApiAuth(mdl.Perm(api.RunHomescriptId, database.PermissionHomescript))).Methods("POST")
	r.HandleFunc("/api/homescript/run/ws", mdl.ApiAuth(mdl.Perm(api.RunHomescriptByIDAsync, database.PermissionHomescript)))
	r.HandleFunc("/api/homescript/run/live", mdl.ApiAuth(mdl.Perm(api.RunHomescriptString, database.PermissionHomescript))).Methods("POST")
	r.HandleFunc("/api/homescript/jobs", mdl.ApiAuth(mdl.Perm(api.GetHMSJobs, database.PermissionHomescript))).Methods("GET")
	r.HandleFunc("/api/homescript/kill/job/{id}", mdl.ApiAuth(mdl.Perm(api.KillJobById, database.PermissionHomescript))).Methods("POST")
	r.HandleFunc("/api/homescript/kill/script/{id}", mdl.ApiAuth(mdl.Perm(api.KillAllHMSIdJobs, database.PermissionHomescript))).Methods("POST")

	// Homescript Arguments
	r.HandleFunc("/api/homescript/arg/add", mdl.ApiAuth(mdl.Perm(api.CreateNewHomescriptArg, database.PermissionHomescript))).Methods("POST")
	r.HandleFunc("/api/homescript/arg/modify", mdl.ApiAuth(mdl.Perm(api.ModifyHomescriptArgument, database.PermissionHomescript))).Methods("PUT")
	r.HandleFunc("/api/homescript/arg/delete", mdl.ApiAuth(mdl.Perm(api.DeleteHomescriptArgument, database.PermissionHomescript))).Methods("DELETE")
	r.HandleFunc("/api/homescript/arg/list/personal", mdl.ApiAuth(mdl.Perm(api.ListUserHomescriptArgs, database.PermissionHomescript))).Methods("GET")
	r.HandleFunc("/api/homescript/arg/list/of/{id}", mdl.ApiAuth(mdl.Perm(api.GetHomescriptArgsByHmsId, database.PermissionHomescript))).Methods("GET")

	// Automations
	r.HandleFunc("/api/automation/list/personal", mdl.ApiAuth(mdl.Perm(api.GetUserAutomations, database.PermissionAutomation))).Methods("GET")
	r.HandleFunc("/api/automation/add", mdl.ApiAuth(mdl.Perm(api.CreateNewAutomation, database.PermissionAutomation))).Methods("POST")
	r.HandleFunc("/api/automation/delete", mdl.ApiAuth(mdl.Perm(api.RemoveAutomation, database.PermissionAutomation))).Methods("DELETE")
	r.HandleFunc("/api/automation/modify", mdl.ApiAuth(mdl.Perm(api.ModifyAutomation, database.PermissionAutomation))).Methods("PUT")

	// Scheduler
	r.HandleFunc("/api/scheduler/list/personal", mdl.ApiAuth(mdl.Perm(api.GetUserSchedules, database.PermissionScheduler))).Methods("GET")
	r.HandleFunc("/api/scheduler/add", mdl.ApiAuth(mdl.Perm(api.CreateNewSchedule, database.PermissionScheduler))).Methods("POST")
	r.HandleFunc("/api/scheduler/delete", mdl.ApiAuth(mdl.Perm(api.RemoveSchedule, database.PermissionScheduler))).Methods("DELETE")
	r.HandleFunc("/api/scheduler/modify", mdl.ApiAuth(mdl.Perm(api.ModifySchedule, database.PermissionScheduler))).Methods("PUT")

	r.HandleFunc("/api/scheduler/state/personal", mdl.ApiAuth(mdl.Perm(api.SetCurrentUserSchedulerEnabled, database.PermissionScheduler))).Methods("PUT")
	r.HandleFunc("/api/scheduler/state/user", mdl.ApiAuth(mdl.Perm(api.SetUserSchedulerEnabled, database.PermissionManageUsers))).Methods("PUT")

	// Reminders
	r.HandleFunc("/api/reminder/add", mdl.ApiAuth(mdl.Perm(api.AddReminder, database.PermissionReminder))).Methods("POST")
	r.HandleFunc("/api/reminder/list", mdl.ApiAuth(mdl.Perm(api.GetReminders, database.PermissionReminder))).Methods("GET")
	r.HandleFunc("/api/reminder/modify", mdl.ApiAuth(mdl.Perm(api.ModifyReminder, database.PermissionReminder))).Methods("PUT")
	r.HandleFunc("/api/reminder/delete", mdl.ApiAuth(mdl.Perm(api.DeleteReminder, database.PermissionReminder))).Methods("DELETE")

	// Weather
	r.HandleFunc("/api/weather/key/modify", mdl.ApiAuth(mdl.Perm(api.UpdateOpenWeatherMapApiKey, database.PermissionSystemConfig))).Methods("PUT")
	r.HandleFunc("/api/weather", mdl.ApiAuth(api.GetWeather)).Methods("GET")
	r.HandleFunc("/api/weather/cached", mdl.ApiAuth(api.GetCachedWeather)).Methods("GET")

	// Cache Purging
	r.HandleFunc("/api/weather/cache", mdl.ApiAuth(mdl.Perm(api.PurgeWeatherCache, database.PermissionSystemConfig))).Methods("DELETE")
	// TODO: what is up with cache purging?
	// r.HandleFunc("/api/power/cache", mdl.ApiAuth(mdl.Perm(api.PurgePowerRecords, database.PermissionSystemConfig))).Methods("DELETE")

	// System Configuration
	r.HandleFunc("/api/automation/state/global", mdl.ApiAuth(mdl.Perm(api.ChangeActivationAutomation, database.PermissionSystemConfig))).Methods("PUT")

	r.HandleFunc("/api/system/config", mdl.ApiAuth(mdl.Perm(api.GetSystemConfig, database.PermissionSystemConfig))).Methods("GET")
	r.HandleFunc("/api/system/location/modify", mdl.ApiAuth(mdl.Perm(api.UpdateLocation, database.PermissionSystemConfig))).Methods("PUT")
	r.HandleFunc("/api/system/location/suntimes", mdl.ApiAuth(api.GetSunTimes)).Methods("GET")
	r.HandleFunc("/api/system/lockdown/modify", mdl.ApiAuth(mdl.Perm(api.UpdateLockDownMode, database.PermissionSystemConfig))).Methods("PUT")
	r.HandleFunc("/api/system/config/export", mdl.ApiAuth(mdl.Perm(api.ExportConfiguration, database.PermissionSystemConfig))).Methods("POST")
	r.HandleFunc("/api/system/config/import", mdl.ApiAuth(mdl.Perm(api.ImportConfiguration, database.PermissionSystemConfig))).Methods("POST")
	r.HandleFunc("/api/system/config/factory", mdl.ApiAuth(mdl.Perm(api.FactoryReset, database.PermissionSystemConfig))).Methods("DELETE")

	// Hardware node management
	// r.HandleFunc("/api/system/hardware/node/list", mdl.ApiAuth(mdl.Perm(api.ListHardwareNodes, database.PermissionSystemConfig))).Methods("GET")
	// r.HandleFunc("/api/system/hardware/node/list/nopriv", mdl.ApiAuth(mdl.Perm(api.ListHardwareNodesNoPriv, database.PermissionPower))).Methods("GET")
	// r.HandleFunc("/api/system/hardware/node/check", mdl.ApiAuth(mdl.Perm(api.ListHardwareNodesWithCheck, database.PermissionSystemConfig))).Methods("GET")
	// r.HandleFunc("/api/system/hardware/node/add", mdl.ApiAuth(mdl.Perm(api.CreateHardwareNode, database.PermissionSystemConfig))).Methods("POST")
	// r.HandleFunc("/api/system/hardware/node/modify", mdl.ApiAuth(mdl.Perm(api.ModifyHardwareNode, database.PermissionSystemConfig))).Methods("PUT")
	// r.HandleFunc("/api/system/hardware/node/delete", mdl.ApiAuth(mdl.Perm(api.DeleteHardwareNode, database.PermissionSystemConfig))).Methods("DELETE")
	// TODO: what to do with hardware nodes?

	// Hardware driver management
	r.HandleFunc("/api/system/hardware/driver/list", mdl.ApiAuth(mdl.Perm(api.ListDeviceDrivers, database.PermissionSystemConfig))).Methods("GET")
	r.HandleFunc("/api/system/hardware/driver/add", mdl.ApiAuth(mdl.Perm(api.CreateDeviceDriver, database.PermissionSystemConfig))).Methods("POST")
	r.HandleFunc("/api/system/hardware/driver/modify", mdl.ApiAuth(mdl.Perm(api.ModifyDeviceDriver, database.PermissionSystemConfig))).Methods("PUT")
	r.HandleFunc("/api/system/hardware/driver/configure", mdl.ApiAuth(mdl.Perm(api.ConfigureDeviceDriver, database.PermissionSystemConfig))).Methods("PUT")
	r.HandleFunc("/api/system/hardware/driver/delete", mdl.ApiAuth(mdl.Perm(api.DeleteDeviceDriver, database.PermissionSystemConfig))).Methods("DELETE")
	// TODO: add driver support

	// Logging
	r.HandleFunc("/api/logs/delete/old", mdl.ApiAuth(mdl.Perm(api.FlushOldLogs, database.PermissionSystemConfig))).Methods("DELETE")
	r.HandleFunc("/api/logs/delete/all", mdl.ApiAuth(mdl.Perm(api.FlushAllLogs, database.PermissionSystemConfig))).Methods("DELETE")
	r.HandleFunc("/api/logs/delete/id/{id}", mdl.ApiAuth(mdl.Perm(api.DeleteLogById, database.PermissionSystemConfig))).Methods("DELETE")
	r.HandleFunc("/api/logs/list/all", mdl.ApiAuth(mdl.Perm(api.ListLogs, database.PermissionSystemConfig))).Methods("GET")

	// Assets & static files
	assetsFilepath := "./web/dist/assets/"
	assetsPathPrefix := "/assets"
	assetsFileserver := http.FileServer(http.Dir(assetsFilepath))
	r.PathPrefix(assetsPathPrefix).Handler(http.StripPrefix(assetsPathPrefix, assetsFileserver))

	// Error handlers
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	log.Debug("Successfully initialized router")
	return r
}
