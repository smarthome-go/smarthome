// Smarthome: A completely self-built Smarthome-system written in Go
// https://github.com/smarthome-go/smarthome
package main

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core"
	"github.com/smarthome-go/smarthome/core/config"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/core/utils"
	"github.com/smarthome-go/smarthome/server/api"
	"github.com/smarthome-go/smarthome/server/middleware"
	"github.com/smarthome-go/smarthome/server/routes"
	"github.com/smarthome-go/smarthome/server/templates"
	"github.com/smarthome-go/smarthome/services/camera"
	"github.com/smarthome-go/smarthome/services/reminder"
)

// Default port on which the server listens,
// can be overwritten using the config file or an environment variable
var port uint16 = 8082

type contextKey string

const (
	ShutdownContextKey = contextKey("shutdown")
)

func main() {
	// Do not change the version manually, use the `make version` command instead
	utils.Version = "0.8.0"

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
	core.InitLogger(log)
	camera.InitLogger(log)
	middleware.InitLogger(log)
	api.InitLogger(log)
	routes.InitLogger(log)
	templates.InitLogger(log)
	reminder.InitLogger(log)

	// Read configuration file
	if err := config.ReadConfigFile(); err != nil {
		log.Fatal("Failed to read config file: startup halted: ", err.Error())
	}
	configStruct := config.GetConfig()
	if configStruct.Server.Port != 0 {
		port = configStruct.Server.Port
	}
	log.Debug("Successfully loaded configuration file")

	// Scan environment variables
	/*
		`SMARTHOME_ADMIN_PASSWORD`: (String) If specified, the admin user that is created on first launch will receive this password instead of `admin`
		`SMARTHOME_ENV_PRODUCTION`: (Bool  ) Whether the server should use production presets
		`SMARTHOME_SESSION_KEY`   : (String) (Only during production) Specifies a manual key for session encryption (used for larger instances): random key generation is skipped
		`SMARTHOME_DB_DATABASE`   : (String) Sets the database name
		`SMARTHOME_DB_HOSTNAME`   : (String) Sets the database hostname
		`SMARTHOME_DB_PORT`       : (Int   ) Sets the database port
		`SMARTHOME_DB_PASSWORD`   : (String) Sets the database user's password
		`SMARTHOME_DB_USER`       : (String) Sets the database user
	*/
	// Admin passord
	newAdminPassword := "admin"
	if adminPassword, adminPasswordOk := os.LookupEnv("SMARTHOME_ADMIN_PASSWORD"); adminPasswordOk {
		newAdminPassword = adminPassword
	}
	// Operational mode
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
	// Web server session-key
	if sessionKey, sessionKeyOk := os.LookupEnv("SMARTHOME_SESSION_KEY"); sessionKeyOk {
		if !configStruct.Server.Production {
			log.Warn("Using manually specified session encryption key during development mode. This will have no effect unless using production")
		} else {
			if configStruct.Server.SessionKey != "" {
				log.Debug("Selected SMARTHOME_SESSION_KEY over value from config file")
			} else {
				log.Info("Using manually specified session encryption key from SMARTHOME_SESSION_KEY")
			}
			configStruct.Server.SessionKey = sessionKey
		}
	}
	// DB variables
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
	if dbPort, dbPortOk := os.LookupEnv("SMARTHOME_DB_PORT"); dbPortOk {
		portInt, err := strconv.Atoi(dbPort)
		if err != nil || portInt > math.MaxUint16 || portInt < 0 {
			log.Warn("Could not parse `SMARTHOME_DB_PORT` to uint16, using value from config.json")
		} else {
			log.Debug("Selected SMARTHOME_DB_PORT over value from config file")
			configStruct.Database.Port = uint16(portInt)
		}
	}
	// Port of the webserver
	if webPort, webPortOk := os.LookupEnv("SMARTHOME_PORT"); webPortOk {
		webPortInt, err := strconv.Atoi(webPort)
		if err != nil || webPortInt > math.MaxUint16 || webPortInt < 0 {
			log.Warn("Could not parse `SMARTHOME_PORT` to uint16, using 8082")
		} else {
			log.Debug("Selected `SMARTHOME_PORT` over default")
			port = uint16(webPortInt)
		}
	}

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
		log.Fatal(fmt.Sprintf("Failed to connect to database after 5 retries. Please ensure a correct database configuration.\nError: %s", dbErr.Error()))
	}

	// Setup file
	if err := config.RunSetup(); err != nil {
		log.Fatal("Could not process setup.json file: ", err.Error())
		os.Exit(1)
	}

	// Obtain the server's configuration
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		log.Fatal("Could not retrieve server configuration")
		os.Exit(1)
	}

	// Homescript Manager initialization
	homescript.InitManager()
	// Initialize Homescript URL cache flushing scheduler
	if err := homescript.StartUrlCacheGC(); err != nil {
		log.Fatal("Failed to start Homescript URL cache GC: ", err.Error())
		os.Exit(1)
	}

	// Schedulers
	if err := homescript.InitAutomations(serverConfig); err != nil { // Initializes the automation scheduler
		log.Error("Failed to activate automation system: ", err.Error())
		os.Exit(1)
	}
	if err := homescript.InitScheduler(); err != nil { // Initializes the normal scheduler
		log.Error("Failed to activate scheduler system: ", err.Error())
		os.Exit(1)
	}
	if err := reminder.InitSchedule(); err != nil { // Initialize notification scheduler for reminders
		log.Error("Failed to activate reminder scheduler: ", err.Error())
		os.Exit(1)
	}
	// Hardware handler
	hardware.Init()
	if err := hardware.StartPowerUsageSnapshotScheduler(); err != nil {
		log.Error("Failed to start periodic power usage snapshot scheduler: ", err.Error())
		os.Exit(1)
	}
	// Server, middleware and routes
	r := routes.NewRouter()
	if !configStruct.Server.Production {
		log.Warn("Using default session encryption. This is a security risk and must only be used during development.\nHint: this message should disappear when using `production` mode")
		middleware.InitWithManualKey("")
	} else {
		if configStruct.Server.SessionKey == "" {
			log.Debug("Manual session key is empty, generating random key...")
			middleware.InitWithRandomKey()
		} else {
			middleware.InitWithManualKey(configStruct.Server.SessionKey)
		}
	}
	if err := templates.LoadTemplates("./web/dist/html/*.html"); err != nil {
		log.Fatal("Failed to load HTML templates: ", err.Error())
	}
	http.Handle("/", r)

	// Finish startup and launch web server
	event.Info("System Started", fmt.Sprintf("The Smarthome server completed startup at %s (%.2f seconds).", time.Now().Format(time.ANSIC), time.Since(startTime).Seconds()))
	operatingMode := "development"
	if configStruct.Server.Production {
		operatingMode = "production"
	}

	///// Start the server /////
	errCh := make(chan error)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	// Register a shutdown handler
	ctx := context.Background()
	shutdownCtx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, ShutdownContextKey, shutdownCtx)

	server := http.Server{
		Addr:        fmt.Sprintf(":%d", port),
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}
	server.RegisterOnShutdown(cancel)

	log.Info(fmt.Sprintf("Smarthome v%s is listening on http://localhost:%d using %s mode", utils.Version, port, operatingMode))
	go func() { errCh <- server.ListenAndServe() }()
	go core.RunBootAutomations(serverConfig)

	// Main loop
mainLoop:
	for err == nil {
		select {
		case s := <-sigCh:
			if s == os.Interrupt {
				break mainLoop
			}
		case err = <-errCh:
			// this will also terminate loop execution
		}
	}

	// Shutdown
	{
		log.Warn("System shutting down...")
		signal.Reset(os.Interrupt)

		// Shutdown the webserver
		server.SetKeepAlivesEnabled(false)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := server.Shutdown(ctx); err != nil {
			log.Error(fmt.Sprintf("Shutdown error: `%s`", err.Error()))
		}
		cancel()

		// Wait for any other tasks
		if err := core.Shutdown(serverConfig); err != nil {
			log.Error(fmt.Sprintf("Error(s) occured during shutdown: `%s`", err.Error()))
		}

		log.Info("Shutdown compete")
	}

	if err != nil {
		log.Fatal("Graceful shutdown failed: ", err.Error())
	}
}
