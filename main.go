// Smarthome: A completely self-built Smarthome-system written in Go
// https://github.com/smarthome-go/smarthome
package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	"github.com/smarthome-go/smarthome/services/reminder"
)

var port = 8082 // Default port on which the server listens, can be overwritten by config file or the environment variable

func main() {
	// Do not change manually, use the `make version` command instead
	utils.Version = "0.0.38"

	startTime := time.Now()

	// Logging configuration
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
	automation.InitLogger(log)
	scheduler.InitLogger(log)
	reminder.InitLogger(log)

	// Read configuration file
	if err := config.ReadConfigFile(); err != nil {
		log.Fatal("Failed to read config file: startup halted: ", err.Error())
	}
	configStruct := config.GetConfig()
	if configStruct.Server.Port != 0 {
		port = int(configStruct.Server.Port)
	}
	log.Debug("Successfully loaded configuration file")

	// Process environment variables
	/*
		`SMARTHOME_ADMIN_PASSWORD`: (String) If set, the admin user that is created on first launch will get this password instead of `admin`
		`SMARTHOME_ENV_PRODUCTION`: (Bool  ) Whether the server should use production presets
		`SMARTHOME_DB_DATABASE`   : (String) Sets the database name
		`SMARTHOME_DB_HOSTNAME`   : (String) Sets the database hostname
		`SMARTHOME_DB_PORT`       : (Int   ) Sets the database port
		`SMARTHOME_DB_PASSWORD`   : (String) Sets the database user's password
		`SMARTHOME_DB_USER`       : (String) Sets the database user
	*/
	newAdminPassword := "admin"
	if adminPassword, adminPasswordOk := os.LookupEnv("SMARTHOME_ADMIN_PASSWORD"); adminPasswordOk {
		newAdminPassword = adminPassword
	}
	if productionEnvStr, productionEnvStrOk := os.LookupEnv("SMARTHOME_ENV_PRODUCTION"); productionEnvStrOk {
		switch productionEnvStr {
		case "TRUE":
			configStruct.Server.Production = true
			log.Debug("Detected `SMARTHOME_ENV_PRODUCTION` (TRUE), server will start using production presets")
		case "FALSE":
			configStruct.Server.Production = false
			log.Debug("Detected `SMARTHOME_ENV_PRODUCTION` (FALSE), server will start in development mode")
		default:
			log.Warn("Could not use `SMARTHOME_ENV_PRODUCTION` as boolean value, using development mode\nValid modes are `TRUE` and `FALSE`")
		}
	}
	if dbUsername, dbUsernameOk := os.LookupEnv("SMARTHOME_DB_USER"); dbUsernameOk {
		log.Debug("Selected SMARTHOME_DB_USER over value from config file")
		configStruct.Database.Username = dbUsername
	}
	if dbPassword, dbPasswordOk := os.LookupEnv("SMARTHOME_DB_PASSWORD"); dbPasswordOk {
		log.Debug("Selected SMARTHOME_DB_PASSWORD over value from config file")
		configStruct.Database.Password = dbPassword
	}
	if dbDatabase, dbDatabaseOk := os.LookupEnv("SMARTHOME_DB_DATABASE"); dbDatabaseOk {
		log.Debug("Selected SMARTHOME_DB_DATABASE over value from config file")
		configStruct.Database.Database = dbDatabase
	}
	if dbHostname, dbHostnameOk := os.LookupEnv("SMARTHOME_DB_HOSTNAME"); dbHostnameOk {
		log.Debug("Selected SMARTHOME_DB_HOSTNAME over value from config file")
		configStruct.Database.Hostname = dbHostname
	}
	if webPort, webPortOk := os.LookupEnv("SMARTHOME_PORT"); webPortOk {
		webPortInt, err := strconv.Atoi(webPort)
		if err != nil {
			log.Warn("Could not parse `SMARTHOME_PORT` to int, using 8082")
		} else {
			log.Debug("Selected `SMARTHOME_PORT` over default")
			port = webPortInt
		}
	}
	if dbPort, dbPortOk := os.LookupEnv("SMARTHOME_DB_PORT"); dbPortOk {
		portInt, err := strconv.Atoi(dbPort)
		if err != nil {
			log.Warn("Could not parse `SMARTHOME_DB_PORT` to int, using value from config.json")
		} else {
			log.Debug("Selected SMARTHOME_DB_PORT over value from config file")
			configStruct.Database.Port = portInt
		}
	}

	// Database connection and initialization
	const retryInterval = 5 // Time to wait before retrying
	var dbErr error = nil

	// Allows up to 5 failed connections before quitting
	for i := 0; i <= 5; i++ {
		dbErr = database.Init(configStruct.Database, newAdminPassword)
		if dbErr == nil {
			break
		} else {
			log.Warn(fmt.Sprintf("Failed to connect to database, retrying in %d seconds", retryInterval))
			time.Sleep(retryInterval * time.Second)
		}
	}
	if dbErr != nil {
		// Quit if 5 attempts failed
		log.Fatal(fmt.Sprintf("Failed to connect to database after 5 retries. Please ensure a correct database configuration.\nError: %s", dbErr.Error()))
	}

	/** Setup file */
	if err := config.RunSetup(); err != nil {
		log.Fatal("Could not process setup.json file: ", err.Error())
	}

	/** Logs */
	if !configStruct.Server.Production { // If the server is in development mode, all logs should be deleted
		if err := event.FlushAllLogs(); err != nil {
			log.Error("Failed to flush all logs: ", err.Error())
		}
	}
	if err := event.FlushOldLogs(); err != nil { // Always flush old logs
		log.Error("Failed to flush logs older that 30 days: ", err.Error()) // TODO: setup deletion of old logs with a scheduler
	}

	/** Schedulers */
	if err := automation.Init(); err != nil { // Initializes the automation scheduler
		log.Error("Failed to activate automation system: ", err.Error())
	}
	if err := scheduler.Init(); err != nil { // Initializes the normal scheduler
		log.Error("Failed to activate scheduler system: ", err.Error())
	}
	if err := reminder.InitSchedule(); err != nil { // Initialize notification scheduler for reminders
		log.Error("Failed to activate reminder scheduler: ", err.Error())
	}

	/** Hardware handler */
	hardware.Init()

	/** Server, middleware and templates */
	r := routes.NewRouter()
	middleware.Init(configStruct.Server.Production)
	templates.LoadTemplates("./web/dist/html/*.html")
	http.Handle("/", r)

	/** Finish startup */
	event.Info("System Started", fmt.Sprintf("The Smarthome server completed startup at %s (%.2f seconds).", time.Now().Format(time.ANSIC), time.Since(startTime).Seconds()))
	operatingMode := "development"
	if configStruct.Server.Production {
		operatingMode = "production"
	}
	log.Info(fmt.Sprintf("Smarthome v%s is listening on http://localhost:%d using %s mode", utils.Version, port, operatingMode))
	if err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal("Web server failed unexpectedly: ", err.Error())
	}
}
