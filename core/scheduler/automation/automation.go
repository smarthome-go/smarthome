package automation

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"

	"github.com/MikMuellerDev/smarthome/core/database"
)

// The scheduler which will run all predefined automation jobs
var scheduler *gocron.Scheduler

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Activates all jobs saved in the database, will be used when the server was restarted
// If a non-critical error occurs, for example the failure to setup a job, it will be returned
// This function will not cancel directly if an error occurs in order to preserve the automation system's uptime
func ActivateAutomationSystem() error {
	automations, err := database.GetAutomations()
	if err != nil {
		log.Error("Failed to activate automation system: database failure whilst starting saved automations: ", err.Error())
		return err // This is a critical error which can not be recovered
	}
	var activatedItems uint = 0
	for _, automation := range automations {
		if !IsValidCronExpression(automation.CronExpression) {
			log.Error(fmt.Sprintf("Could not activate automation '%d': invalid cron expression", automation.Id))
			continue // non-critical error
		}
		if !automation.Enabled {
			log.Debug(fmt.Sprintf("Skipping activation of automation %d: automation is disabled", automation.Id))
			continue // Skip disabled automations
		}
		automationJob := scheduler.Cron(automation.CronExpression)
		automationJob.Tag(fmt.Sprintf("%d", automation.Id))
		_, err := automationJob.Do(automationRunnerFunc, automation.Id)
		if err != nil {
			log.Error("Failed to register cron job: ", err.Error())
			return err
		}
		activatedItems += 1
		log.Debug(fmt.Sprintf("Successfully activated automation '%d' of user '%s'", automation.Id, automation.Owner))
	}
	log.Debug(fmt.Sprintf("Activated saved automations: registered %d total automation jobs", activatedItems))
	return nil
}

// Stops all jobs in the automation scheduler
func DeactivateAutomationSystem() error {
	automations, err := database.GetAutomations()
	if err != nil {
		log.Error("Failed to deactivate automation system: database failure whilst deactivating automations: ", err.Error())
		return err // This is a critical error which can not be recovered from
	}
	for _, automation := range automations {
		if automation.Enabled {
			if err := scheduler.RemoveByTag(fmt.Sprintf("%d", automation.Id)); err != nil {
				log.Error(fmt.Sprintf("Failed to deactivate automation '%d': could not stop scheduler: %s", automation.Id, err.Error()))
				continue
			}
			log.Debug(fmt.Sprintf("Successfully deactivated automation '%d' of user '%s'", automation.Id, automation.Owner))
		}
	}
	log.Debug("Successfully disabled automation system")
	return nil
}

// Initializes the scheduler
func Init() error {
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil {
		log.Error("Failed to initialize automation scheduler: could not retrieve server configuration: ", err.Error())
		return err
	}
	if !found {
		log.Error("Failed to initialize automation scheduler: could not retrieve server configuration: no results in query")
		return err
	}
	scheduler = gocron.NewScheduler(time.Local)
	scheduler.TagsUnique()
	if serverConfig.AutomationEnabled {
		if err := ActivateAutomationSystem(); err != nil {
			log.Error("Failed to activate automation system: could not activate persistent jobs: ", err.Error())
			return err
		}
		log.Info("Successfully activated automation scheduler system")
	} else {
		log.Info("Not activating scheduler automation system because it is disabled")
	}
	scheduler.StartAsync()
	return nil
}
