package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

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
