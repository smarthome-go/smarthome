package config

import (
	"fmt"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
)

type SetupStruct struct {
	Users               []database.FullUser   `json:"users"`
	Rooms               []database.Room       `json:"rooms"`
	HardwareNodes       []setupHardwareNode   `json:"hardwareNodes"`
	ServerConfiguration database.ServerConfig `json:"serverConfiguration"`
}

type setupUser struct {
	User        setupUserData             `json:"user"`
	Homescripts []setupHomescript         `json:"homescripts"`
	Automations []database.AutomationData `json:"automations"`
	Reminders   []setupReminder           `json:"reminders"`

	// Permissions
	Permissions       []database.PermissionType `json:"permissions"`
	SwitchPermissions []string                  `json:"switchPermissions"`
	CameraPermissions []string                  `json:"cameraPermissions"`
}

type setupHardwareNode struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"` // Can be used to temporarely deactivate a node in case of maintenance
	Url     string `json:"url"`
	Token   string `json:"token"`
}

type setupReminder struct {
	Name              string                        `json:"name"`
	Description       string                        `json:"description"`
	Priority          database.NotificationPriority `json:"priority"`
	CreatedDate       time.Time                     `json:"createdDate"`
	DueDate           time.Time                     `json:"dueDate"`
	UserWasNotified   bool                          `json:"userWasNotified"`
	UserWasNotifiedAt time.Time                     `json:"userWasNotifiedAt"`
}

type setupHomescript struct {
	Data      database.HomescriptData      `json:"data"`
	Arguments []database.HomescriptArgData `json:"arguments"`
}

type setupUserData struct {
	Username          string `json:"username"`
	Forename          string `json:"forename"`
	Surname           string `json:"surname"`
	PrimaryColorDark  string `json:"primaryColorDark"`
	PrimaryColorLight string `json:"primaryColorLight"`
	Password          string `json:"password"`
	SchedulerEnabled  bool   `json:"schedulerEnabled"`
	DarkTheme         bool   `json:"darkTheme"`
}

func Export() (SetupStruct, error) {
	// Server configuration
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil {
		return SetupStruct{}, err
	}
	if !found {
		return SetupStruct{}, fmt.Errorf("No configuration could be found")
	}
	// Rooms configuration
	rooms, err := database.ListAllRoomsWithData()
	if err != nil {
		return SetupStruct{}, err
	}
	hwNodes, err := database.GetHardwareNodes()
	if err != nil {
		return SetupStruct{}, err
	}
	hwNodesNew := make([]setupHardwareNode, 0)
	for _, node := range hwNodes {
		hwNodesNew = append(hwNodesNew, setupHardwareNode{
			Name:    node.Name,
			Enabled: node.Enabled,
			Token:   node.Token,
			Url:     node.Url,
		})
	}
	// TODO: add users

	return SetupStruct{
		ServerConfiguration: serverConfig,
		Rooms:               rooms,
	}, nil
}
