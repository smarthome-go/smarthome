package database

// Permission-related
type Permission struct {
	Permission  string `json:"permission"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func GetPermissions() []Permission {
	permissions := []Permission{
		{
			// If the user is allowed to authenticate and login, if disabled, a user is `disabled`
			Permission:  "authentication",
			Name:        "Authentication",
			Description: "Allows the user to authenticate",
		},
		{
			Permission:  "getUserSwitches",
			Name:        "Get Personal Switches",
			Description: "Get all allowed switches for current user.",
		},
		{
			Permission:  "setPower",
			Name:        "Set Power",
			Description: "Interact with switches",
		},
		{
			Permission:  "deleteOldLogs",
			Name:        "Flush Old Logs",
			Description: "Delete logs events which are older than 30 days",
		},
		{
			Permission:  "deleteAllLogs",
			Name:        "Flush All Logs",
			Description: "Delete all log event records",
		},
		{
			Permission:  "listLogs",
			Name:        "List Logs",
			Description: "List all internal logs",
		},
		{
			Permission:  "uploadAvatar",
			Name:        "Upload Avatar",
			Description: "Allows the user to upload a customized avatar",
		},
		{
			Permission:  "deleteAvatar",
			Name:        "Delete Avatar",
			Description: "Allows the user to delete their customized avatar",
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
