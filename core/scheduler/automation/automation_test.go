package automation

import (
	"testing"
	"time"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/sirupsen/logrus"
)

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
