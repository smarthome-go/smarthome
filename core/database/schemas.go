package database

import "time"

// Identified by a username, has a password and an avatar path
type User struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	AvatarPath string `json:"avatarPath"`
	// TODO: add bg image, frontend themes and colors
}

// Identified by a Switch Id, has a name and belongs to a room
type Switch struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	RoomId string `json:"roomId"`
}

//Contains the switch id and a matching boolean
// Used when requesting global power states
type PowerState struct {
	SwitchId string `json:"switch"`
	PowerOn  bool   `json:"powerOn"`
}

// Identified by a unique Id, has a Name and Description
// When used in config file, the Switches slice is also populated
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
