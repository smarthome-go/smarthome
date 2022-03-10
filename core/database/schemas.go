package database

import "time"

// Used in user.go
type User struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	AvatarPath string `json:"avatarPath"`
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
