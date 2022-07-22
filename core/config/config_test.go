package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/smarthome/core/database"
)

func TestMain(m *testing.M) {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	InitLogger(log)
	if err := initDB(true); err != nil {
		panic(err.Error())
	}
	code := m.Run()
	os.Exit(code)
}

func initDB(args ...bool) error {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	database.InitLogger(log)
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

func TestReadConfig(t *testing.T) {
	configPath = t.TempDir()
	if err := ReadConfigFile(); err != nil {
		t.Error(err.Error())
		return
	}
	_, err := os.Stat(fmt.Sprintf("%s/config.json", configPath))
	if !os.IsNotExist(err) && err != nil {
		t.Errorf("Config file %s does not exist after creation", configPath)
		return
	}
	// Wipe existing config first
	config = Config{}
	// Read the file again to test serialization
	if err := ReadConfigFile(); err != nil {
		t.Error(err.Error())
		return
	}
	configTemp := GetConfig()
	if configTemp.Database.Username != "smarthome" ||
		configTemp.Database.Database != "smarthome" ||
		configTemp.Database.Hostname != "smarthome-mariadb" ||
		configTemp.Database.Password != "password" ||
		configTemp.Database.Port != 3306 {
		t.Errorf("Database configuration after creation is empty: got: %v", configTemp)
		return
	}

	// Write non-json to the file and test if the error is handled
	if err := ioutil.WriteFile(
		fmt.Sprintf("%s/config.json", configPath),
		[]byte("not_valid"),
		0755,
	); err != nil {
		t.Error(err.Error())
		return
	}
	// Read the file again after it has been scrambled
	if err := ReadConfigFile(); err == nil {
		t.Error("Expected error but non was returned")
		return
	}

	// Test if error handling works when the file does not exists and is not writable
	configPath = "/does/not/exists"
	if err := ReadConfigFile(); err == nil {
		t.Error("Expected error but non was returned")
		return
	}
}
