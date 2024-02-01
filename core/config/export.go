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
	Users []SetupUser `json:"users"`
	Rooms []SetupRoom `json:"rooms"`
	// HardwareNodes       []SetupHardwareNode   `json:"hardwareNodes"`
	ServerConfiguration database.ServerConfig `json:"serverConfiguration"`
	CacheData           SetupCacheData        `json:"cacheData"`
}

type SetupRoom struct {
	Data    database.RoomData `json:"data"`
	Devices []SetupDevice     `json:"devices"`
	Cameras []SetupCamera     `json:"cameras"`
}

type SetupDevice struct {
	DeviceType    database.DEVICE_TYPE `json:"deviceType"`
	Id            string               `json:"id"`
	Name          string               `json:"name"`
	VendorId      string               `json:"vendorId"`
	ModelId       string               `json:"modelId"`
	SingletonJSON string               `json:"SingletonJson"`
}

type SetupCamera struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type SetupUser struct {
	Data              SetupUserData            `json:"user"`
	Tokens            []SetupAuthToken         `json:"tokens"`
	Homescripts       []SetupHomescript        `json:"homescripts"`
	HomescriptStorage []SetupHomescriptStorage `json:"homescriptStorage"`
	Reminders         []SetupReminder          `json:"reminders"`

	// Profile picture as B64
	ProfilePicture *SetupUserProfilePicture `json:"profilePicture"`

	// Permissions
	Permissions       []string `json:"permissions"`
	DevicePermissions []string `json:"devicePermissions"`
	CameraPermissions []string `json:"cameraPermissions"`
}

type SetupUserProfilePicture struct {
	B64Data       string `json:"data"`
	FileExtension string `json:"fileExtension"`
}

type SetupHardwareNode struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"` // Can be used to temporarily deactivate a node in case of maintenance
	Url     string `json:"url"`
	Token   string `json:"token"`
}

type SetupAuthToken struct {
	Token string `json:"token"`
	Label string `json:"label"`
}

type SetupReminder struct {
	Name              string                    `json:"name"`
	Description       string                    `json:"description"`
	Priority          database.ReminderPriority `json:"priority"`
	CreatedDate       time.Time                 `json:"createdDate"`
	DueDate           time.Time                 `json:"dueDate"`
	UserWasNotified   bool                      `json:"userWasNotified"`
	UserWasNotifiedAt time.Time                 `json:"userWasNotifiedAt"`
}

type SetupHomescript struct {
	Data        database.HomescriptData `json:"data"`
	Arguments   []SetupHomescriptArg    `json:"arguments"`
	Automations []SetupAutomation       `json:"automations"`
}

type SetupHomescriptArg struct {
	ArgKey    string                   `json:"argKey"`    // The unique key of the argument
	Prompt    string                   `json:"prompt"`    // The prompt the user will see
	MDIcon    string                   `json:"mdIcon"`    // A Google MD icon which will be displayed
	InputType database.HmsArgInputType `json:"inputType"` // Specifies the expected data type
	Display   database.HmsArgDisplay   `json:"display"`   // Specifies the visual display of the prompt (handled by GUI)
}

type SetupHomescriptStorage struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SetupAutomation struct {
	Name                   string                     `json:"name"`
	Description            string                     `json:"description"`
	Enabled                bool                       `json:"enabled"`
	Trigger                database.AutomationTrigger `json:"trigger"`
	TriggerCronExpression  *string                    `json:"cronExpression"`
	TriggerIntervalSeconds *uint                      `json:"intervalSeconds"`
}

type SetupUserData struct {
	Username          string `json:"username"`
	Forename          string `json:"forename"`
	Surname           string `json:"surname"`
	PrimaryColorDark  string `json:"primaryColorDark"`
	PrimaryColorLight string `json:"primaryColorLight"`
	Password          string `json:"password"`
	SchedulerEnabled  bool   `json:"schedulerEnabled"`
	DarkTheme         bool   `json:"darkTheme"`
}

type SetupCacheData struct {
	WeatherHistory []SetupWeatherMeasurement               `json:"weatherHistory"`
	PowerUsageData []hardware.PowerDrawDataPointUnixMillis `json:"powerUsageData"`
}

type SetupWeatherMeasurement struct {
	Id                 uint64  `json:"id"`
	Time               uint64  `json:"time"` // unix millis
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
	rooms := make([]SetupRoom, 0)
	for _, room := range roomsDB {
		roomDevices := make([]SetupDevice, 0)
		for _, sw := range room.Devices {
			roomDevices = append(roomDevices, SetupDevice{
				DeviceType: sw.DeviceType,
				Id:         sw.Id,
				Name:       sw.Name,
				VendorId:   sw.VendorId,
				ModelId:    sw.ModelId,
			})
		}

		roomCameras := make([]SetupCamera, 0)
		for _, cam := range room.Cameras {
			roomCameras = append(roomCameras, SetupCamera{
				Id:   cam.Id,
				Name: cam.Name,
				Url:  cam.Url,
			})
		}

		rooms = append(rooms, SetupRoom{
			Data: database.RoomData{
				Id:          room.Data.Id,
				Name:        room.Data.Name,
				Description: room.Data.Description,
			},
			Devices: roomDevices,
			Cameras: roomCameras,
		})
	}

	// hwNodes, err := database.GetHardwareNodes()
	// if err != nil {
	// 	return SetupStruct{}, err
	// }
	// hwNodesNew := make([]SetupHardwareNode, 0)
	// for _, node := range hwNodes {
	// 	hwNodesNew = append(hwNodesNew, SetupHardwareNode{
	// 		Name:    node.Name,
	// 		Enabled: node.Enabled,
	// 		Token:   node.Token,
	// 		Url:     node.Url,
	// 	})
	// }

	usersTemp, err := database.ListUsers()
	if err != nil {
		return SetupStruct{}, nil
	}
	users := make([]SetupUser, 0)
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
		tokens := make([]SetupAuthToken, 0)
		for _, t := range tokensDB {
			tokens = append(tokens, SetupAuthToken{
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
		homescripts := make([]SetupHomescript, 0)
		for _, hms := range homescriptsDB {
			args := make([]SetupHomescriptArg, 0)
			for _, arg := range hms.Arguments {
				args = append(args, SetupHomescriptArg{
					ArgKey:    arg.Data.ArgKey,
					Prompt:    arg.Data.Prompt,
					MDIcon:    arg.Data.MDIcon,
					InputType: arg.Data.InputType,
					Display:   arg.Data.Display,
				})
			}
			// Filter out automations using this Homescript
			automationsThis := make([]SetupAutomation, 0)
			for _, aut := range automationsDB {
				if aut.Data.HomescriptId == hms.Data.Data.Id {
					automationsThis = append(automationsThis, SetupAutomation{
						Name:                   aut.Data.Name,
						Description:            aut.Data.Description,
						Enabled:                aut.Data.Enabled,
						Trigger:                aut.Data.Trigger,
						TriggerCronExpression:  aut.Data.TriggerCronExpression,
						TriggerIntervalSeconds: aut.Data.TriggerIntervalSeconds,
					})
				}
			}
			homescripts = append(homescripts, SetupHomescript{
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

		storageOutput := make([]SetupHomescriptStorage, 0)
		for key, value := range storage {
			storageOutput = append(storageOutput, SetupHomescriptStorage{
				Key:   key,
				Value: value,
			})
		}

		// Reminders
		remindersDB, err := database.GetUserReminders(userData.Username)
		if err != nil {
			return SetupStruct{}, err
		}
		reminders := make([]SetupReminder, 0)
		for _, reminder := range remindersDB {
			reminders = append(reminders, SetupReminder{
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
		devPermissions, err := database.GetUserDevicePermissions(userData.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Camera permissions
		camPermissions, err := database.GetUserCameraPermissions(userData.Username)
		if err != nil {
			return SetupStruct{}, err
		}

		// Include profile picture if desired
		var profilePicture *SetupUserProfilePicture = nil
		if includeProfilePictures {
			picture, err := user.GetUserAvatar(userData.Username)
			if err != nil {
				return SetupStruct{}, err
			}

			filetype, err := filetype.Match(picture)
			if err != nil {
				return SetupStruct{}, err
			}

			picTemp := SetupUserProfilePicture{
				B64Data:       base64.StdEncoding.EncodeToString(picture),
				FileExtension: filetype.Extension,
			}

			profilePicture = &picTemp
		}

		// Append assembled user
		users = append(users, SetupUser{
			Data: SetupUserData{
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
			DevicePermissions: devPermissions,
			CameraPermissions: camPermissions,
		})
	}

	// Include cache data if desired
	cacheData := SetupCacheData{
		WeatherHistory: make([]SetupWeatherMeasurement, 0),
		PowerUsageData: make([]hardware.PowerDrawDataPointUnixMillis, 0),
	}
	if includedCacheData {
		// Weather history
		weatherHistory, err := database.GetWeatherDataRecords(-1)
		if err != nil {
			return SetupStruct{}, err
		}

		// Transform to the type that uses unix millis
		weatherHistoryOut := make([]SetupWeatherMeasurement, 0)
		for _, measurement := range weatherHistory {
			weatherHistoryOut = append(weatherHistoryOut, SetupWeatherMeasurement{
				Id:                 measurement.Id,
				Time:               uint64(measurement.Time.UnixMilli()),
				WeatherTitle:       measurement.WeatherTitle,
				WeatherDescription: measurement.WeatherDescription,
				Temperature:        measurement.Temperature,
				FeelsLike:          measurement.FeelsLike,
				Humidity:           measurement.Humidity,
			})
		}
		cacheData.WeatherHistory = weatherHistoryOut

		// Power usage data
		powerData, err := hardware.GetPowerUsageRecordsUnixMillis(-1)
		if err != nil {
			return SetupStruct{}, err
		}
		cacheData.PowerUsageData = powerData
	}

	return SetupStruct{
		Users: users,
		Rooms: rooms,
		// HardwareNodes:       hwNodesNew,
		ServerConfiguration: serverConfig,
		CacheData:           cacheData,
	}, nil
}
