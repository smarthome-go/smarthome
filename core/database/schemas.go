package database

import "time"

// Used in user.go
type User struct {
	Username string
	Password string
}

// Permission-related

type Permission struct {
	Permission  string
	Name        string
	Description string
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

type Room struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Switches    []Switch `json:"switches"`
}

// Loggin
type LogEvent struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Level       int       `json:"level"`
	Date        time.Time `json:"date"`
}
