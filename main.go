// Smarthome: A completely self-built Smarthome-system written in Go
// https://github.com/MikMuellerDev/smarthome
package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/MikMuellerDev/smarthome/core/config"
	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/core/hardware"
	"github.com/MikMuellerDev/smarthome/core/homescript"
	"github.com/MikMuellerDev/smarthome/core/scheduler/automation"
	"github.com/MikMuellerDev/smarthome/core/user"
	"github.com/MikMuellerDev/smarthome/core/utils"
	"github.com/MikMuellerDev/smarthome/server/api"
	"github.com/MikMuellerDev/smarthome/server/middleware"
	"github.com/MikMuellerDev/smarthome/server/routes"
	"github.com/MikMuellerDev/smarthome/server/templates"
	"github.com/MikMuellerDev/smarthome/services/camera"
	"github.com/sirupsen/logrus"
)

var port = 8082 // Port used during development, can be overridden by config file or environment variables

func main() {
	utils.Version = "0.0.17-beta"

	startTime := time.Now()
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

	// Read config file
	if err := config.ReadConfigFile(); err != nil {
		log.Fatal("Failed to read config file: startup halted: ", err.Error())
	}
	configStruct := config.GetConfig()
	if configStruct.Server.Port != 0 {
		port = int(configStruct.Server.Port)
	}
	log.Debug("Loaded and successfully initialized config")

	// Environment variables
	/*
		`SMARTHOME_ADMIN_PASSWORD`: If set, the admin user that is created on first launch will get this password instead of `admin`
		`SMARTHOME_DB_DATABASE`   : Sets the database name
		`SMARTHOME_DB_HOSTNAME`   : Sets the database hostname
		`SMARTHOME_DB_PASSWORD`   : Sets the database user's password
		`SMARTHOME_DB_USER`       : Sets the database user
	*/

	newAdminPassword := "admin"
	if adminPassword, adminPasswordOk := os.LookupEnv("SMARTHOME_ADMIN_PASSWORD"); adminPasswordOk {
		newAdminPassword = adminPassword
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

	// Initialize database, try 5 times before giving up
	var dbErr error = nil

	for i := 0; i <= 5; i++ {
		dbErr = database.Init(configStruct.Database, newAdminPassword)
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

	// Run setup file if it exists
	if err := config.RunSetup(); err != nil {
		log.Fatal("Could not run setup: ", err.Error())
	}

	// If the server is in development mode, all logs should be flushed
	if !configStruct.Server.Production {
		if err := database.FlushAllLogs(); err != nil {
			log.Fatal("Failed to flush logs: ", err.Error())
		}
	}

	// BEGIN REMOVE ME
	tmp, err := database.CreateNewSchedule(database.Schedule{
		Name:           "test",
		Owner:          "admin",
		Hour:           1,
		Minute:         1,
		HomescriptCode: "print('hello')",
	})
	if err != nil {
		log.Error(err.Error())
	}
	fmt.Println(database.GetScheduleById(tmp))
	fmt.Println(database.GetUserSchedules("admin"))
	fmt.Println(database.GetSchedules())
	if err := database.ModifySchedule(tmp, database.ScheduleWithoudIdAndUsername{
		Name:           "test 2",
		Hour:           2,
		Minute:         2,
		HomescriptCode: "exit(12)",
	}); err != nil {
		log.Error(err.Error())
	}
	fmt.Println(database.DeleteScheduleById(tmp))
	// END REMOVE ME

	// Always flush old logs
	// TODO: move deletion of old logs to a scheduler
	log.Info("Flushing logs older than 30 days")
	if err := database.FlushOldLogs(); err != nil {
		log.Fatal("Failed to flush logs older that 30 days: ", err.Error())
	}

	automation.Init() // Initializes the automation scheduler

	r := routes.NewRouter()
	middleware.Init(configStruct.Server.Production)
	templates.LoadTemplates("./web/html/**/*.html")
	http.Handle("/", r)

	event.Info("System Started", fmt.Sprintf("The Smarthome server completed startup in %.2f seconds", time.Since(startTime).Seconds()))
	log.Info(fmt.Sprintf("Smarthome v%s is running on port %d", utils.Version, port))
	if err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}
