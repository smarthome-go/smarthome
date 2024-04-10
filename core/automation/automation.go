package automation

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/homescript/types"
	"github.com/smarthome-go/smarthome/core/user/notify"
)

type AutomationManager struct {
	// The underlying scheduler which will run all predefined automation jobs
	automationScheduler *gocron.Scheduler
	Hms                 types.Manager
}

var Manager AutomationManager

func InitManager(hms types.Manager, config database.ServerConfig) error {
	Manager = AutomationManager{
		automationScheduler: &gocron.Scheduler{},
		Hms:                 hms,
	}

	Manager.automationScheduler = gocron.NewScheduler(time.Local)
	Manager.automationScheduler.TagsUnique()
	if config.AutomationEnabled {
		if err := Manager.ActivateAutomationSystem(config); err != nil {
			log.Error("Failed to activate automation system: could not activate persistent jobs: ", err.Error())
			return err
		}
		log.Info("Successfully activated automation system")
	} else {
		log.Info("Skipping activation of automation system due to it being disabled")
	}
	Manager.automationScheduler.StartAsync()

	return nil
}

// Activates all jobs saved in the database, will be used when the server was restarted
// If a non-critical error occurs, for example the failure to setup a job, it will be returned
// This function will not cancel directly if an error occurs in order to preserve the automation system's uptime
func (m *AutomationManager) ActivateAutomationSystem(config database.ServerConfig) error {
	automations, err := database.GetAutomations()
	if err != nil {
		log.Error("Failed to activate automation system: database failure whilst starting saved automations: ", err.Error())
		return err
	}

	var error error
	for _, automationItem := range automations {
		if err := m.RegisterAutomation(automationItem.Id, automationItem.Data, config); err != nil {
			// Log the error
			log.Error(fmt.Sprintf("Could not activate automation '%d': invalid cron expression", automationItem.Id))
			event.Error("Automation Activation Failure", fmt.Sprintf("The automation %s could not be activated due to an internal error. Please remove it from the system.", automationItem.Data.Name))

			if _, err := notify.Manager.Notify(
				automationItem.Owner,
				"Automation Activation Failure",
				fmt.Sprintf("The automation %s could not be activated due to an internal error. Please remove it from the system.", automationItem.Data.Name),
				notify.NotificationLevelError,
				false,
			); err != nil {
				log.Error("Failed to notify user about failing automation: ", err.Error())
			}
			error = err
			continue // non-critical error, will only affect this automation
		}
	}

	if error != nil {
		event.Info("Automation System Activated", "Successfully activated saved automations")
		log.Info("Successfully activated saved automations")
	}
	return error
}

// Stops all jobs in the automation scheduler
func (m AutomationManager) DeactivateAutomationSystem(config database.ServerConfig) error {
	automations, err := database.GetAutomations()
	if err != nil {
		log.Error("Failed to deactivate automation system: database failure whilst deactivating automations: ", err.Error())
		return err // This is a critical error which can not be recovered from
	}

	for _, automation := range automations {
		if err := m.UnregisterAutomation(automation.Id, automation.Data, config); err != nil {
			return fmt.Errorf("Could not deactivate automation system: failed to unregister automation %d: %s", automation.Id, err.Error())
		}
	}

	log.Info("Successfully disabled automation system: all jobs were stopped")
	event.Info("Disabled Automation System", "Successfully disabled automation system: all jobs were stopped")
	return nil
}

// Runs all automations of the passed user with the given trigger
func (m AutomationManager) RunAllAutomationsWithTrigger(username string, trigger database.AutomationTrigger, context types.AutomationContext) {
	config, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		log.Error("Could not run automations with certain trigger: server configuration not found or errored")
		return
	}

	if !config.AutomationEnabled {
		log.Debug("Not running automations with trigger, automation system disabled")
		return
	}

	automations, err := GetUserAutomations(username)
	if err != nil {
		log.Error("Could not run all automations with certain trigger: could not get user automations: ", err.Error())
		return
	}

	var wg sync.WaitGroup

	for _, job := range automations {
		if job.Trigger != trigger || !job.Enabled {
			continue
		}

		wg.Add(1)
		go func(jobId uint) {
			AutomationRunnerFunc(jobId, context)
			wg.Done()
		}(job.Id)
	}

	wg.Wait()
}
