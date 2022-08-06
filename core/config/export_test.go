package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/scheduler/automation"
	"github.com/smarthome-go/smarthome/core/user"
	"github.com/stretchr/testify/assert"
)

func TestExportGeneration(t *testing.T) {
	assert.NoError(t, initDB(true))
	log := logrus.New()
	event.InitLogger(log)
	user.InitLogger(log)
	automation.InitLogger(log)
	assert.NoError(t, automation.Init())
	///
	/// Part 1: Mock data creation
	///
	// Modify server configuration
	assert.NoError(t, database.SetServerConfiguration(database.ServerConfig{
		AutomationEnabled: true,
		LockDownMode:      false,
		Latitude:          3.14159265,
		Longitude:         3.14159265,
	}))
	// Create Hardware nodes
	assert.NoError(t, database.CreateHardwareNode(database.HardwareNode{
		Name:    "Living Room",
		Url:     "http://10.0.0.1:7000",
		Token:   "secret_t0ken",
		Enabled: true,
	}))
	// Create a room with contents
	assert.NoError(t, database.CreateRoom(database.RoomData{
		Id:          "living_room",
		Name:        "Living Room",
		Description: "Where the people live...",
	}))
	// Create switches
	assert.NoError(t, database.CreateSwitch(
		"big_lamp",
		"Big Lamp",
		"living_room",
		42,
	))
	assert.NoError(t, database.CreateSwitch(
		"desk_lamp",
		"Desk Lamp",
		"living_room",
		24,
	))
	// Create cameras
	assert.NoError(t, database.CreateCamera(database.Camera{
		Id:     "lvr_main_door",
		Name:   "Main Door",
		RoomId: "living_room",
		Url:    "http://example.com/1",
	}))
	assert.NoError(t, database.CreateCamera(database.Camera{
		Id:     "lvr_shelf",
		Name:   "Shelf",
		RoomId: "living_room",
		Url:    "http://example.com/2",
	}))
	// Create an additional, empty room
	assert.NoError(t, database.CreateRoom(database.RoomData{
		Id:          "server_room",
		Name:        "Server Room",
		Description: "Where the server serves...",
	}))
	// Create user
	assert.NoError(t, database.AddUser(
		database.FullUser{
			Username:          "test",
			Password:          "test",
			Forename:          "Forename",
			Surname:           "Surname",
			PrimaryColorDark:  "#88FF70",
			PrimaryColorLight: "#2E7D32",
		}))
	// Grant user some permissions
	permissions := []database.PermissionType{
		database.PermissionAutomation,
		database.PermissionHomescript,
		database.PermissionReminder,
		database.PermissionPower,
		database.PermissionViewCameras,
	}
	for _, permission := range permissions {
		_, err := user.AddPermission(
			"test",
			permission,
		)
		assert.NoError(t, err)
	}
	// Grant the user one switch permission
	_, err := database.AddUserSwitchPermission("test", "big_lamp")
	assert.NoError(t, err)
	// Grant the user one camera permission
	_, err = database.AddUserCameraPermission("test", "lvr_shelf")
	assert.NoError(t, err)
	// Create Homescript
	assert.NoError(t, database.CreateNewHomescript(database.Homescript{
		Owner: "test",
		Data: database.HomescriptData{
			Id:                  "my_homescript",
			Name:                "My Homescript",
			Description:         "This is a Homecsript!",
			QuickActionsEnabled: true,
			SchedulerEnabled:    false,
			Code:                "print('Hello World!')",
			MDIcon:              "code",
		},
	}))
	// Create Homescript args
	for i := 0; i < 2; i++ {
		_, err := database.AddHomescriptArg(database.HomescriptArgData{
			ArgKey:       fmt.Sprintf("my_key_%d", i),
			HomescriptId: "my_homescript",
			Prompt:       fmt.Sprintf("Enter value for %d.", i),
			MDIcon:       fmt.Sprintf("print(%d)", i),
			InputType:    database.String,
			Display:      database.TypeDefault,
		})
		assert.NoError(t, err)
	}
	// Create another Homescript for the automation
	assert.NoError(t, database.CreateNewHomescript(database.Homescript{
		Owner: "test",
		Data: database.HomescriptData{
			Id:                  "automation_homescript",
			Name:                "Automation HMS",
			Description:         "This is a Homecsript for an automation!",
			QuickActionsEnabled: true,
			SchedulerEnabled:    true,
			Code:                "print('Hello World!')",
			MDIcon:              "code",
		},
	}))
	// Create automation
	_, err = automation.CreateNewAutomation(
		"My Automation",
		"This is a description.",
		4,
		2,
		[]uint8{0, 1, 2, 3, 4},
		"my_homescript",
		"test",
		true,
		database.TimingNormal,
	)
	assert.NoError(t, err)
	// Create Reminder
	_, err = database.CreateNewReminder(
		"Do something!",
		"You should do this [...]",
		time.Now().Add(time.Duration(time.Hour*24)),
		"test",
		database.Urgent,
	)
	assert.NoError(t, err)
	///
	/// Part 2: Export testing
	///
	_, err = Export()
	assert.NoError(t, err)
}
