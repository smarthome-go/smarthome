package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/hardware"
)

type Config struct {
	Database database.DatabaseConfig `json:"database"`
	Hardware hardware.HardwareComfig `json:"hardware"`
}

var config Config

const configPath = "./config/config.json"

func ReadConfigFile() error {
	// Read file from <configPath> on disk
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Error("Failed to read config file: ", err.Error())
		return err
	}

	// Parse config file to struct <configFile>
	var configFile Config
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&configFile)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to parse config file at `%s` into Config struct: %s", configPath, err.Error()))
		return err
	}
	config = configFile
	return nil
}

func GetConfig() Config {
	return config
}
