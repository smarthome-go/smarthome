package homescript

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-co-op/gocron"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/homescript/automation"
	"github.com/smarthome-go/smarthome/core/user"
)

// The automationScheduler which will run all predefined automation jobs
var automationScheduler *gocron.Scheduler

// Activates all jobs saved in the database, will be used when the server was restarted
// If a non-critical error occurs, for example the failure to setup a job, it will be returned
// This function will not cancel directly if an error occurs in order to preserve the automation system's uptime
func ActivateAutomationSystem() error {
	automations, err := database.GetAutomations()
	if err != nil {
		log.Error("Failed to activate automation system: database failure whilst starting saved automations: ", err.Error())
		return err // This is a critical error which can not be recovered from
	}
	var activatedItems uint = 0
	for _, automationItem := range automations {
		if !automation.IsValidCronExpression(automationItem.Data.CronExpression) {
			// Log the error
			log.Error(fmt.Sprintf("Could not activate automation '%d': invalid cron expression", automationItem.Id))
			event.Error("Automation Activation Failure", fmt.Sprintf("The automation %s could not be activated due to an internal error. Please remove it from the system.", automationItem.Data.Name))

			if err := user.Notify(
				automationItem.Owner,
				"Automation Activation Failure",
				fmt.Sprintf("The automation %s could not be activated due to an internal error. Please remove it from the system.", automationItem.Data.Name), 3); err != nil {
				log.Error("Failed to notify user about failing automation: ", err.Error())
			}
			continue // non-critical error, will only affect this automation
		}
		if !automationItem.Data.Enabled {
			event.Trace("Disabled Automation Skipped", fmt.Sprintf("Automation `%d` was not started because it is disabled", automationItem.Id))
			log.Debug(fmt.Sprintf("Skipping activation of automation '%d': automation is disabled", automationItem.Id))
			continue // Skip disabled automations
		}
		automationJob := automationScheduler.Cron(automationItem.Data.CronExpression)
		automationJob.Tag(fmt.Sprintf("%d", automationItem.Id))
		_, err := automationJob.Do(automationRunnerFunc, automationItem.Id)
		if err != nil {
			event.Error("Automation Activation Failure", fmt.Sprintf("Could not activate automation '%d': failed to register cron job: %s", automationItem.Id, err.Error()))
			log.Error(fmt.Sprintf("Could not activate automation '%d': failed to register cron job: %s", automationItem.Id, err.Error()))
			return err
		}
		activatedItems += 1
		event.Debug("Automation Activated", fmt.Sprintf("Successfully activated automation '%d' of user '%s'", automationItem.Id, automationItem.Owner))
		log.Debug(fmt.Sprintf("Successfully activated automation '%d' of user '%s'", automationItem.Id, automationItem.Owner))
	}
	if activatedItems > 0 {
		event.Info("Automation System Activated", fmt.Sprintf("Successfully activated saved automations: started %d total automation jobs", activatedItems))
		log.Info(fmt.Sprintf("Successfully activated saved automations: started %d total automation jobs", activatedItems))
	}
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
		if automation.Data.Enabled {
			if err := automationScheduler.RemoveByTag(fmt.Sprintf("%d", automation.Id)); err != nil {
				log.Error(fmt.Sprintf("Failed to deactivate automation '%d': could not stop scheduler: '%s'", automation.Id, err.Error()))
				log.Error("Automation System Deactivation Failure: ", fmt.Sprintf("Failed to deactivate automation '%d': could not stop scheduler: '%s'", automation.Id, err.Error()))
				continue
			}
			event.Debug("Deactivated Automation", fmt.Sprintf("Successfully deactivated automation '%d' of user '%s'", automation.Id, automation.Owner))
			log.Debug(fmt.Sprintf("Successfully deactivated automation '%d' of user '%s'", automation.Id, automation.Owner))
		}
	}
	log.Info("Successfully disabled automation system: all jobs were stopped")
	event.Info("Disabled Automation System", "Successfully disabled automation system: all jobs were stopped")
	return nil
}

// Given a jobId and whether sunrise or sunset should is activated, the next execution time is modified
// Used when an automation with non-normal Timing-Mode is executed in order to update its next start time
func UpdateJobTime(id uint, useSunRise bool) error {
	// Obtain the server's configuration in order to determine the latitude and longitude
	// config, found, err := database.GetServerConfiguration()
	// if err != nil || !found {
	// 	log.Error("Failed to update job launch time: could not obtain the server's configuration")
	// 	return errors.New("could not update launch time: failed to obtain server config")
	// }
	// Retrieve the current job in order to get its current cron-expression (for the days)
	job, found, err := database.GetAutomationById(id)
	if err != nil || !found {
		return errors.New("could not update launch time: invalid id supplied")
	}
	// // Calculate both the sunrise and sunset time
	// sunRise, sunSet := CalculateSunRiseSet(config.Latitude, config.Longitude)
	// // Select the time which is desired
	// var finalTime SunTime
	// if useSunRise {
	// 	finalTime = sunRise
	// } else {
	// 	finalTime = sunSet
	// }
	// // Extract the days from the cron-expression
	// days, err := GetDaysFromCronExpression(job.CronExpression)
	// if err != nil {
	// 	log.Error(fmt.Sprintf("Failed to extract days from cron-expression '%s': Error: %s", job.CronExpression, err))
	// 	return err
	// }
	// cronExpression, err := GenerateCronExpression(uint8(finalTime.Hour), uint8(finalTime.Minute), days)
	// if err != nil {
	// 	return err
	// }

	// TODO: why is this off?

	// Only triggers the generic modification due to a lot of work being done in the modification function
	if err := ModifyAutomationById(id, database.AutomationData{
		Name:           job.Data.Name,
		Description:    job.Data.Description,
		CronExpression: job.Data.CronExpression,
		HomescriptId:   job.Data.HomescriptId,
		Enabled:        job.Data.Enabled,
		TimingMode:     job.Data.TimingMode,
	}); err != nil {
		log.Error(fmt.Sprintf("Failed to update next execution time of automation '%d': could not modify automation: %s", id, err.Error()))
		return err
	}
	log.Trace(fmt.Sprintf("Successfully updated the next execution time of automation '%d'", id))
	return nil
}

// Initializes the scheduler
func InitAutomations() error {
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil {
		log.Error("Failed to initialize automation scheduler: could not retrieve server configuration: ", err.Error())
		return err
	}
	if !found {
		log.Error("Failed to initialize automation scheduler: could not retrieve server configuration: no results for server config")
		return err
	}
	automationScheduler = gocron.NewScheduler(time.Local)
	automationScheduler.TagsUnique()
	if serverConfig.AutomationEnabled {
		if err := ActivateAutomationSystem(); err != nil {
			log.Error("Failed to activate automation system: could not activate persistent jobs: ", err.Error())
			return err
		}
		log.Info("Successfully activated automation system")
	} else {
		log.Info("Skipping activation of automation system due to it being disabled")
	}
	automationScheduler.StartAsync()
	return nil
}
