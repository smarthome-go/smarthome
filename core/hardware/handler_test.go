package hardware

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/event"
)

func TestSetPower(t *testing.T) {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	InitLogger(log)
	event.InitLogger(log)
	if err := initDB(true); err != nil {
		t.Error(err.Error())
		return
	}
	if err := database.CreateHardwareNode(database.HardwareNode{
		Name:    "test",
		Online:  true,
		Enabled: true,
		Url:     "http://localhost",
		Token:   "",
	}); err != nil {
		t.Error(err.Error())
		return
	}
	table := []struct {
		Switch string
		Power  bool
		Error  string
	}{
		{
			Switch: "test",
			Power:  true,
			// Only the first request will throw an error due to node being marked as offline
			Error: `Post "http://localhost/power?token=": dial tcp`, // Different on other machines
		},
		{
			Switch: "test",
			Power:  false,
			Error:  ``,
		},
		{
			Switch: "test2",
			Power:  true,
			Error:  ``,
		},
		{
			Switch: "test2",
			Power:  false,
			Error:  ``,
		},
	}
	// Create a test room
	if err := database.CreateRoom("test", "test", "test"); err != nil {
		t.Error("Failed to create room:", err.Error())
		return
	}
	for _, req := range table {
		if err := database.CreateSwitch(req.Switch, req.Switch, "test", 0); err != nil {
			t.Error(err.Error())
			return
		}
		if err := SetPower(req.Switch, req.Power); err != nil {
			if !strings.Contains(err.Error(), req.Error) || req.Error == "" {
				t.Errorf("Unexpected error: want: `%s` got: `%s`", req.Error, err.Error())
				return
			}
		} else if req.Error != "" {
			t.Errorf("Expected error: want: `%s` got: `%s`", req.Error, "")
			return
		}
		powerState, err := GetPowerState(req.Switch)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if powerState != req.Power {
			t.Errorf("Power state unaffected: want: `%t` got: `%t`", req.Power, powerState)
			return
		}
	}
	// When no node is registered, the request should not fail
	if err := database.DeleteHardwareNode("http://localhost"); err != nil {
		t.Error(err.Error())
		return
	}
	if err := SetPower("test", false); err != nil {
		t.Error(err.Error())
		return
	}
}
