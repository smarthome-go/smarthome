package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/config"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/core/scheduler/automation"
	"github.com/smarthome-go/smarthome/core/scheduler/scheduler"
	"github.com/smarthome-go/smarthome/core/user"
	"github.com/smarthome-go/smarthome/core/utils"
	"github.com/smarthome-go/smarthome/server/api"
	"github.com/smarthome-go/smarthome/server/middleware"
	"github.com/smarthome-go/smarthome/server/routes"
	"github.com/smarthome-go/smarthome/server/templates"
	"github.com/smarthome-go/smarthome/services/camera"
)

func TestServer(t *testing.T) {
	// Create logger
	log, err := utils.NewLogger(logrus.FatalLevel)
	if err != nil {
		t.Error(err.Error())
		return
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
		t.Errorf("Failed to connect to database after 5 retries, exiting now: %s", dbErr.Error())
	}

	// Run setup file if it exists
	if err := config.RunSetup(); err != nil {
		t.Errorf("Could not run setup: %s", err.Error())
	}

	if err := database.FlushAllLogs(); err != nil {
		t.Errorf("Failed to flush logs: %s", err.Error())
	}

	// Always flush old logs
	log.Info("Flushing logs older than 30 days")
	if err := database.FlushOldLogs(); err != nil {
		t.Errorf("Failed to flush logs older that 30 days: %s", err.Error())
	}

	if err := database.SetAutomationSystemActivation(true); err != nil {
		t.Error(err.Error())
	}

	// Initializes the automation scheduler
	if err := automation.Init(); err != nil {
		t.Errorf("Failed to activate automation system: %s", err.Error())
	}
	// Initializes the normal scheduler
	if err := scheduler.Init(); err != nil {
		t.Errorf("Failed to activate scheduler system: %s", err.Error())
	}

	r := routes.NewRouter()
	middleware.Init(true)
	templates.LoadTemplates("./web/dist/html/*.html")
	http.Handle("/", r)
}
