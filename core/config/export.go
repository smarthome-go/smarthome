package config

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/h2non/filetype"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/core/user"
)

type SetupStruct struct {
	Users               []setupUser           `json:"users"`
	Rooms               []setupRoom           `json:"rooms"`
	HardwareNodes       []setupHardwareNode   `json:"hardwareNodes"`
	ServerConfiguration database.ServerConfig `json:"serverConfiguration"`
	CacheData           setupCacheData        `json:"cacheData"`
}

type setupRoom struct {
	Data     database.RoomData `json:"data"`
	Switches []setupSwitch     `json:"switches"`
	Cameras  []setupCamera     `json:"cameras"`
}

type setupSwitch struct {
	Id         string  `json:"id"`
	Name       string  `json:"name"`
	PowerOn    bool    `json:"powerOn"`
	Watts      uint16  `json:"watts"`
	TargetNode *string `json:"targetNode"`
}

type setupCamera struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type setupUser struct {
	Data              setupUserData            `json:"user"`
	Tokens            []setupAuthToken         `json:"tokens"`
	Homescripts       []setupHomescript        `json:"homescripts"`
	HomescriptStorage []setupHomescriptStorage `json:"homescriptStorage"`
	Reminders         []setupReminder          `json:"reminders"`

	// Profile picture as B64
	ProfilePicture *setupUserProfilePicture `json:"profilePicture"`

	// Permissions
	Permissions       []string `json:"permissions"`
	SwitchPermissions []string `json:"switchPermissions"`
	CameraPermissions []string `json:"cameraPermissions"`
}

type setupUserProfilePicture struct {
	B64Data       string `json:"data"`
	FileExtension string `json:"fileExtension"`
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

type setupCacheData struct {
	WeatherHistory []setupWeatherMeasurement               `json:"weatherHistory"`
	PowerData      []hardware.PowerDrawDataPointUnixMillis `json:"powerData"`
}

type setupWeatherMeasurement struct {
	Id                 uint    `json:"id"`
	Time               uint    `json:"time"` // unix millis
	WeatherTitle       string  `json:"weatherTitle"`
	WeatherDescription string  `json:"weatherDescription"`
	Temperature        float32 `json:"temperature"`
	FeelsLike          float32 `json:"feelsLike"`
	Humidity           uint8   `json:"humidity"`
}

func Export(
	includeProfilePictures bool, // The user's profile pictures in base64
	includedCacheData bool, // Weather history and power data
) (SetupStruct, error) {
	// Server configuration
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil {
		return SetupStruct{}, err
	}
	if !found {
		return SetupStruct{}, fmt.Errorf("No configuration could be found")
	}
	// Rooms configuration
	roomsDB, err := database.ListAllRoomsWithData(false) // camera URLs shall not be redacted
	if err != nil {
		return SetupStruct{}, err
	}
	rooms := make([]setupRoom, 0)
	for _, room := range roomsDB {
		roomSwitches := make([]setupSwitch, 0)
		for _, sw := range room.Switches {
			roomSwitches = append(roomSwitches, setupSwitch{
				Id:         sw.Id,
				Name:       sw.Name,
				PowerOn:    sw.PowerOn,
				Watts:      sw.Watts,
				TargetNode: sw.TargetNode,
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
	for _, userData := range usersTemp {
		// Password Hash
		pwHash, err := database.GetUserPasswordHash(userData.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Authentication tokens
		tokensDB, err := database.GetUserTokensOfUser(userData.Username)
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
		automationsDB, err := database.GetUserAutomations(userData.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Homescripts
		homescriptsDB, err := homescript.ListPersonalHomescriptWithArgs(userData.Username)
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
		storage, err := database.GetPersonalHomescriptStorage(userData.Username)
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
		remindersDB, err := database.GetUserReminders(userData.Username)
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
		permissions, err := database.GetUserPermissions(userData.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Switch p0ermissionws
		swPermissions, err := database.GetUserSwitchPermissions(userData.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Camera permissions
		camPermissions, err := database.GetUserCameraPermissions(userData.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Include profile picture if desired
		var profilePicture *setupUserProfilePicture = nil
		if includeProfilePictures {
			picture, err := user.GetUserAvatar(userData.Username)
			if err != nil {
				return SetupStruct{}, err
			}

			filetype, err := filetype.Match(picture)
			if err != nil {
				return SetupStruct{}, err
			}

			picTemp := setupUserProfilePicture{
				B64Data:       base64.StdEncoding.EncodeToString(picture),
				FileExtension: filetype.Extension,
			}

			profilePicture = &picTemp
		}

		// Append assembled user
		users = append(users, setupUser{
			Data: setupUserData{
				Username:          userData.Username,
				Forename:          userData.Forename,
				Surname:           userData.Surname,
				PrimaryColorDark:  userData.PrimaryColorDark,
				PrimaryColorLight: userData.PrimaryColorLight,
				Password:          pwHash,
				SchedulerEnabled:  userData.SchedulerEnabled,
				DarkTheme:         userData.DarkTheme,
			},
			ProfilePicture:    profilePicture,
			Tokens:            tokens,
			Homescripts:       homescripts,
			HomescriptStorage: storageOutput,
			Reminders:         reminders,
			Permissions:       permissions,
			SwitchPermissions: swPermissions,
			CameraPermissions: camPermissions,
		})
	}

	// Include cache data if desired
	cacheData := setupCacheData{
		WeatherHistory: make([]setupWeatherMeasurement, 0),
		PowerData:      make([]hardware.PowerDrawDataPointUnixMillis, 0),
	}
	if includedCacheData {
		// Weather history
		weatherHistory, err := database.GetWeatherDataRecords(-1)
		if err != nil {
			return SetupStruct{}, err
		}

		// Transform to the type that uses unix millis
		weatherHistoryOut := make([]setupWeatherMeasurement, 0)
		for _, measurement := range weatherHistory {
			weatherHistoryOut = append(weatherHistoryOut, setupWeatherMeasurement{
				Id:                 measurement.Id,
				Time:               uint(measurement.Time.UnixMilli()),
				WeatherTitle:       measurement.WeatherTitle,
				WeatherDescription: measurement.WeatherDescription,
				Temperature:        measurement.Temperature,
				FeelsLike:          measurement.FeelsLike,
				Humidity:           measurement.Humidity,
			})
		}
		cacheData.WeatherHistory = weatherHistoryOut

		// Power data
		powerData, err := hardware.GetPowerUsageRecordsUnixMillis(-1)
		if err != nil {
			return SetupStruct{}, err
		}
		cacheData.PowerData = powerData
	}

	return SetupStruct{
		Users:               users,
		Rooms:               rooms,
		HardwareNodes:       hwNodesNew,
		ServerConfiguration: serverConfig,
		CacheData:           cacheData,
	}, nil
}
