package database

import "time"

// Used in user.go
type User struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	AvatarPath string `json:"avatarPath"`
}

// Permission-related
type Permission struct {
	Permission  string `json:"permission"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TODO: Move to separate permissionsSchema.go file
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
			Permission:  "*",
			Name:        "Permission WIldcard *",
			Description: "Allows all permissions",
		},
	}
	return permissions
}

// Rooms and Switches
type Switch struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	RoomId string `json:"roomId"`
}

type PowerState struct {
	SwitchId string `json:"switch"`
	PowerOn  bool   `json:"powerOn"`
}

// TODO: add documentation comments
type Room struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Switches    []Switch `json:"switches"`
}

// internal logging-related
type LogEvent struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Level       int       `json:"level"`
	Date        time.Time `json:"date"`
}
