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
	"github.com/MikMuellerDev/smarthome/core/user"
	"github.com/MikMuellerDev/smarthome/core/utils"
	"github.com/MikMuellerDev/smarthome/server/middleware"
	"github.com/MikMuellerDev/smarthome/server/routes"
	"github.com/MikMuellerDev/smarthome/server/templates"
	"github.com/MikMuellerDev/smarthome/services/camera"
	"github.com/sirupsen/logrus"
)

const version = "0.0.4"

var port = 8082

func main() {
	// Create logger
	log, err := utils.NewLogger(logrus.TraceLevel)
	if err != nil {
		panic(err.Error())
	}

	// Initialize <module> loggers
	config.InitLogger(log)
	camera.InitLogger(log)
	database.InitLogger(log)
	middleware.InitLogger(log)
	routes.InitLogger(log)
	templates.InitLogger(log)
	user.InitLogger(log)
	hardware.InitLogger(log)
	event.InitLogger(log)

	// Read config file
	if err := config.ReadConfigFile(); err != nil {
		log.Fatal("Failed to read config file: startup halted: ", err.Error())
	}
	configStruct := config.GetConfig()
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

	// Initialize database
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
		log.Error("Failed to connect to database after 10 retries, exiting now")
		panic(dbErr.Error())
	}

	// Run a potential setup file
	if err := config.RunSetup(); err != nil {
		log.Fatal("Could not run setup: ", err.Error())
	}

	if !configStruct.Server.Production {
		// If the server is in development mode, all logs should be flushed
		if err := database.FlushAllLogs(); err != nil {
			log.Fatal("Failed to flush logs: ", err.Error())
		}
	}

	// Flush old logs
	log.Info("Flushing logs older than 30 days")
	if err := database.FlushOldLogs(); err != nil {
		log.Fatal("Failed to flush logs older that 30 days: ", err.Error())
	}

	// TODO: remove this
	camera.TestImageProxy()

	r := routes.NewRouter()
	middleware.Init(configStruct.Server.Production)
	templates.LoadTemplates("./web/html/**/*.html")
	http.Handle("/", r)
	go event.Info("System Started", "The Smarthome server completed startup.")
	log.Info(fmt.Sprintf("Smarthome v%s is running.", version))
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}

// TODO: make a separate logging module which would eliminate the need to initialize a new logger for each module
