package automation

import (
	"fmt"
	"testing"

	"github.com/MikMuellerDev/smarthome/core/database"
)

func TestCreateAutomation(t *testing.T) {
	TestInit(t)
	if err := CreateNewAutomation(
		"name",
		"description",
		18,
		56,
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
	automations, err := GetUserAutomations("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	valid := false
	for _, temp := range automations {
		if temp.Name == "name" && temp.Description == "description" && temp.Enabled && temp.Owner == "admin" {
			valid = true
		}
	}
	if !valid {
		t.Error("invalid metadata of created automation")
		return
	}
}

func TestModifyAutomation(t *testing.T) {
	TestInit(t)
	if err := CreateNewAutomation(
		"name",
		"description",
		18,
		56,
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
	automations, err := GetUserAutomations("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(automations) == 0 {
		t.Error("Retrieved automations slice contains no elements")
		return
	}
	testId := automations[0].Id
	if err := ModifyAutomationById(automations[0].Id, database.AutomationWithoutIdAndUsername{
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
	automations, err = GetUserAutomations("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	temp, found, err := GetUserAutomationById("admin", testId)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Automation with id %d not found", testId)
		return
	}
	if temp.Name == "name2" && temp.Description == "description2" && temp.Enabled && temp.Owner == "admin" && temp.CronExpression == "* * * * *" {
		t.Errorf("invalid metadata of modified automation. Got: (Name: %s, Desc: %s, Enabled: %t, Cron: %s) | Want: (`name2`, `description2`, `true`, `* * * * *`)", temp.Name, temp.Description, temp.Enabled, temp.CronExpression)
	}
}

func TestRemoveAutomation(t *testing.T) {
	TestInit(t)
	if err := CreateNewAutomation(
		"name",
		"description",
		18,
		56,
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
	automations, err := GetUserAutomations("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(automations) == 0 {
		t.Error("Automation could not be added: 0 elements in result slice")
	}
	for _, item := range automations {
		if err := RemoveAutomation(item.Id); err != nil {
			t.Error(err.Error())
			return
		}
	}
	automations, err = GetUserAutomations("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(automations) != 0 {
		t.Error("More than 0 elements in result slice after deletion")
	}
}

func TestGetUserAutomations(t *testing.T) {
	TestInit(t)
	for i := 0; i < 100; i++ {
		if err := CreateNewAutomation(
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
			fromDb.Owner != item.Owner ||
			fromDb.CronDescription != "At 01:01 AM, only on Sunday" {
			fmt.Println("Want:", item)
			fmt.Println("Got:", fromDb)
			t.Error("Adding and retrieving automations failed: values are not equal")
			return
		}
		if err := RemoveAutomation(item.Id); err != nil {
			t.Error("Failed to remove created automation: ", err.Error())
			return
		}
	}
}
