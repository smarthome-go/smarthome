package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/sirupsen/logrus"
)

type ServerConfig struct {
	Production bool   `json:"production"`
	Port       uint16 `json:"port"`
}

// TODO: finish moving config system to setup.json
type Config struct {
	Server   ServerConfig            `json:"server"`
	Database database.DatabaseConfig `json:"database"`
}

var config Config

const configPath = "./data/config/config.json"

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func ReadConfigFile() error {
	// Read file from <configPath> on disk
	// If this file does not exist, create a new blank one
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		configTemp, errCreate := createNewConfigFile()
		if errCreate != nil {
			log.Error("Failed to read config file: ", err.Error())
			log.Fatal("Failed to initialize config: could not read or create a config file: ", errCreate.Error())
			return err
		}
		config = configTemp
		log.Info("Failed to read config file: but managed to create a new config file")
		return nil
	}
	// Parse config file to struct <Config>
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

// Creates an empty config file, can return an error if it fails
func createNewConfigFile() (Config, error) {
	config := Config{
		Server: ServerConfig{
			Production: false,
		},
		Database: database.DatabaseConfig{
			Username: "smarthome",
			Password: "password",
			Hostname: "smarthome-mariadb",
			Database: "smarthome",
			Port:     3306,
		},
	}
	fileContent, err := json.MarshalIndent(config, "", "	")
	if err != nil {
		log.Error("Failed to create config file: creating file content from JSON failed: ", err.Error())
		return Config{}, err
	}
	if err := os.MkdirAll("./data/config", 0644); err != nil {
		log.Error("Failed to create new config file: creating data directory failed: ", err.Error())
		return Config{}, err
	}
	if err = ioutil.WriteFile("data/config/config.json", fileContent, 0644); err != nil {
		log.Error("Failed to write file to disk: ", err.Error())
		return Config{}, err
	}
	return config, nil
}

func GetConfig() Config {
	return config
}
