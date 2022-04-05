package automation

import (
	"fmt"
	"testing"

	"github.com/MikMuellerDev/smarthome/core/database"
)

func TestCreateAutomation(t *testing.T) {
	TestInit(t)
	id, err := CreateNewAutomation(
		"name",
		"description",
		18,
		56,
		[]uint8{0},
		"test",
		"admin",
		true,
		database.TimingNormal,
	)
	if err != nil {
		fmt.Println(err.Error())
		t.Error(err.Error())
		return
	}
	fromDb, found, err := database.GetAutomationById(id)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Automation %d not found after creation", id)
		return
	}
	if fromDb.Name != "name" ||
		fromDb.Description != "description" ||
		!fromDb.Enabled ||
		fromDb.Owner != "admin" {
		t.Errorf("Automation %d has invalid metadata", id)
	}
}

func TestModifyAutomation(t *testing.T) {
	TestInit(t)
	id, err := CreateNewAutomation(
		"name",
		"description",
		18,
		56,
		[]uint8{0},
		"test",
		"admin",
		true,
		database.TimingNormal,
	)
	if err != nil {
		fmt.Println(err.Error())
		t.Error(err.Error())
		return
	}
	if err := ModifyAutomationById(id, database.AutomationWithoutIdAndUsername{
		Name:           "name2",
		Description:    "description2",
		CronExpression: "* * * * *",
		HomescriptId:   "test",
		Enabled:        false,
		TimingMode:     database.TimingNormal,
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
		t.Errorf("Automation with id %d not found", id)
		return
	}
	if temp.Name == "name2" && temp.Description == "description2" && temp.Enabled && temp.Owner == "admin" && temp.CronExpression == "* * * * *" {
		t.Errorf("invalid metadata of modified automation. Got: (Name: %s, Desc: %s, Enabled: %t, Cron: %s) | Want: (`name2`, `description2`, `true`, `* * * * *`)", temp.Name, temp.Description, temp.Enabled, temp.CronExpression)
	}
}

func TestRemoveAutomation(t *testing.T) {
	TestInit(t)
	id, err := CreateNewAutomation(
		"name",
		"description",
		18,
		56,
		[]uint8{0},
		"test",
		"admin",
		true,
		database.TimingNormal,
	)
	if err != nil {
		fmt.Println(err.Error())
		t.Error(err.Error())
		return
	}
	_, found, err := database.GetAutomationById(id)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Automation %d not found after creation", id)
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
		t.Errorf("Automation %d still found after deletion", id)
		return
	}
}

func TestGetUserAutomations(t *testing.T) {
	TestInit(t)
	for i := 0; i < 100; i++ {
		if _, err := CreateNewAutomation(
			"name",
			"description",
			1,
			1,
			[]uint8{0},
			"test",
			"admin",
			true,
			database.TimingNormal,
		); err != nil {
			fmt.Println(err.Error())
			t.Error(err.Error())
			return
		}
	}
	automations, err := GetUserAutomations("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, item := range automations {
		fromDb, found, err := GetUserAutomationById("admin", item.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !found {
			t.Errorf("Automation with id %d could not be found after creation", item.Id)
			return
		}
		if fromDb.Name != item.Name ||
			fromDb.Description != item.Description ||
			fromDb.CronExpression != item.CronExpression ||
			fromDb.Enabled != item.Enabled ||
			fromDb.HomescriptId != item.HomescriptId ||
			fromDb.TimingMode != item.TimingMode ||
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
