package homescript

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/automation"
)

func removeAllAutomations(t *testing.T) {
	automations, err := database.GetAutomations()
	assert.NoError(t, err)

	for _, autom := range automations {
		assert.NoError(t, RemoveAutomation(autom.Id))
	}
}

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
				Code:                "STORAGE.set('AUTOMATION_RAN', true);",
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

	removeAllAutomations(t)
	TestInit(t)

	err := database.InsertHmsStorageEntry("admin", "AUTOMATION_RAN", "false")
	assert.NoError(t, err)

	var hour uint = uint(then.Hour())
	var minute uint = uint(then.Minute())
	days := []uint8{0, 1, 2, 3, 4, 5, 6}

	// Normal automation
	if _, err := CreateNewAutomation(
		"Name",
		"Description",
		"test",
		"admin",
		true,
		&hour,
		&minute,
		&days,
		database.TriggerCron,
		nil,
	); err != nil {
		t.Error(err.Error())
		return
	}
	valid := false
	for i := 0; i < 7; i++ {
		time.Sleep(time.Second * 10)
		storageMap, err := database.GetPersonalHomescriptStorage("admin")
		if err != nil {
			t.Error(err.Error())
			return
		}

		value := storageMap["AUTOMATION_RAN"]

		if value == "true" {
			valid = true
			break
		}
	}

	if !valid {
		t.Error("Automation did not report its success")
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

	removeAllAutomations(t)
	TestInit(t)

	err := database.InsertHmsStorageEntry("admin", "AUTOMATION_RAN", "false")
	assert.NoError(t, err)

	var hour uint = uint(then.Hour())
	var minute uint = uint(then.Minute())
	days := []uint8{0, 1, 2, 3, 4, 5, 6}

	// Normal automation
	id, err := CreateNewAutomation(
		"Name",
		"Description",
		"test",
		"admin",
		true,
		&hour,
		&minute,
		&days,
		database.TriggerCron,
		nil,
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
	if err := ModifyAutomationById(id,
		database.AutomationData{
			Name:                  "name",
			Description:           "description",
			TriggerCronExpression: &cronExpression,
			HomescriptId:          "test",
			Enabled:               true,
			DisableOnce:           false,
			Trigger:               database.TriggerCron,
		}); err != nil {
		t.Error(err.Error())
		return
	}
	// Wait for approx. a minute in order to determine if the modification succeeded
	valid := false
	for i := 0; i < 7; i++ {
		time.Sleep(time.Second * 10)
		storageMap, err := database.GetPersonalHomescriptStorage("admin")
		if err != nil {
			t.Error(err.Error())
			return
		}

		value := storageMap["AUTOMATION_RAN"]

		if value == "true" {
			valid = true
			break
		}
	}
	// Check if the updated script did change the power
	if !valid {
		t.Error("Automation did not report its success")
		return
	}
}

// Creates a regular automation which is then modified to be disabled
// Checks if the automation still executes despite being disabled
func TestModificationToAbort(t *testing.T) {
	now := time.Now()
	then := now.Add(time.Minute)

	removeAllAutomations(t)
	TestInit(t)

	err := database.InsertHmsStorageEntry("admin", "AUTOMATION_RAN", "false")
	assert.NoError(t, err)

	var hour uint = uint(then.Hour())
	var minute uint = uint(then.Minute())
	days := []uint8{0, 1, 2, 3, 4, 5, 6}

	// Normal automation
	id, err := CreateNewAutomation(
		"Name",
		"Description",
		"test",
		"admin",
		true,
		&hour,
		&minute,
		&days,
		database.TriggerCron,
		nil,
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	// Gets the initial automation in order to copy its cron-expression
	// The old cron-expression is required in order to preserve the time window in
	// which the automation could be executed if the modification fails
	automation, found, err := database.GetAutomationById(id)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Automation %d not found in database", id)
		return
	}
	// Set its activation status to `disabled`
	if err := ModifyAutomationById(id,
		database.AutomationData{
			Name:                  "name",
			Description:           "description",
			TriggerCronExpression: automation.Data.TriggerCronExpression,
			HomescriptId:          "test_abort",
			Enabled:               false,
			Trigger:               database.TriggerCron,
		}); err != nil {
		t.Error(err.Error())
		return
	}
	// Wait for approx. a minute until the test can be considered to be successful
	for i := 0; i < 7; i++ {
		time.Sleep(time.Second * 10)
		storageMap, err := database.GetPersonalHomescriptStorage("admin")
		if err != nil {
			t.Error(err.Error())
			return
		}

		value := storageMap["AUTOMATION_RAN"]

		if value == "true" {
			t.Error("Automation ran but was aborted earlier")
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
	var hour uint = uint(then.Hour())
	var minute uint = uint(then.Minute())
	days := []uint8{0, 1, 2, 3, 4, 5, 6}

	removeAllAutomations(t)
	TestInit(t)

	err := database.InsertHmsStorageEntry("admin", "AUTOMATION_RAN", "false")
	assert.NoError(t, err)

	// Normal automation
	_, err = CreateNewAutomation(
		"Name",
		"Description",
		"test",
		"admin",
		false,
		&hour,
		&minute,
		&days,
		database.TriggerCron,
		nil,
	)
	if err != nil {
		t.Error(err.Error())
		return
	}

	// Wait for approx. a minute until to decide that the test was executed successfully
	for i := 0; i < 7; i++ {
		time.Sleep(time.Second * 10)
		storageMap, err := database.GetPersonalHomescriptStorage("admin")
		if err != nil {
			t.Error(err.Error())
			return
		}

		value := storageMap["AUTOMATION_RAN"]

		if value == "true" {
			t.Error("Automation ran but is inactive")
			return
		}
	}
}

// Tests if the different triggers `sunrise` and `sunset` will run at the correct time
func SunRiseSet(t *testing.T) {
	removeAllAutomations(t)
	TestInit(t)

	sunriseId, err := CreateNewAutomation(
		"Name",
		"Description",
		"test",
		"admin",
		true,
		nil,
		nil,
		nil,
		database.TriggerSunrise,
		nil,
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	sunSetId, err := CreateNewAutomation(
		"Name",
		"Description",
		"test",
		"admin",
		true,
		nil,
		nil,
		nil,
		database.TriggerSunset,
		nil,
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

	sunriseJob, err := automationScheduler.FindJobsByTag(fmt.Sprint(sunrise.Id))
	assert.NoError(t, err)

	sunsetJob, err := automationScheduler.FindJobsByTag(fmt.Sprint(sunSet.Id))
	assert.NoError(t, err)

	config, found, err := database.GetServerConfiguration()
	assert.NoError(t, err)
	assert.True(t, found)

	sunriseTime, sunsetTime := automation.CalculateSunRiseSet(config.Latitude, config.Longitude)

	if sunriseJob[0].NextRun().Hour() != int(sunriseTime.Hour) || sunriseJob[0].NextRun().Minute() != int(sunriseTime.Minute) {
		t.Error("Sunrise automation will not run at sunrise")
	}

	if sunsetJob[0].NextRun().Hour() != int(sunsetTime.Hour) || sunriseJob[0].NextRun().Minute() != int(sunsetTime.Minute) {
		t.Error("Sunset automation will not run at sunrise")
	}
}

func TestUserDisabled(t *testing.T) {
	now := time.Now()
	then := now.Add(time.Minute)

	removeAllAutomations(t)
	TestInit(t)

	err := database.InsertHmsStorageEntry("admin", "AUTOMATION_RAN", "false")
	assert.NoError(t, err)

	hour := uint(then.Hour())
	minute := uint(then.Minute())
	days := []uint8{0, 1, 2, 3, 4, 5, 6}

	// Set the user's schedules and automations to off
	assert.NoError(t, database.SetUserSchedulerEnabled("admin", false))

	// Normal automation
	if _, err := CreateNewAutomation(
		"Name",
		"Description",
		"test",
		"admin",
		true,
		&hour,
		&minute,
		&days,
		database.TriggerCron,
		nil,
	); err != nil {
		t.Error(err.Error())
		return
	}
	invalid := false
	for i := 0; i < 7; i++ {
		time.Sleep(time.Second * 10)
		storageMap, err := database.GetPersonalHomescriptStorage("admin")
		if err != nil {
			t.Error(err.Error())
			return
		}

		value := storageMap["AUTOMATION_RAN"]

		if value == "true" {
			invalid = true
			break
		}
	}

	if invalid {
		t.Error("Automation ran but the user has disabled automations")
		return
	}
}

// Tests if the automation system can be initialized
func TestInit(t *testing.T) {
	if err := createMockData(); err != nil {
		t.Error(err.Error())
		return
	}

	config, found, err := database.GetServerConfiguration()
	assert.NoError(t, err)
	assert.True(t, found)

	if err := InitAutomations(config); err != nil {
		t.Error(err.Error())
		return
	}
}

// Deactivates and Reactivates the automation system in order to check for errors
func TestActivate(t *testing.T) {
	TestInit(t)

	config, found, err := database.GetServerConfiguration()
	assert.NoError(t, err)
	assert.True(t, found)

	if err := DeactivateAutomationSystem(config); err != nil {
		t.Error(err.Error())
	}
	if err := ActivateAutomationSystem(config); err != nil {
		t.Error(err.Error())
	}
}
