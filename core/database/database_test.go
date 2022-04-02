package database

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	InitLogger(log)
	if err := initDB(true); err != nil {
		panic(err.Error())
	}
	// Create a test homescript for some tests
	if err := CreateNewHomescript(Homescript{
		Id:    "test",
		Owner: "admin",
	}); err != nil {
		panic(err.Error())
	}
	// Create test user for some tests
	if err := AddUser(FullUser{
		Username: "testing",
	}); err != nil {
		panic(err.Error())
	}
	code := m.Run()
	os.Exit(code)
}

func initDB(args ...bool) error {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	InitLogger(log)
	if err := Init(DatabaseConfig{
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
		if err := DeleteTables(); err != nil {
			return err
		}
		time.Sleep(time.Second)
		initDB()
	}
	return nil
}
