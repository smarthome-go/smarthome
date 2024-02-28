package core

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
)

func TestRunSetup(t *testing.T) {
	cronExpression := "1 2 * * *"
	var interval uint = 42

	setup := SetupStruct{
		Users: []SetupUser{
			{
				Data: SetupUserData{
					Username:          "setup",
					Forename:          "Set",
					Surname:           "Up",
					PrimaryColorDark:  "#000",
					PrimaryColorLight: "#fff",
					Password:          "would-be-a-hash",
					SchedulerEnabled:  true,
					DarkTheme:         true,
				},
				Homescripts: []SetupHomescript{
					{
						Data: database.HomescriptData{
							Id:                  "setup_hms",
							Name:                "Setup Test",
							Description:         "A HMS for testing the setup",
							QuickActionsEnabled: true,
							SchedulerEnabled:    true,
							Code:                "print('Hello World!')",
							MDIcon:              "code",
						},
						Arguments: []SetupHomescriptArg{
							{
								ArgKey:    "a_key",
								Prompt:    "Enter your value",
								MDIcon:    "code",
								InputType: database.String,
								Display:   database.TypeDefault,
							},
							{
								ArgKey:    "another_key",
								Prompt:    "Enter your second value",
								MDIcon:    "code",
								InputType: database.Number,
								Display:   database.NumberHour,
							},
						},
						Automations: []SetupAutomation{
							{
								Name:                  "automation (cron)",
								Description:           "An automation for testing the setup using cron",
								TriggerCronExpression: &cronExpression,
								Enabled:               true,
								Trigger:               database.TriggerCron,
							},
							{
								Name:                   "automation (interval)",
								Description:            "An automation for testing the setup using interval",
								TriggerCronExpression:  nil,
								Enabled:                true,
								Trigger:                database.TriggerInterval,
								TriggerIntervalSeconds: &interval,
							},
							{
								Name:                   "automation (on notification)",
								Description:            "An automation for testing the setup using on_notification",
								TriggerCronExpression:  nil,
								Enabled:                true,
								Trigger:                database.TriggerOnNotification,
								TriggerIntervalSeconds: nil,
							},
						},
					},
				},
				Reminders: []SetupReminder{
					{
						Name:              "Do something",
						Description:       "This is an important task",
						Priority:          database.Urgent,
						CreatedDate:       time.Now(),
						DueDate:           time.Now().Add(time.Hour * 24),
						UserWasNotifiedAt: time.Time{},
					},
				},
				Permissions: []string{
					string(database.PermissionAutomation),
					string(database.PermissionPower),
					string(database.PermissionHomescript),
					string(database.PermissionReminder),
				},
				DevicePermissions: []string{"lvr_big_lamp"},
				CameraPermissions: []string{"lvr_main_door"},
			},
		},
		Rooms: []SetupRoom{
			{
				Data: database.RoomData{
					ID:          "living_room",
					Name:        "Living Room",
					Description: "This is the room where people live in",
				},
				Devices: []SetupDevice{
					{
						DeviceType:    database.DEVICE_TYPE_OUTPUT,
						Id:            "lvr_big_lamp",
						Name:          "Big Lamp",
						VendorId:      "foo-vendor",
						ModelId:       "bar-model",
						SingletonJSON: "{\"runs\": 42}",
					},
				},
				Cameras: []SetupCamera{
					{
						Id:   "lvr_main_door",
						Name: "Living Room Main Door",
						Url:  "http://example.com/1",
					},
				},
			},
		},
		ServerConfiguration: database.ServerConfig{
			AutomationEnabled: false,
			LockDownMode:      false,
			Latitude:          0.0,
			Longitude:         0.0}}
	// Write the json to a temp directory so it can be read later
	SetupPath = fmt.Sprintf("%s/setup.json", t.TempDir()) // Global variable is changed here in order to use the temp location
	content, err := json.Marshal(setup)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if err := os.WriteFile(
		SetupPath,
		content,
		0755,
	); err != nil {
		t.Error(err.Error())
		return
	}

	// Run the setup
	if err := RunSetup(); err != nil {
		t.Error(err.Error())
		return
	}

	for _, device := range setup.Rooms[0].Devices {
		_, exists, err := database.GetDeviceById(device.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !exists {
			t.Errorf("Device %s does not exist after setup was completed", device.Id)
			return
		}
	}

	rooms, err := database.ListRooms()
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, setupRoom := range setup.Rooms {
		roomValid := false
		for _, room := range rooms {
			if room.ID == setupRoom.Data.ID && room.Description == setupRoom.Data.Description {
				roomValid = true
			}
		}
		if !roomValid {
			t.Errorf("Room %s does not exist after creation", setupRoom.Data.ID)
			return
		}
	}

	// TODO: test drivers
}

func TestReadBrokenSetupFile(t *testing.T) {
	// Write the bad contents to another temp directory so it can be read later
	SetupPath = fmt.Sprintf("%s/setup_invalid.json", t.TempDir())
	if err := os.WriteFile(
		SetupPath,
		[]byte("invalid_content"),
		0755,
	); err != nil {
		t.Error(err.Error())
		return
	}
	if err := RunSetup(); err == nil {
		t.Error("Error expected whilst parsing broken setup file but none occurred")
		return
	}
}

func TestSetupFileDoesNotExist(t *testing.T) {
	SetupPath = "/does/not/exist"
	_, fileExists, err := readSetupFile()
	if err != nil {
		t.Error(err.Error())
		return
	}
	if fileExists {
		t.Errorf("Non-existent file %s was readable by function", SetupPath)
		return
	}
}
