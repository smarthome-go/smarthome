package config

import (
	"fmt"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type SetupStruct struct {
	Users               []setupUser           `json:"users"`
	Rooms               []setupRoom           `json:"rooms"`
	HardwareNodes       []setupHardwareNode   `json:"hardwareNodes"`
	ServerConfiguration database.ServerConfig `json:"serverConfiguration"`
}

type setupRoom struct {
	Data     database.RoomData `json:"data"`
	Switches []setupSwitch     `json:"switches"`
	Cameras  []setupCamera     `json:"cameras"`
}

type setupSwitch struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	PowerOn bool   `json:"powerOn"`
	Watts   uint16 `json:"watts"`
}

type setupCamera struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type setupUser struct {
	User              setupUserData            `json:"user"`
	Tokens            []setupAuthToken         `json:"tokens"`
	Homescripts       []setupHomescript        `json:"homescripts"`
	HomescriptStorage []setupHomescriptStorage `json:"homescriptStorage"`
	Reminders         []setupReminder          `json:"reminders"`

	// Permissions
	Permissions       []string `json:"permissions"`
	SwitchPermissions []string `json:"switchPermissions"`
	CameraPermissions []string `json:"cameraPermissions"`
}

type setupHardwareNode struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"` // Can be used to temporarily deactivate a node in case of maintenance
	Url     string `json:"url"`
	Token   string `json:"token"`
}

type setupAuthToken struct {
	Token string `json:"token"`
	Label string `json:"label"`
}

type setupReminder struct {
	Name              string                    `json:"name"`
	Description       string                    `json:"description"`
	Priority          database.ReminderPriority `json:"priority"`
	CreatedDate       time.Time                 `json:"createdDate"`
	DueDate           time.Time                 `json:"dueDate"`
	UserWasNotified   bool                      `json:"userWasNotified"`
	UserWasNotifiedAt time.Time                 `json:"userWasNotifiedAt"`
}

type setupHomescript struct {
	Data        database.HomescriptData `json:"data"`
	Arguments   []setupHomescriptArg    `json:"arguments"`
	Automations []setupAutomation       `json:"automations"`
}

type setupHomescriptArg struct {
	ArgKey    string                   `json:"argKey"`    // The unique key of the argument
	Prompt    string                   `json:"prompt"`    // The prompt the user will see
	MDIcon    string                   `json:"mdIcon"`    // A Google MD icon which will be displayed
	InputType database.HmsArgInputType `json:"inputType"` // Specifies the expected data type
	Display   database.HmsArgDisplay   `json:"display"`   // Specifies the visual display of the prompt (handled by GUI)
}

type setupHomescriptStorage struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type setupAutomation struct {
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	CronExpression string              `json:"cronExpression"`
	Enabled        bool                `json:"enabled"`
	TimingMode     database.TimingMode `json:"timingMode"`
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
	roomsDB, err := database.ListAllRoomsWithData()
	if err != nil {
		return SetupStruct{}, err
	}
	rooms := make([]setupRoom, 0)
	for _, room := range roomsDB {
		roomSwitches := make([]setupSwitch, 0)
		for _, sw := range room.Switches {
			roomSwitches = append(roomSwitches, setupSwitch{
				Id:      sw.Id,
				Name:    sw.Name,
				PowerOn: sw.PowerOn,
				Watts:   sw.Watts,
			})
		}
		roomCameras := make([]setupCamera, 0)
		for _, cam := range room.Cameras {
			roomCameras = append(roomCameras, setupCamera{
				Id:   cam.Id,
				Name: cam.Name,
				Url:  cam.Url,
			})
		}
		rooms = append(rooms, setupRoom{
			Data: database.RoomData{
				Id:          room.Data.Id,
				Name:        room.Data.Name,
				Description: room.Data.Description,
			},
			Switches: roomSwitches,
			Cameras:  roomCameras,
		})
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
	users := make([]setupUser, 0)
	for _, user := range usersTemp {
		// Password Hash
		pwHash, err := database.GetUserPasswordHash(user.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Authentication tokens
		tokensDB, err := database.GetUserTokensOfUser(user.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Transform the tokens into a setup-compatible version
		tokens := make([]setupAuthToken, 0)
		for _, t := range tokensDB {
			tokens = append(tokens, setupAuthToken{
				Token: t.Token,
				Label: t.Data.Label,
			})
		}

		// Automations
		automationsDB, err := database.GetUserAutomations(user.Username)
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
			args := make([]setupHomescriptArg, 0)
			for _, arg := range hms.Arguments {
				args = append(args, setupHomescriptArg{
					ArgKey:    arg.Data.ArgKey,
					Prompt:    arg.Data.Prompt,
					MDIcon:    arg.Data.MDIcon,
					InputType: arg.Data.InputType,
					Display:   arg.Data.Display,
				})
			}
			// Filter out automations using this Homescript
			automationsThis := make([]setupAutomation, 0)
			for _, aut := range automationsDB {
				if aut.Data.HomescriptId == hms.Data.Data.Id {
					automationsThis = append(automationsThis, setupAutomation{
						Name:           aut.Data.Name,
						Description:    aut.Data.Description,
						CronExpression: aut.Data.CronExpression,
						Enabled:        aut.Data.Enabled,
						TimingMode:     aut.Data.TimingMode,
					})
				}
			}
			homescripts = append(homescripts, setupHomescript{
				Data:        hms.Data.Data,
				Arguments:   args,
				Automations: automationsThis,
			})
		}

		// Homescript storage
		storage, err := database.GetPersonalHomescriptStorage(user.Username)
		if err != nil {
			return SetupStruct{}, nil
		}

		storageOutput := make([]setupHomescriptStorage, 0)
		for key, value := range storage {
			storageOutput = append(storageOutput, setupHomescriptStorage{
				Key:   key,
				Value: value,
			})
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
			Tokens:            tokens,
			Homescripts:       homescripts,
			HomescriptStorage: storageOutput,
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
