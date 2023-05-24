package homescript

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/homescript/automation"
)

// Creates mock data, including a room, switches and homescripts
func createMockData() error {
	if err := database.CreateRoom(database.RoomData{Id: "test_room"}); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch", "", "test_room", 0, nil); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch_modify", "", "test_room", 0, nil); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch_inactive", "", "test_room", 0, nil); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch_abort", "", "test_room", 0, nil); err != nil {
		panic(err.Error())
	}
	_, doesExists, err := database.GetUserHomescriptById("test", "admin")
	if err != nil {
		panic(err.Error())
	}
	if !doesExists {
		// Create Homescript
		if err := database.CreateNewHomescript(database.Homescript{
			Owner: "admin",
			Data: database.HomescriptData{
				Id:                  "test",
				Name:                "Testing",
				Description:         "A Homescript for testing purposes",
				QuickActionsEnabled: false,
				SchedulerEnabled:    false,
				Code:                "switch('test_switch', on);",
			},
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
			Owner: "admin",
			Data: database.HomescriptData{
				Id:                  "test_modify",
				Name:                "Testing 2",
				Description:         "Another Homescript for testing purposes",
				QuickActionsEnabled: false,
				SchedulerEnabled:    false,
				Code:                "switch('test_switch_modify', on);",
			},
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
			Owner: "admin",
			Data: database.HomescriptData{
				Id:                  "test_inactive",
				Name:                "Testing 2",
				Description:         "Another Homescript for testing purposes",
				QuickActionsEnabled: false,
				SchedulerEnabled:    false,
				Code:                "switch('test_switch_inactive', on);",
			},
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
			Owner: "admin",
			Data: database.HomescriptData{
				Id:                  "test_abort",
				Name:                "Testing 2",
				Description:         "Another Homescript for testing purposes",
				QuickActionsEnabled: false,
				SchedulerEnabled:    false,
				Code:                "switch('test_switch_abort', on);",
			},
		}); err != nil {
			panic(err.Error())
		}
	}
	return nil
}

// Creates a regular automation and checks if it is executed on time
func TestAutomation(t *testing.T) {
	now := time.Now()
	then := now.Add(time.Minute)
	TestInit(t)
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
	TestInit(t)
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
	cronExpression, err := automation.GenerateCronExpression(uint8(then.Hour()), uint8(then.Minute()), []uint8{0, 1, 2, 3, 4, 5, 6})
	if err != nil {
		t.Error(err.Error())
		return
	}
	if err := ModifyAutomationById(modifyId,
		database.AutomationData{
			Name:                  "name",
			Description:           "description",
			TriggerCronExpression: cronExpression,
			HomescriptId:          "test_modify",
			Enabled:               true,
			DisableOnce:           false,
			TimingMode:            database.TimingNormal,
		}); err != nil {
		t.Error(err.Error())
		return
	}
	// Wait for approx. a minute in order to determine if the modification succeeded
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
	TestInit(t)
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
		database.AutomationData{
			Name:                  "name",
			Description:           "description",
			TriggerCronExpression: automation.Data.TriggerCronExpression,
			HomescriptId:          "test_abort",
			Enabled:               false,
			TimingMode:            database.TimingNormal,
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
	TestInit(t)
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
	if sunrise.Data.TriggerCronExpression == sunSet.Data.TriggerCronExpression {
		t.Errorf("Cron expression of sunrise and sunset is not valid. `%s`|`%s`", sunSet.Data.TriggerCronExpression, sunrise.Data.TriggerCronExpression)
	}
}

func TestUserDisabled(t *testing.T) {
	now := time.Now()
	then := now.Add(time.Minute)
	TestInit(t)

	testSwitch, _, err := database.GetSwitchById("test_switch")
	if err != nil {
		t.Error(err)
		return
	}

	// Set the switch to off
	assert.NoError(t, hardware.SetPower(testSwitch, false))
	// Set the user's schedules and automations to off
	assert.NoError(t, database.SetUserSchedulerEnabled("admin", false))
	// Normal automation
	if _, err := CreateNewAutomation(
		"name_pers",
		"description_pers",
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
	if valid {
		t.Error("Power of `test_switch` changed but should not")
		return
	}
}

func TestDisableOnce(t *testing.T) {
	now := time.Now()
	then := now.Add(time.Minute)
	TestInit(t)

	testSwitch, _, err := database.GetSwitchById("test_switch")
	if err != nil {
		t.Error(err)
		return
	}

	// Set the switch to off
	assert.NoError(t, hardware.SetPower(testSwitch, false))
	// Normal automation
	id, err := CreateNewAutomation(
		"name_once",
		"description_once",
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
	// Create a manual cron expression
	cronExpr, err := automation.GenerateCronExpression(uint8(then.Hour()), uint8(then.Minute()), []uint8{0, 1, 2, 3, 4, 5, 6})
	assert.NoError(t, err)

	// Disable the automation once
	assert.NoError(t, ModifyAutomationById(id, database.AutomationData{
		Name:                  "name_once",
		Description:           "description_once",
		TriggerCronExpression: cronExpr,
		HomescriptId:          "test",
		Enabled:               true,
		DisableOnce:           true,
		TimingMode:            database.TimingNormal,
	}))

	// Check if the `DisableOnce` boolean has been set to `true`
	automationDb, found, err := GetUserAutomationById("admin", id)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.True(t, automationDb.DisableOnce)

	invalid := false
	for i := 0; i < 9; i++ {
		time.Sleep(time.Second * 10)
		switchItem, found, err := database.GetSwitchById("test_switch")
		if err != nil {
			t.Error(err.Error())
			return
		}
		assert.True(t, found, "Switch not found")
		if switchItem.PowerOn {
			invalid = true
			break
		}
	}
	if invalid {
		t.Error("Power of `test_switch` changed but should have not")
		return
	}

	// Check if the `DisableOnce` boolean has reset to `false`
	automationDb, found, err = GetUserAutomationById("admin", id)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.False(t, automationDb.DisableOnce)

	// Check if the automation runs the second time
	now = time.Now()
	then = now.Add(time.Minute)
	TestInit(t)
	// Create a manual cron expression
	cronExpr, err = automation.GenerateCronExpression(uint8(then.Hour()), uint8(then.Minute()), []uint8{0, 1, 2, 3, 4, 5, 6})
	assert.NoError(t, err)
	// Update the next run-time
	assert.NoError(t, ModifyAutomationById(id, database.AutomationData{
		Name:                  "name_once",
		Description:           "description_once",
		TriggerCronExpression: cronExpr,
		HomescriptId:          "test",
		Enabled:               true,
		DisableOnce:           false,
		TimingMode:            database.TimingNormal,
	}))

	// Toggle the switch to off
	assert.NoError(t, hardware.SetSwitchPowerAll("test_switch", false, "admin"))

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
		t.Error("Power of `test_switch` did not change but should have")
		return
	}
}

// Tests if the automation system can be initialized
func TestInit(t *testing.T) {
	if err := createMockData(); err != nil {
		t.Error(err.Error())
		return
	}

	if err := InitAutomations(); err != nil {
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
