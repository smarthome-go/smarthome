package automation_test

import (
	"testing"

	"github.com/smarthome-go/smarthome/core/automation"
	"github.com/smarthome-go/smarthome/core/database"
)

func CreateAutomationTest(t *testing.T) {
	TestInit(t)

	var hour uint = 2
	var minute uint = 42
	days := []uint8{3, 1, 4}

	id, err := automation.Manager.CreateNewAutomation(
		"name",
		"description",
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
	fromDb, found, err := database.GetAutomationById(id)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Automation '%d' not found after creation", id)
		return
	}
	if fromDb.Data.Name != "name" ||
		fromDb.Data.Description != "description" ||
		fromDb.Data.Enabled ||
		fromDb.Owner != "admin" {
		t.Errorf("Automation '%d' has invalid metadata", id)
	}
}

func TestModifyAutomation(t *testing.T) {
	TestInit(t)
	var hour uint = 2
	var minute uint = 42
	days := []uint8{3, 1, 4}

	id, err := automation.Manager.CreateNewAutomation(
		"name",
		"description",
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
	cronExpression1 := "* * * * *"
	if err := automation.Manager.ModifyAutomationById(id, database.AutomationData{
		Name:                  "name2",
		Description:           "description2",
		TriggerCronExpression: &cronExpression1,
		HomescriptId:          "test",
		Enabled:               true,
		Trigger:               database.TriggerCron,
	}); err != nil {
		t.Error(err.Error())
		return
	}
	temp, found, err := automation.GetUserAutomationById("admin", id)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Automation '%d' not found", id)
		return
	}
	if temp.Name != "name2" ||
		temp.Description != "description2" ||
		*temp.TriggerCronExpression != "* * * * *" ||
		!temp.Enabled ||
		temp.Trigger != database.TriggerCron {
		t.Errorf("invalid metadata of modified automation. Want: (`name2`, `description2`, `true`, `* * * * *`) | Got: (Name: %s, Desc: %s, Enabled: %t, Cron: %s)", temp.Name, temp.Description, temp.Enabled, *temp.TriggerCronExpression)
		return
	}
}

// Test if the deletion signal is correctly sent to the database
// For actual execution tests, have a look at `automation_test.go`
func TestRemoveAutomation(t *testing.T) {
	TestInit(t)
	var hour uint = 2
	var minute uint = 42
	days := []uint8{3, 1, 4}

	id, err := automation.Manager.CreateNewAutomation(
		"name",
		"description",
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
	_, found, err := database.GetAutomationById(id)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Automation '%d' not found after creation", id)
		return
	}
	if err := automation.Manager.RemoveAutomation(id); err != nil {
		t.Error(err)
		return
	}
	_, found, err = database.GetAutomationById(id)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if found {
		t.Errorf("Automation '%d' still found after deletion", id)
		return
	}
}

func TestGetUserAutomations(t *testing.T) {
	TestInit(t)
	var hour uint = 2
	var minute uint = 42
	days := []uint8{3, 1, 4}
	for i := 0; i < 100; i++ {
		_, err := automation.Manager.CreateNewAutomation(
			"name",
			"description",
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
			t.Error(err)
			return
		}
	}
	automations, err := database.GetUserAutomations("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	// Matches every existent automation against the return value of `GetUserAutomationById`
	for _, item := range automations {
		fromDb, found, err := automation.GetUserAutomationById("admin", item.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !found {
			t.Errorf("Automation '%d' could not be found after creation", item.Id)
			return
		}
		if fromDb.Name != item.Data.Name ||
			fromDb.Description != item.Data.Description ||
			*fromDb.TriggerCronExpression != *item.Data.TriggerCronExpression ||
			fromDb.Enabled != item.Data.Enabled ||
			fromDb.HomescriptId != item.Data.HomescriptId ||
			fromDb.Trigger != item.Data.Trigger ||
			fromDb.Owner != item.Owner {
			t.Errorf("Adding and retrieving automations failed: values are not equal. want: %v got: %v", item, fromDb)
			return
		}
		if err := automation.Manager.RemoveAutomation(item.Id); err != nil {
			t.Error(err.Error())
			return
		}
	}
}
