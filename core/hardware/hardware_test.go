package hardware

import (
	"testing"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/sirupsen/logrus"
)

func testDB() error {
	database.InitLogger(logrus.New())
	if err := database.Init(database.DatabaseConfig{
		Username: "smarthome",
		Password: "testing",
		Hostname: "localhost",
		Database: "smarthome",
		Port:     3330,
	}, "admin"); err != nil {
		return err
	}
	return nil
}

func TestPower(t *testing.T) {
	if err := testDB(); err != nil {
		t.Error(err.Error())
	}
	table := []struct {
		Switch string
		Power  bool
	}{
		{"1", true},
		{"1", false},
		{"2", true},
		{"2", false},
		{"3", true},
		{"3", false},
		{"4", true},
		{"4", false},
		{"5", true},
		{"5", false},
		{"6", true},
		{"6", false},
	}
	if err := database.CreateRoom("testing", "testing", "testing"); err != nil {
		t.Error(err.Error())
		return
	}
	for _, item := range table {
		if err := database.CreateSwitch(item.Switch, item.Switch, "testing", 0); err != nil {
			t.Error(err.Error())
			return
		}
		if err := setPowerOnAllNodes(item.Switch, item.Power); err != nil {
			t.Error(err.Error())
			return
		}
		power, err := GetPowerState(item.Switch)
		if err != nil {
			t.Error(err.Error())
		}
		if power != item.Power {
			t.Errorf("Failed to set power: got: %t, want: %t", power, item.Power)
			return
		}
	}
}
