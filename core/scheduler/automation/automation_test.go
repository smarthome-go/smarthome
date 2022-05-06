package automation

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/homescript"
)

// Sets up the tests dependencies
func TestMain(m *testing.M) {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	InitLogger(log)
	event.InitLogger(log)
	homescript.InitLogger(log)
	hardware.InitLogger(log)
	if err := initDB(true); err != nil {
		panic(err.Error())
	}
	if err := createMockData(); err != nil {
		panic(err.Error())
	}
	hardware.Init()
	code := m.Run()
	os.Exit(code)
}

// Creates mock data, including a room, switches and homescripts
func createMockData() error {
	if err := database.CreateRoom(database.RoomData{Id: "test_room"}); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch", "", "test_room", 0); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch_modify", "", "test_room", 0); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch_inactive", "", "test_room", 0); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch_abort", "", "test_room", 0); err != nil {
		panic(err.Error())
	}
	_, doesExists, err := database.GetUserHomescriptById("test", "admin")
	if err != nil {
		panic(err.Error())
	}
	if !doesExists {
		// Create Homescript
		if err := database.CreateNewHomescript(database.Homescript{
			Id:                  "test",
			Owner:               "admin",
			Name:                "Testing",
			Description:         "A Homescript for testing purposes",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "switch('test_switch', on)",
		}); err != nil {
			panic(err.Error())
		}
	}
	_, doesExists, err = database.GetUserHomescriptById("test_modify", "admin")
	if err != nil {
		panic(err.Error())
	}
	if !doesExists {
		// Create another Homescript
		if err := database.CreateNewHomescript(database.Homescript{
			Id:                  "test_modify",
			Owner:               "admin",
			Name:                "Testing 2",
			Description:         "Another Homescript for testing purposes",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "switch('test_switch_modify', on)",
		}); err != nil {
			panic(err.Error())
		}
	}
	_, doesExists, err = database.GetUserHomescriptById("test_inactive", "admin")
	if err != nil {
		panic(err.Error())
	}
	if !doesExists {
		// Create another Homescript
		if err := database.CreateNewHomescript(database.Homescript{
			Id:                  "test_inactive",
			Owner:               "admin",
			Name:                "Testing 2",
			Description:         "Another Homescript for testing purposes",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "switch('test_switch_inactive', on)",
		}); err != nil {
			panic(err.Error())
		}
	}
	_, doesExists, err = database.GetUserHomescriptById("test_abort", "admin")
	if err != nil {
		panic(err.Error())
	}
	if !doesExists {
		// Create another Homescript
		if err := database.CreateNewHomescript(database.Homescript{
			Id:                  "test_abort",
			Owner:               "admin",
			Name:                "Testing 2",
			Description:         "Another Homescript for testing purposes",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "switch('test_switch_abort', on)",
		}); err != nil {
			panic(err.Error())
		}
	}
	return nil
}

func initDB(args ...bool) error {
	database.InitLogger(logrus.New())
	if err := database.Init(database.DatabaseConfig{
		Username: "smarthome",
		Password: "testing",
		Hostname: "localhost",
		Database: "smarthome",
		Port:     3330,
	}, "admin",
	); err != nil {
		return err
	}
	if len(args) > 0 {
		if err := database.DeleteTables(); err != nil {
			return err
		}
		time.Sleep(time.Second)
		return initDB()
	}
	return nil
}

// Creates a regular automation and checks if it is executed on time
func TestAutomation(t *testing.T) {
	now := time.Now()
	then := now.Add(time.Minute)
	if err := Init(); err != nil {
		t.Error(err.Error())
		return
	}
	// Normal automation
	if _, err := CreateNewAutomation(
		"name",
		"description",
		uint8(then.Hour()),
		uint8(then.Minute()),
		[]uint8{0, 1, 2, 3, 4, 5, 6},
		"test",
		"admin",
		true,
		database.TimingNormal,
	); err != nil {
		t.Error(err.Error())
		return
	}
	valid := false
	for i := 0; i < 7; i++ {
		time.Sleep(time.Second * 10)
		switchItem, found, err := database.GetSwitchById("test_switch")
		if err != nil {
			t.Error(err.Error())
			return
		}
		assert.True(t, found, "Switch not found")
		if switchItem.PowerOn {
			valid = true
			break
		}
	}
	if !valid {
		t.Error("Power of `test_switch` did not change")
		return
	}
}

// Creates a automation which executes the script `x`
// Modifies the automation so that it executes the script `y` instead
// X and Y turn on a different switch which represents the script
// If the switch of Y was not turned on, the test is considered a failure
func TestModificationToDifferentScript(t *testing.T) {
	now := time.Now()
	then := now.Add(time.Minute)
	if err := Init(); err != nil {
		t.Error(err.Error())
		return
	}
	// Create initial automation
	modifyId, err := CreateNewAutomation(
		"name",
		"description",
		uint8(then.Hour()),
		uint8(then.Minute()),
		[]uint8{0, 1, 2, 3, 4, 5, 6},
		"test",
		"admin",
		true,
		database.TimingNormal,
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	// Modify second automation to use the other homescript file
	cronExpression, err := GenerateCronExpression(uint8(then.Hour()), uint8(then.Minute()), []uint8{0, 1, 2, 3, 4, 5, 6})
	if err != nil {
		t.Error(err.Error())
		return
	}
	if err := ModifyAutomationById(modifyId,
		database.AutomationWithoutIdAndUsername{
			Name:           "name",
			Description:    "description",
			CronExpression: cronExpression,
			HomescriptId:   "test_modify",
			Enabled:        true,
			TimingMode:     database.TimingNormal,
		}); err != nil {
		t.Error(err.Error())
		return
	}
	// Wait for approx. a minute in order to dertermine if the modification succeeded
	valid := false
	for i := 0; i < 7; i++ {
		time.Sleep(time.Second * 10)
		switchItem, found, err := database.GetSwitchById("test_switch_modify")
		if err != nil {
			t.Error(err.Error())
			return
		}
		assert.True(t, found, "Switch not found")
		if switchItem.PowerOn {
			valid = true
		}
	}
	// Check if the updated script did change the power
	if !valid {
		t.Error("Power of `test_switch_modify` did not change")
		return
	}
}

// Creates a regular automation which is then modified to be disabled
// Checks if the automation still executes despite being disabled
func TestModificationToAbort(t *testing.T) {
	now := time.Now()
	then := now.Add(time.Minute)
	if err := Init(); err != nil {
		t.Error(err.Error())
		return
	}
	// Create initial automation
	abortId, err := CreateNewAutomation(
		"name",
		"description",
		uint8(then.Hour()),
		uint8(then.Minute()),
		[]uint8{0, 1, 2, 3, 4, 5, 6},
		"test_abort",
		"admin",
		true,
		database.TimingNormal,
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	// Gets the initial automation in order to copy its cron-expression
	// The old cron-expression is required in order to preserve the time window in
	// which the automation could be executed if the modification fails
	automation, found, err := database.GetAutomationById(abortId)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Automation %d not found in database", abortId)
		return
	}
	// Set its activation status to `disabled`
	if err := ModifyAutomationById(abortId,
		database.AutomationWithoutIdAndUsername{
			Name:           "name",
			Description:    "description",
			CronExpression: automation.CronExpression,
			HomescriptId:   "test_abort",
			Enabled:        false,
			TimingMode:     database.TimingNormal,
		}); err != nil {
		t.Error(err.Error())
		return
	}
	// Wait for approx. a minute until the test can be considered to be successful
	for i := 0; i < 7; i++ {
		time.Sleep(time.Second * 10)
		switchItem, found, err := database.GetSwitchById("test_switch_abort")
		if err != nil {
			t.Error(err.Error())
			return
		}
		assert.True(t, found, "Switch not found")
		if switchItem.PowerOn {
			t.Errorf("Power of `test_switch_abort` changed but should not")
			return
		}
	}
}

// TODO: add a initially disabled automation which is then enabled

// Creates an automation which is initially disabled
// Checks if the automation runs despite being disabled
func TestStartInactiveAutomation(t *testing.T) {
	now := time.Now()
	then := now.Add(time.Minute)
	if _, err := CreateNewAutomation(
		"name",
		"description",
		uint8(then.Hour()),
		uint8(then.Minute()),
		[]uint8{0, 1, 2, 3, 4, 5, 6},
		"test_inactive",
		"admin",
		false,
		database.TimingNormal,
	); err != nil {
		t.Error(err.Error())
		return
	}
	// Wait for approx. a minute until to decide that the test was executed successfully
	for i := 0; i < 7; i++ {
		time.Sleep(time.Second * 10)
		switchItem, found, err := database.GetSwitchById("test_switch_inactive")
		if err != nil {
			t.Error(err.Error())
			return
		}
		assert.True(t, found, "Switch not found")
		if switchItem.PowerOn {
			t.Errorf("Power of `test_switch_inactive` changed but should not")
			return
		}
	}
}

// Tests if the different timing modes `sunrise` and `sunset` generate appropriate Cron-Expressions
func TestTimingModes(t *testing.T) {
	if err := Init(); err != nil {
		t.Error(err.Error())
		return
	}
	sunriseId, err := CreateNewAutomation(
		"name",
		"description",
		23,
		59,
		[]uint8{0, 1, 2, 3, 4, 5, 6},
		"test",
		"admin",
		true,
		database.TimingSunrise,
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	sunSetId, err := CreateNewAutomation(
		"name",
		"description",
		23,
		59,
		[]uint8{0, 1, 2, 3, 4, 5, 6},
		"test",
		"admin",
		true,
		database.TimingSunset,
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	sunrise, found, err := database.GetAutomationById(sunriseId)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Automation %d was not found after creation", sunriseId)
		return
	}
	sunSet, found, err := database.GetAutomationById(sunSetId)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Automation %d was not found after creation", sunSetId)
		return
	}
	if sunrise.CronExpression == sunSet.CronExpression {
		t.Errorf("Cron expression of sunrise and sunset is not valid. `%s`|`%s`", sunSet.CronExpression, sunrise.CronExpression)
	}
}

// Tests if the automation system can be initialized
func TestInit(t *testing.T) {
	if err := Init(); err != nil {
		t.Error(err.Error())
		return
	}
}

// Deactivates and Reactivates the automation system in order to check for errors
func TestActivate(t *testing.T) {
	TestInit(t)
	if err := DeactivateAutomationSystem(); err != nil {
		t.Error(err.Error())
	}
	if err := ActivateAutomationSystem(); err != nil {
		t.Error(err.Error())
	}
}
