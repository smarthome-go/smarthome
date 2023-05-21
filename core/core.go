package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/smarthome/core/config"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/core/user"
)

var log *logrus.Logger

// Maximum time to wait until everyting is shutdown
const SHUTDOWN_TIMEOUT = time.Second * 20

// Initialize core loggers
func InitLogger(logger *logrus.Logger) {
	log = logger

	config.InitLogger(log)
	homescript.InitLogger(log)
	database.InitLogger(log)
	hardware.InitLogger(log)
	event.InitLogger(log)
	user.InitLogger(log)
	log.Trace("Core loggers initialized")
}

func waitForHomescripts(wg *sync.WaitGroup) {
	for len(homescript.HmsManager.GetJobList()) > 0 {
		time.Sleep(time.Millisecond * 100)
		log.Trace(fmt.Sprintf("Waiting for %d Homescripts to finish execution...", len(homescript.HmsManager.GetJobList())))
	}
	// Signal that the HMS wait task is finished
	wg.Done()
}

func waitForPowerJobs(wg *sync.WaitGroup) {
	for hardware.GetPendingJobCount() > 0 {
		time.Sleep(time.Millisecond * 100)
		log.Trace(fmt.Sprintf("Waiting for %d power jobs to finish...", hardware.GetPendingJobCount()))
	}
	// Signal that the power wait task is finished
	wg.Done()
}

func Shutdown() error {
	// TODO: introduce a deadline for the background tasks
	var wg sync.WaitGroup
	var error error

	// Shutdown automations
	if err := homescript.DeactivateAutomationSystem(); err != nil {
		error = err
	}

	// HMS jobs
	wg.Add(1)
	go waitForHomescripts(&wg)

	// Power jobs
	wg.Add(1)
	go waitForPowerJobs(&wg)

	// Take a power snapshot
	if err := hardware.SaveCurrentPowerUsage(); err != nil && error != nil {
		error = err
	}

	// Fait for all jobs
	wg.Wait()

	log.Debug("All core shutdown tasks have finished")
	event.Info("System Shutdown", "System shutdown completed")
	return error
}
