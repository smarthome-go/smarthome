package database

// This file defines which permissions exists and describes their attributes
type Permission struct {
	Permission  PermissionType `json:"permission"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
}

type PermissionType string

// Different types of permissions
const (
	PermissionPower             PermissionType = "setPower"
	PermissionViewCameras       PermissionType = "viewCameras"
	PermissionManageUsers       PermissionType = "manageUsers"
	PermissionDebug             PermissionType = "debug"
	PermissionLogging           PermissionType = "logging"
	PermissionAutomation        PermissionType = "automation"
	PermissionScheduler         PermissionType = "scheduler"
	PermissionReminder          PermissionType = "reminder"
	PermissionSystemConfig      PermissionType = "modifyServerConfig"
	PermissionModifyRooms       PermissionType = "modifyRooms"
	PermissionHomescript        PermissionType = "homescript"
	PermissionHomescriptNetwork PermissionType = "hmsNetwork"
	PermissionWildCard          PermissionType = "*"
)

var (
	Permissions = []Permission{
		{
			// User is allowed to request power jobs, interact with outlets, still dependent on switch permissions
			Permission:  PermissionPower,
			Name:        "Power",
			Description: "Interact with switches",
		},
		{
			// (Admin) is allowed to add logs to the internal logging system
			Permission:  PermissionLogging,
			Name:        "Event Logs",
			Description: "Add records to the internal logging system",
		},
		{
			// (Admin) is allowed to read debug information from the server
			Permission:  PermissionDebug,
			Name:        "Debug Features",
			Description: "Obtain debug information about Smarthome",
		},
		{
			// (Admin) is allowed to create new users or delete users and manage their permissions
			Permission:  PermissionManageUsers,
			Name:        "Manage Users",
			Description: "Manage users and their permissions",
		},
		{
			// User is allowed to run, add, delete, and modify Homescript, scheduler homescript excluded
			Permission:  PermissionHomescript,
			Name:        "Homescript",
			Description: "Use the Homescript scripting language",
		},
		{
			// User is allowed to make network requests from Homescript
			Permission:  PermissionHomescriptNetwork,
			Name:        "HMS Network",
			Description: "Perform network requests from Homescript",
		},
		{
			// User is allowed to set up, modify, delete, and view personal automations
			Permission:  PermissionAutomation,
			Name:        "Automations",
			Description: "Use the automation app",
		},
		{
			// User is allowed to set up, modify, delete, and view personal schedules
			Permission:  PermissionScheduler,
			Name:        "Scheduler",
			Description: "Use the scheduler app",
		},
		{
			// User is allowed to set up, modify, delete, and view personal reminders
			Permission:  PermissionReminder,
			Name:        "Reminders",
			Description: "Use the reminder app",
		},
		{
			// (Admin) is allowed to modify rooms, switches and cameras
			Permission:  PermissionModifyRooms,
			Name:        "Manage Rooms",
			Description: "Modify rooms, switches and cameras. Also grants access to every switch",
		},
		{
			// User is allowed to view the video feed of cameras to which he has access
			Permission:  PermissionViewCameras,
			Name:        "View Cameras",
			Description: "View camera image feeds (depends on camera-permissions)",
		},
		{
			// (Admin) is allowed to change global config parameters
			Permission:  PermissionSystemConfig,
			Name:        "System Config",
			Description: "Manage and export system configuration (includes sensitive data)",
		},
		{
			// WARNING: This allows a user to do everything, should only be allowed to admin users
			Permission:  PermissionWildCard,
			Name:        "Permission Wildcard",
			Description: "Allows all permissions",
		},
	}
)
