package database

import (
	"time"
)

// Identified by a username, has a password and an avatar path
type User struct {
	Username     string `json:"username"`
	Firstname    string `json:"firstname"`
	Surname      string `json:"surname"`
	PrimaryColor string `json:"primaryColor"`
	Password     string `json:"password"`
	AvatarPath   string `json:"avatarPath"`
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
	Cameras     []Camera `json:"cameras"`
}

// Camera struct, used in `config.rooms.cameras``
type Camera struct {
	Id     int    `json:"id"`
	RoomId string `json:"roomId"`
	Url    string `json:"url"`
	Name   string `json:"name"`
}

// internal logging-related
type LogEvent struct {
	Id          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Level       int       `json:"level"`
	Date        time.Time `json:"date"`
}

// User notification
type Notification struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Date        time.Time
	// Username is left out due to not being required in the service layer
}

// Hardware node
type HardwareNode struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Token string `json:"token"`
}
