// Smarthome: A completely self-built Smarthome-system written in Go
// https://github.com/smarthome-go/smarthome
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/utils"
)

// Logger
var log *logrus.Logger

// Default port on which the server listens,
// can be overwritten using the config file or an environment variable
var port uint16 = 8082

type contextKey string

const (
	ShutdownContextKey = contextKey("shutdown")
)

func main() {
	// Do not change the version manually, use the `make version` command instead
	utils.Version = "0.10.0-alpha"

	initLoggers()

	// Read configuration file
	if err := core.ReadConfigFile(); err != nil {
		log.Fatal("Failed to read config file: startup halted: ", err.Error())
	}
	configStruct := core.GetConfig()
	if configStruct.Server.Port != 0 {
		port = configStruct.Server.Port
	}
	log.Debug("Successfully loaded configuration file")

	// Scan environment variables
	newAdminPassword := scanEnv(configStruct)

	// Database connection and initialization
	const retryInterval = 5 // Time to wait before retrying
	var dbErr error         // Saves a potential connection error

	// Allows up to 5 failed connections before quitting
	for i := 0; i <= 5; i++ {
		dbErr = database.Init(configStruct.Database, newAdminPassword)
		if dbErr == nil {
			break // Successfully connected to database
		} else {
			log.Warn(fmt.Sprintf("Failed to connect to database, retrying in %d seconds", retryInterval))
			time.Sleep(retryInterval * time.Second)
		}
	}
	if dbErr != nil {
		// Quit (if 5 attempts failed)
		log.Error(fmt.Sprintf("Failed to connect to database after 5 retries. Please ensure a correct database configuration.\nError: %s", dbErr.Error()))
		os.Exit(1)
	}

	// Setup file
	if err := core.RunSetup(); err != nil {
		log.Error("Could not process setup.json file: ", err.Error())
		os.Exit(1)
	}

	// Obtain the server's configuration
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		log.Error("Could not retrieve server configuration")
		os.Exit(1)
	}

	if err := core.Init(serverConfig); err != nil {
		log.Error("Core init failed: ", err.Error())
		os.Exit(1)
	}

	// Launch webserver
	runWebServer(*configStruct, serverConfig)
}
