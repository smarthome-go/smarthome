package core

import (
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/services/reminder"
)

func Init(config database.ServerConfig) error {
	// Homescript Manager initialization
	homescript.InitManager()

	// Schedulers
	if err := homescript.InitAutomations(config); err != nil { // Initializes the automation scheduler
		return fmt.Errorf("Failed to activate automation system: %s", err.Error())
	}
	if err := homescript.InitScheduler(); err != nil { // Initializes the normal scheduler
		return fmt.Errorf("Failed to activate scheduler system: %s", err.Error())
	}
	if err := reminder.InitSchedule(); err != nil { // Initialize notification scheduler for reminders
		return fmt.Errorf("Failed to activate reminder scheduler: %s", err.Error())
	}

	// Hardware handler
	hardware.Init()
	if err := hardware.StartPowerUsageSnapshotScheduler(); err != nil {
		return fmt.Errorf("Failed to start periodic power usage snapshot scheduler: %s", err.Error())
	}
	return nil
}
