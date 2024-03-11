package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/smarthome/core"
	"github.com/smarthome-go/smarthome/core/automation"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/device/driver"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/user/notify"
	"github.com/smarthome-go/smarthome/core/utils"
	"github.com/smarthome-go/smarthome/server/api"
	"github.com/smarthome-go/smarthome/server/middleware"
	"github.com/smarthome-go/smarthome/server/routes"
	"github.com/smarthome-go/smarthome/server/templates"
	"github.com/smarthome-go/smarthome/services/camera"
	"github.com/smarthome-go/smarthome/services/reminder"
)

// Initializes logging configuration
func initLoggers() {
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
	logTemp, err := utils.NewLogger(logLevel)
	if err != nil {
		panic(err.Error())
	}

	// Initialize module loggers
	log = logTemp
	core.InitLoggers(log)
	automation.InitLogger(log)
	notify.InitLogger(log)
	camera.InitLogger(log)
	middleware.InitLogger(log)
	api.InitLogger(log)
	routes.InitLogger(log)
	templates.InitLogger(log)
	reminder.InitLogger(log)
	driver.InitLogger(log)
}

const httpRootPath = "/"

func runWebServer(configStruct core.Config, serverConfig database.ServerConfig) {
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
		log.Error("Failed to load HTML templates: ", err.Error())
		os.Exit(1)
	}
	http.Handle(httpRootPath, r)

	// Finish startup and launch web server
	event.Info("System Started", fmt.Sprintf("The Smarthome server completed startup at %s", time.Now().Format(time.ANSIC)))
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
	for {
		select {
		case s := <-sigCh:
			if s == os.Interrupt {
				break mainLoop
			}
		case err := <-errCh:
			log.Error("Web server failed: ", err.Error())
			break
		}
	}

	// Shutdown
	{
		log.Info("System shutting down...")
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
			log.Fatal(fmt.Sprintf("Graceful shutdown failed: `%s`", err.Error()))
		}

		log.Info("Shutdown compete")
		os.Exit(0)
	}
}
