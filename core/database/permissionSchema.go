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
	PermissionAuthentication          PermissionType = "authentication"
	PermissionSetPower                PermissionType = "setPower"
	PermissionAddLogEvent             PermissionType = "addLogEvent"
	PermissionDeleteLogs              PermissionType = "deleteLogs"
	PermissionListLogs                PermissionType = "listLogs"
	PermissionChangeAvatar            PermissionType = "changeAvatar"
	PermissionChangeUserPermissions   PermissionType = "changeUserPermissions"
	PermissionChangeSwitchPermissions PermissionType = "changeSwitchPermissions"
	PermissionGetDebugInfo            PermissionType = "getDebugInfo"
	PermissionChangeUsers             PermissionType = "changeUsers"
	PermissionListUsers               PermissionType = "listUsers"
	PermissionRunHomescript           PermissionType = "runHomescript"

	// Dangerous
	PermissionWildCard PermissionType = "*"
)

var (
	Permissions = []Permission{
		{
			// User is allowed to authenticate and login, if disabled, a user is `disabled`
			Permission:  PermissionAuthentication,
			Name:        "Authentication",
			Description: "Allows the user to authenticate",
		},
		{
			// User is allowed to request power jobs, interact with outlets, still dependent on switch permissions
			Permission:  PermissionSetPower,
			Name:        "Set Power",
			Description: "Interact with switches",
		},
		{
			// (Admin) is allowed to use the internal logging system
			Permission:  PermissionAddLogEvent,
			Name:        "Add Log Event",
			Description: "Use the internal logging system",
		},
		{
			// (Admin) is allowed to delete logs older than 30 days
			Permission:  PermissionDeleteLogs,
			Name:        "Flush Logs All or Old",
			Description: "Delete logs events which are older than 30 days or delete all log events",
		},
		{
			// (Admin) is allowed to request logs
			Permission:  PermissionListLogs,
			Name:        "List Logs",
			Description: "List all internal logs",
		},
		{
			// User is allowed to upload a custom avatar
			Permission:  PermissionChangeAvatar,
			Name:        "Upload / Delete / Change Avatar",
			Description: "Allows the user to customize their avatar",
		},
		{
			// (Admin) is allowed to add / delete permissions to / from users
			Permission:  PermissionChangeUserPermissions,
			Name:        "Change User Permissions",
			Description: "Add / delete permissions to / from users",
		},
		{
			// (Admin) is allowed to add / delete switch permissions to / from users
			Permission:  PermissionChangeSwitchPermissions,
			Name:        "Change User Switch Permissions",
			Description: "Add / delete switch permissions to / from users",
		},
		{
			// (Admin) is allowed to read debug information from the server
			Permission:  PermissionGetDebugInfo,
			Name:        "Display Debug Info",
			Description: "Obtain debug information about the system",
		},
		{
			// (Admin) is allowed to create new users or delete users
			Permission:  PermissionChangeUsers,
			Name:        "Add / Delete users",
			Description: "Create a new user or delete users",
		},
		{
			// (Admin) is allowed to list all users
			Permission:  PermissionListUsers,
			Name:        "List users",
			Description: "See a list of all users",
		},
		{
			// User is allowed to run Homescript, scheduler homescript excluded
			Permission:  PermissionRunHomescript,
			Name:        "Run Homescript",
			Description: "Run predefined Homescript files or send live code to be executed by the server",
		},
		{
			// WARNING: This allows a user to do everything, should only be allowed to the `admin` user
			Permission:  PermissionWildCard,
			Name:        "Permission Wildcard *",
			Description: "Allows all permissions",
		},
	}
)
