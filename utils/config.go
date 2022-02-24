package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/MikMuellerDev/smarthome/database"
)

type Config struct {
	Database database.DatabaseConfig `json:"database"`
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

	var configFile Config
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&configFile)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to parse config file at `%s` into Config struct: %s", configPath, err.Error()))
		return err
	}
	config = configFile
	log.Debug("Loaded config file from: ", configPath)
	return nil
}

func GetConfig() Config {
	return config
}
