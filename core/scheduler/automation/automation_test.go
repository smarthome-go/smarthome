package automation

import (
	"testing"

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
		initDB()
	}
	return nil
}

func TestInit(t *testing.T) {
	InitLogger(logrus.New())
	if err := initDB(); err != nil {
		t.Error(err.Error())
	}
	if err := Init(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestDeactivate(t *testing.T) {
	InitLogger(logrus.New())
	TestInit(t)
	if err := initDB(); err != nil {
		t.Error(err.Error())
	}
	if err := DeactivateAutomationSystem(); err != nil {
		t.Error(err.Error())
	}
}

func TestActivate(t *testing.T) {
	InitLogger(logrus.New())
	TestInit(t)
	if err := initDB(); err != nil {
		t.Error(err.Error())
	}
	if err := ActivateAutomationSystem(); err != nil {
		t.Error(err.Error())
	}
}
