package homescript

import (
	"testing"

	"github.com/smarthome-go/smarthome/core/database"
)

func TestCreateAutomation(t *testing.T) {
	TestInit(t)

	var hour uint = 2
	var minute uint = 42
	days := []uint8{3, 1, 4}

	id, err := CreateNewAutomation(
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

	id, err := CreateNewAutomation(
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
	if err := ModifyAutomationById(id, database.AutomationData{
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
	temp, found, err := GetUserAutomationById("admin", id)
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

	id, err := CreateNewAutomation(
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
	if err := RemoveAutomation(id); err != nil {
		t.Errorf(err.Error())
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
		_, err := CreateNewAutomation(
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
	automations, err := GetUserAutomations("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	// Matches every existent automation against the return value of `GetUserAutomationById`
	for _, item := range automations {
		fromDb, found, err := GetUserAutomationById("admin", item.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !found {
			t.Errorf("Automation '%d' could not be found after creation", item.Id)
			return
		}
		if fromDb.Name != item.Name ||
			fromDb.Description != item.Description ||
			*fromDb.TriggerCronExpression != *item.TriggerCronExpression ||
			fromDb.Enabled != item.Enabled ||
			fromDb.HomescriptId != item.HomescriptId ||
			fromDb.Trigger != item.Trigger ||
			fromDb.Owner != item.Owner {
			t.Errorf("Adding and retrieving automations failed: values are not equal. want: %v got: %v", item, fromDb)
			return
		}
		if err := RemoveAutomation(item.Id); err != nil {
			t.Error(err.Error())
			return
		}
	}
}
