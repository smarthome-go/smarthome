package main

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/MikMuellerDev/smarthome/core/config"
	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/core/hardware"
	"github.com/MikMuellerDev/smarthome/core/homescript"
	"github.com/MikMuellerDev/smarthome/core/scheduler/automation"
	"github.com/MikMuellerDev/smarthome/core/scheduler/scheduler"
	"github.com/MikMuellerDev/smarthome/core/user"
	"github.com/MikMuellerDev/smarthome/core/utils"
	"github.com/MikMuellerDev/smarthome/server/api"
	"github.com/MikMuellerDev/smarthome/server/middleware"
	"github.com/MikMuellerDev/smarthome/server/routes"
	"github.com/MikMuellerDev/smarthome/server/templates"
	"github.com/MikMuellerDev/smarthome/services/camera"
	"github.com/sirupsen/logrus"
)

func TestServer(t *testing.T) {
	// Create logger
	logLevel := logrus.TraceLevel
	if newLogLevel, newLogLevelOk := os.LookupEnv("SMARTHOME_LOG_LEVEL"); newLogLevelOk {
		switch newLogLevel {
		case "TRACE":
			logLevel = logrus.TraceLevel
		case "DEBUG":
			logLevel = logrus.DebugLevel
		case "INFO":
			logLevel = logrus.InfoLevel
		case "WARN":
			logLevel = logrus.WarnLevel
		case "ERROR":
			logLevel = logrus.ErrorLevel
		case "FATAL":
			logLevel = logrus.FatalLevel
		default:
			fmt.Printf("Invalid log level from environment variable: '%s'. Using TRACE\n", newLogLevel)
		}
	}
	log, err := utils.NewLogger(logLevel)
	if err != nil {
		panic(err.Error())
	}

	// Initialize <module> loggers
	config.InitLogger(log)
	camera.InitLogger(log)
	database.InitLogger(log)
	middleware.InitLogger(log)
	api.InitLogger(log)
	routes.InitLogger(log)
	templates.InitLogger(log)
	user.InitLogger(log)
	hardware.InitLogger(log)
	event.InitLogger(log)
	homescript.InitLogger(log)
	automation.InitLogger(log)
	scheduler.InitLogger(log)

	// Initialize database, try 5 times before giving up
	var dbErr error = nil
	for i := 0; i <= 5; i++ {
		dbErr = database.Init(database.DatabaseConfig{
			Username: "smarthome",
			Password: "testing",
			Hostname: "localhost",
			Database: "smarthome",
			Port:     3330,
		}, "admin")
		if dbErr == nil {
			break
		} else {
			log.Warn("Failed to connect to database, retrying in 2 seconds")
			time.Sleep(time.Second * 5)
		}
	}

	if dbErr != nil {
		log.Error("Failed to connect to database after 5 retries, exiting now")
		panic(dbErr.Error())
	}

	// Run setup file if it exists, nil is passed because the file should be read from disk
	if err := config.RunSetup(&config.Setup{}); err != nil {
		log.Fatal("Could not run setup: ", err.Error())
	}

	if err := database.FlushAllLogs(); err != nil {
		log.Fatal("Failed to flush logs: ", err.Error())
	}

	// Always flush old logs
	log.Info("Flushing logs older than 30 days")
	if err := database.FlushOldLogs(); err != nil {
		log.Fatal("Failed to flush logs older that 30 days: ", err.Error())
	}

	if err := database.SetAutomationSystemActivation(true); err != nil {
		t.Error(err.Error())
	}
	automation.Init() // Initializes the automation scheduler
	scheduler.Init()  // Initializes the normal scheduler

	r := routes.NewRouter()
	middleware.Init(true)
	templates.LoadTemplates("./web/html/**/*.html")
	http.Handle("/", r)
}
