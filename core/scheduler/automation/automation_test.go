package automation

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/core/homescript"
)

func TestMain(m *testing.M) {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	InitLogger(log)
	if err := initDB(); err != nil {
		panic(err.Error())
	}
	_, doesExists, err := database.GetUserHomescriptById("test", "admin")
	if err != nil {
		panic(err.Error())
	}
	if err := database.CreateRoom("test_room", "", ""); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch", "", "test_room", 0); err != nil {
		panic(err.Error())
	}
	if !doesExists {
		// Create Homescript
		if err := database.CreateNewHomescript(database.Homescript{
			Id:                  "test",
			Owner:               "admin",
			Name:                "Testing",
			Description:         "A Homescript for testing purposes",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "log('automation_trigger', '', 0); switch('test_switch', on)",
		}); err != nil {
			panic(err.Error())
		}
	}
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

func TestAutomation(t *testing.T) {
	now := time.Now()
	then := now.Add(time.Minute)
	log := logrus.New()
	log.Level = logrus.FatalLevel
	event.InitLogger(log)
	homescript.InitLogger(log)
	// Flush all logs before automation runs
	if err := database.FlushAllLogs(); err != nil {
		t.Error(err.Error())
		return
	}
	if err := Init(); err != nil {
		t.Error(err.Error())
		return
	}
	if err := CreateNewAutomation(
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
	time.Sleep(time.Minute * 2)
	logs, err := database.GetLogs()
	if err != nil {
		t.Error(err.Error())
		return
	}
	valid := false
	for _, l := range logs {
		if l.Name == "automation_trigger" && l.Level == 0 {
			valid = true
		}
	}
	if !valid {
		t.Error("Log from automation not found, could not verify that automation has been executed")
		return
	}
	power, err := database.GetPowerStateOfSwitch("test_switch")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !power {
		t.Error("Power did not change: want: true got: false")
		return
	}
}

func TestInit(t *testing.T) {
	if err := Init(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestActivate(t *testing.T) {
	TestInit(t)
	if err := DeactivateAutomationSystem(); err != nil {
		t.Error(err.Error())
	}
	if err := ActivateAutomationSystem(); err != nil {
		t.Error(err.Error())
	}
}
