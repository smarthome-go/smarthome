package main

import (
	"fmt"

	"github.com/MikMuellerDev/smarthome/database"
	"github.com/MikMuellerDev/smarthome/utils"
	"github.com/sirupsen/logrus"
)

const version = "0.0.1"

func main() {
	// Create new logger
	log, err := utils.NewLogger(logrus.TraceLevel)
	if err != nil {
		panic(err.Error())
	}
	// Initialize <module> loggers
	utils.InitLogger(log)
	database.InitLogger(log)
	log.Trace("Logging initialized.")

	// Read config file
	err = utils.ReadConfigFile()
	if err != nil {
		log.Fatal("Failed to read config file: startup halted.")
	}
	config := utils.GetConfig()

	// Initialize database
	database.Init(config.Database)

	log.Info(fmt.Sprintf("Smarthome v%s is running.", version))
}
