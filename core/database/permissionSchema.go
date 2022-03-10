package database

// This file defines which permissions exists and describes their attributes
type Permission struct {
	Permission  string `json:"permission"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func GetPermissions() []Permission {
	permissions := []Permission{
		{
			// User is allowed to authenticate and login, if disabled, a user is `disabled`
			Permission:  "authentication",
			Name:        "Authentication",
			Description: "Allows the user to authenticate",
		},
		{
			// User is allowed to request a list of their personal switches (which they have access to)
			Permission:  "getUserSwitches",
			Name:        "Get Personal Switches",
			Description: "Get all allowed switches for current user.",
		},
		{
			// User is allowed to request power jobs, interact with outlets, still dependent on switch permissions
			Permission:  "setPower",
			Name:        "Set Power",
			Description: "Interact with switches",
		},
		{
			// (Admin) is allowed to delete logs older than 30 days
			Permission:  "deleteOldLogs",
			Name:        "Flush Old Logs",
			Description: "Delete logs events which are older than 30 days",
		},
		{
			// (Admin) is allowed to delete all logs
			Permission:  "deleteAllLogs",
			Name:        "Flush All Logs",
			Description: "Delete all log event records",
		},
		{
			// (Admin) is allowed to request logs
			Permission:  "listLogs",
			Name:        "List Logs",
			Description: "List all internal logs",
		},
		{
			// User is allowed tp upload a custom avatar
			Permission:  "changeAvatar",
			Name:        "Upload / Delete / Change Avatar",
			Description: "Allows the user to customize their avatar",
		},
		{
			Permission:  "addUserPermission",
			Name:        "Add Permission to user",
			Description: "Adds a given permission to a given user",
		},
		{
			Permission:  "removeUserPermission",
			Name:        "Remove Permission from User",
			Description: "Removes a given permission from a user",
		},
		{
			Permission:  "addSwitchPermission",
			Name:        "Add Switch Permission",
			Description: "Add a switch permission to a user",
		},
		{
			Permission:  "removeSwitchPermission",
			Name:        "Remove Switch Permission",
			Description: "Removes a given switch permission from a user",
		},
		{
			Permission:  "getDebugInfo",
			Name:        "Get Debug Info",
			Description: "Obtain debug information about the system",
		},
		{
			Permission:  "addUser",
			Name:        "Add User",
			Description: "Create a new user",
		},
		{
			Permission:  "removeUser",
			Name:        "Delete User",
			Description: "Delete a given user"},
		{
			Permission:  "*",
			Name:        "Permission Wildcard *",
			Description: "Allows all permissions",
		},
	}
	return permissions
}
