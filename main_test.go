package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/smarthome-go/smarthome/core/config"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/homescript"
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
	assert.NoError(t, err)

	// Initialize module loggers
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

	// Simulates a typical server startup

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

	// If the connection failed after 5 retries, give up
	assert.NoError(t, dbErr)

	// Run setup file if it exists
	assert.NoError(t, config.RunSetup())

	// Always flush old logs
	assert.NoError(t, event.FlushOldLogs())

	assert.NoError(t, database.SetAutomationSystemActivation(true))

	// Initializes the automation scheduler
	assert.NoError(t, homescript.InitAutomations())

	// Initializes the normal scheduler
	assert.NoError(t, homescript.InitScheduler())

	r := routes.NewRouter()
	middleware.InitWithRandomKey()
	assert.NoError(t, templates.LoadTemplates("./web/dist/html/*.html"))
	http.Handle("/", r)
}
