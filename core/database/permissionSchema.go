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
	PermissionPower              PermissionType = "setPower"
	PermissionLogs               PermissionType = "logs"
	PermissionManageUsers        PermissionType = "manageUsers"
	PermissionDebug              PermissionType = "debug"
	PermissionHomescript         PermissionType = "homescript"
	PermissionAutomation         PermissionType = "automation"
	PermissionScheduler          PermissionType = "scheduler"
	PermissionReminder           PermissionType = "reminder"
	PermissionModifyServerConfig PermissionType = "modifyServerConfig"
	PermissionModifyRooms        PermissionType = "modifyRooms"

	// Dangerous
	PermissionWildCard PermissionType = "*"
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
			// (Admin) is allowed to use and manage the internal logging system
			Permission:  PermissionLogs,
			Name:        "Manage Logging",
			Description: "Use and manage the internal logging system",
		},
		{
			// (Admin) is allowed to read debug information from the server
			Permission:  PermissionDebug,
			Name:        "Debug Features",
			Description: "Obtain debug information about the system",
		},
		{
			// (Admin) is allowed to create new users or delete users and manage their permissions
			Permission:  PermissionManageUsers,
			Name:        "Manage Users",
			Description: "Create / remove and manage users and manage their permissions",
		},
		{
			// User is allowed to run, add, delete, and modify Homescript, scheduler homescript excluded
			Permission:  PermissionHomescript,
			Name:        "Use Homescript",
			Description: "List, add, delete, run, and modify Homescripts",
		},
		{
			// User is allowed to set up, modify, delete, and view personal automations
			Permission:  PermissionAutomation,
			Name:        "Use Automations",
			Description: "List, add, delete, and modify automations",
		},
		{
			// User is allowed to set up, modify, delete, and view personal schedules
			Permission:  PermissionScheduler,
			Name:        "Use the Scheduler",
			Description: "List, add, delete, and modify schedules",
		},
		{
			// User is allowed to set up, modify, delete, and view personal reminders
			Permission:  PermissionReminder,
			Name:        "Use Reminders",
			Description: "List, add, delete, and modify reminders",
		},
		{
			// (Admin) is allowed to modify rooms, switches and cameras
			Permission:  PermissionModifyRooms,
			Name:        "Manage Rooms",
			Description: "Add, modify and delete rooms and room contents",
		},
		{
			// (Admin) is allowed to change global config parameters
			Permission:  PermissionModifyServerConfig,
			Name:        "Manage Server Config",
			Description: "Change global server configuration values",
		},
		{
			// WARNING: This allows a user to do everything, should only be allowed to admin users
			Permission:  PermissionWildCard,
			Name:        "Permission Wildcard *",
			Description: "Allows all permissions",
		},
	}
)
