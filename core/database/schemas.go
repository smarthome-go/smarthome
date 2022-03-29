package database

import (
	"time"
)

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
	Id          uint      `json:"id"`
	Priority    uint8     `json:"priority"` // Includes 1: info, 2: warning, 3: alert
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	// Username is left out due to not being required in the service layer
}
