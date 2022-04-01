package automation

import (
	"fmt"
	"testing"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/sirupsen/logrus"
)

func TestCreateAutomation(t *testing.T) {
	InitLogger(logrus.New())
	TestInit(t)
	if err := initDB(); err != nil {
		t.Error(err.Error())
	}

	if err := CreateNewAutomation(
		"name",
		"description",
		1,
		2,
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
	if len(automations) != 1 {
		t.Errorf("length of user automations after creation is not 1: length: %d", len(automations))
		return
	}
	a := automations[0]
	if a.Name != "name" || a.Description != "description" || !a.Enabled || a.Owner != "admin" {
		t.Error("invalid metadata of created automation")

		// TODO: test timing
	}
}
