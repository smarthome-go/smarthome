package automation

import (
	"os"
	"testing"
	"time"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/sirupsen/logrus"
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
	if !doesExists {
		// Create Homescript
		if err := database.CreateNewHomescript(database.Homescript{
			Id:                  "test",
			Owner:               "admin",
			Name:                "Testing",
			Description:         "A Homescript for testing purposes",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "log('automation_trigger', '', 0)",
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
		initDB()
	}
	return nil
}

func TestInit(t *testing.T) {
	if err := Init(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestDeactivate(t *testing.T) {
	TestInit(t) // Initialize the system first
	if err := DeactivateAutomationSystem(); err != nil {
		t.Error(err.Error())
	}
}

func TestActivate(t *testing.T) {
	TestInit(t)
	if err := ActivateAutomationSystem(); err != nil {
		t.Error(err.Error())
	}
}
