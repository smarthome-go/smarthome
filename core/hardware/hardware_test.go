package hardware

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
)

func TestMain(m *testing.M) {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	InitLogger(log)
	event.InitLogger(log)
	if err := initDB(true); err != nil {
		panic(err.Error())
	}
	// Create a room for some tests
	if err := database.CreateRoom(database.RoomData{Id: "testing", Name: "testing", Description: "testing"}); err != nil {
		panic(err.Error())
	}
	Init() // For initializing atomic slice
	code := m.Run()
	os.Exit(code)
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

func TestPower(t *testing.T) {
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

// Requests
func TestCheckNodeOnline(t *testing.T) {
	if err := checkNodeOnline(database.HardwareNode{
		Name:    "test",
		Online:  true,
		Enabled: true,
		Url:     "https://example.com",
		Token:   "",
	}); err != nil {
		t.Error("Node check failed: ", err.Error())
	}
}

func TestSendPowerRequest(t *testing.T) {
	table := []struct {
		Node  database.HardwareNode
		Error bool
	}{
		{
			Node: database.HardwareNode{
				Name:    "test1",
				Online:  true,
				Enabled: true,
				Url:     "http://1.1.1.1:1",
				Token:   "",
			},
			Error: true,
		},
		{
			Node: database.HardwareNode{
				Name:    "test2",
				Online:  false,
				Enabled: true,
				Url:     "http://1.1.1.1:2",
				Token:   "",
			},
			Error: true,
		},
		{
			Node: database.HardwareNode{
				Name:    "test3",
				Online:  true,
				Enabled: false,
				Url:     "http://1.1.1.1:3",
				Token:   "",
			},
			Error: true,
		},
	}
	for _, item := range table {
		if got := sendPowerRequest(item.Node, "", false); got != nil {
			if !item.Error {
				t.Errorf("Node: %s Error is not expected: want: '', got %s", item.Node.Name, got.Error())
				return
			}
		} else {
			if item.Error {
				t.Errorf("Node: %s Expected error which did not occur", item.Node.Name)
			}
		}
	}
}
