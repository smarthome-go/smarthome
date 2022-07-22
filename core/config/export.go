package config

import (
	"fmt"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type SetupStruct struct {
	Users               []setupUser           `json:"users"`
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
	Permissions       []string `json:"permissions"`
	SwitchPermissions []string `json:"switchPermissions"`
	CameraPermissions []string `json:"cameraPermissions"`
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

	usersTemp, err := database.ListUsers()
	if err != nil {
		return SetupStruct{}, nil
	}
	// TODO: add users
	users := make([]setupUser, 0)

	for _, user := range usersTemp {
		// Password Hash
		pwHash, err := database.GetUserPasswordHash(user.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Homescripts
		homescriptsDB, err := homescript.ListPersonalHomescriptWithArgs(user.Username)
		if err != nil {
			return SetupStruct{}, err
		}
		homescripts := make([]setupHomescript, 0)
		for _, hms := range homescriptsDB {
			args := make([]database.HomescriptArgData, 0)
			for _, arg := range hms.Arguments {
				args = append(args, arg.Data)
			}
			homescripts = append(homescripts, setupHomescript{
				Data:      hms.Data.Data,
				Arguments: args,
			})
		}

		// Automations
		automationsDB, err := database.GetUserAutomations(user.Username)
		if err != nil {
			return SetupStruct{}, err
		}
		automations := make([]database.AutomationData, 0)
		for _, automation := range automationsDB {
			automations = append(automations, automation.Data)
		}

		// Reminders
		remindersDB, err := database.GetUserReminders(user.Username)
		if err != nil {
			return SetupStruct{}, err
		}
		reminders := make([]setupReminder, 0)
		for _, reminder := range remindersDB {
			reminders = append(reminders, setupReminder{
				Name:              reminder.Name,
				Description:       reminder.Description,
				Priority:          reminder.Priority,
				CreatedDate:       reminder.CreatedDate,
				DueDate:           reminder.DueDate,
				UserWasNotified:   reminder.UserWasNotified,
				UserWasNotifiedAt: reminder.UserWasNotifiedAt,
			})
		}

		// Permissions
		permissions, err := database.GetUserPermissions(user.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Switch p0ermissionws
		swPermissions, err := database.GetUserSwitchPermissions(user.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Camera permissions
		camPermissions, err := database.GetUserCameraPermissions(user.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Append assembled user
		users = append(users, setupUser{
			User: setupUserData{
				Username:          user.Username,
				Forename:          user.Forename,
				Surname:           user.Surname,
				PrimaryColorDark:  user.PrimaryColorDark,
				PrimaryColorLight: user.PrimaryColorLight,
				Password:          pwHash,
				SchedulerEnabled:  user.SchedulerEnabled,
				DarkTheme:         user.DarkTheme,
			},
			Homescripts:       homescripts,
			Automations:       automations,
			Reminders:         reminders,
			Permissions:       permissions,
			SwitchPermissions: swPermissions,
			CameraPermissions: camPermissions,
		})
	}

	return SetupStruct{
		Users:               users,
		Rooms:               rooms,
		HardwareNodes:       hwNodesNew,
		ServerConfiguration: serverConfig,
	}, nil
}
