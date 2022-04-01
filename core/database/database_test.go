package database

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func testDB() error {
	if err := Init(DatabaseConfig{
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

func TestInit(t *testing.T) {
	InitLogger(logrus.New())

}

func TestDeleteTables(t *testing.T) {
	InitLogger(logrus.New())
	if err := testDB(); err != nil {
		t.Error(err.Error())
		return
	}
	if err := DeleteTables(); err != nil {
		t.Error(err.Error())
		return
	}
}
