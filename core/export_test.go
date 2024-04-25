package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/core/user"
	"github.com/stretchr/testify/assert"
)

func TestExportGeneration(t *testing.T) {
	assert.NoError(t, initDB(true))
	log := logrus.New()
	event.InitLogger(log)
	user.InitLogger(log)
	homescript.InitLogger(log)

	config, found, err := database.GetServerConfiguration()
	assert.NoError(t, err)
	assert.True(t, found)

	assert.NoError(t, homescript.InitAutomations(config))
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

	testDriver := database.DeviceDriver{
		VendorID:       "go",
		ModelID:        "test",
		Name:           "Golang-Test",
		Version:        "0.0.1",
		HomescriptCode: homescript.DefaultDriverHomescriptCode,
		SingletonJSON:  nil,
	}

	assert.NoError(t, database.CreateNewDeviceDriver(testDriver))

	// TODO: remove this.
	// Create Hardware nodes
	// testNode := database.HardwareNode{
	// 	Name:    "Living Room",
	// 	Url:     "http://10.0.0.1:7000",
	// 	Token:   "secret_t0ken",
	// 	Enabled: true,
	// }

	// assert.NoError(t, database.CreateHardwareNode(testNode))

	// Create a room with contents
	assert.NoError(t, database.CreateRoom(database.RoomData{
		ID:          "living_room",
		Name:        "Living Room",
		Description: "Where the people live...",
	}))

	// Create devices
	driverFound, hmsErr, dbErr := homescript.CreateDevice(
		database.DEVICE_TYPE_OUTPUT,
		"big_lamp",
		"Big Lamp",
		"living_room",
		testDriver.VendorID,
		testDriver.ModelID,
	)

	assert.True(t, driverFound)
	assert.NoError(t, dbErr)
	assert.NoError(t, hmsErr)

	driverFound, hmsErr, dbErr = homescript.CreateDevice(
		database.DEVICE_TYPE_OUTPUT,
		"desk_lamp",
		"Desk Lamp",
		"living_room",
		testDriver.VendorID,
		testDriver.ModelID,
	)

	assert.True(t, driverFound)
	assert.NoError(t, dbErr)
	assert.NoError(t, hmsErr)

	// Create cameras
	assert.NoError(t, database.CreateCamera(database.Camera{
		ID:     "lvr_main_door",
		Name:   "Main Door",
		RoomID: "living_room",
		Url:    "http://example.com/1",
	}))
	assert.NoError(t, database.CreateCamera(database.Camera{
		ID:     "lvr_shelf",
		Name:   "Shelf",
		RoomID: "living_room",
		Url:    "http://example.com/2",
	}))

	// Create an additional, empty room
	assert.NoError(t, database.CreateRoom(database.RoomData{
		ID:          "server_room",
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
	_, err = database.AddUserDevicePermission("test", "big_lamp")
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
	var interval uint = 42
	_, err = homescript.CreateNewAutomation(
		"My Automation",
		"This is a description",
		"automation_homescript",
		"test",
		true,
		nil,
		nil,
		nil,
		database.TriggerInterval,
		&interval,
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
	_, err = ExportConfig(false, true)
	assert.NoError(t, err)
}
