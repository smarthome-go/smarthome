package automation

import (
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/core/user"
)

// Is called when the scheduler executes the given automation
// The automationRunnerFunc automatically tries to fetch the required configuration from the provided id
// Error handling is accomplished by logging to the internal event system and notifying the user about their automations failure
func automationRunnerFunc(id uint) {
	job, jobFound, err := database.GetAutomationById(id)
	if err != nil {
		log.Error(fmt.Sprintf("Automation with id: '%d' could not be executed: database failure: %s", id, err.Error()))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation with id: '%d' could not be executed due to database failure: %s", id, err.Error()),
		)
		return
	}
	if !jobFound {
		log.Error(fmt.Sprintf("Automation with id: '%d' could not be executed: Id not found in database", id))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation with id: '%d' could not be executed because it could not be found in the database", id),
		)
		// Abort this automation in order to prevent future errors
		if err := scheduler.RemoveByTag(fmt.Sprintf("%d", id)); err != nil {
			log.Error("Failed to remove dangling automation: could not stop cron job: ", err.Error())
			return
		}
		log.Info(fmt.Sprintf("Successfully aborted dangling automation: %d", id))
		return
	}
	// Check if the user has blocked their automations & schedules
	userData, found, err := database.GetUserByUsername(job.Owner)
	if err != nil {
		log.Error("Automation failed because owner user could not be determined")
		return
	}
	if !found {
		log.Warn("Automation failed because owner user does not exist anymore, deleting automation...")
		if err := database.DeleteAutomationById(id); err != nil {
			log.Error("Cleaning up dangling automation failed: could not remove automation from database: ", err.Error())
		}
		return
	}
	if !userData.SchedulerEnabled {
		log.Debug(fmt.Sprintf("Automation '%s' was not executed because its owner has disabled their schedules & automations", job.Data.Name))
		event.Debug(
			"Automation Skipped",
			fmt.Sprintf("Automation '%s' has been skipped", job.Data.Name),
		)
		return
	}
	// Check if the automation's next run (this run) is disalbed
	if job.Data.DisableOnce {
		// Re-enable the automation again
		if err := ModifyAutomationById(job.Id, database.AutomationData{
			Name:           job.Data.Name,
			Description:    job.Data.Description,
			CronExpression: job.Data.CronExpression,
			HomescriptId:   job.Data.HomescriptId,
			Enabled:        job.Data.Enabled,
			DisableOnce:    false,
			TimingMode:     job.Data.TimingMode,
		}); err != nil {
			event.Error("Could not re-enable automation", fmt.Sprintf("Could not re-enable automation `%s`: %s", job.Data.Name, err.Error()))
			return
		}
		// Notify the user
		log.Info(fmt.Sprintf("Automation '%s' was skipped once", job.Data.Name))
		if err := user.Notify(
			job.Owner,
			"Automation Skipped Once",
			fmt.Sprintf("Automation '%s' was skipped once. It will run regularely the next time.", job.Data.Name),
			user.NotificationLevelInfo,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}
	// If the timing mode is set to either 'sunrise' or 'sunset', a new time with according cron-expression should be generated
	if job.Data.TimingMode != database.TimingNormal {
		if err := updateJobTime(id, job.Data.TimingMode == database.TimingSunrise); err != nil {
			log.Error("Failed to run automation: could not update next launch time: ", err.Error())
			event.Error(
				"Automation Failed",
				fmt.Sprintf("Automation '%s' failed because its next launch time could not be adjusted: %s", job.Data.Name, err.Error()),
			)
			if err := user.Notify(
				job.Owner,
				"Automation Failed",
				fmt.Sprintf("Automation '%s' was not executed because the next time it should run could not be determined. This is caused by the automations timing mode which is currently set to '%s'", job.Data.Name, job.Data.TimingMode),
				user.NotificationLevelError,
			); err != nil {
				log.Error("Failed to notify user: ", err.Error())
				return
			}
			return
		}
	}
	log.Debug(fmt.Sprintf("Automation '%d' is running", id))
	_, scriptExists, err := database.GetUserHomescriptById(job.Data.HomescriptId, job.Owner)
	if err != nil {
		log.Error(fmt.Sprintf("Automation '%s' failed because its Homescript Id could not be retrieved from the database: %s", job.Data.Name, err.Error()))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation '%s' could not be executed because it s Homescript Id could not be retrieved from the database: %s", job.Data.Name, err.Error()),
		)
		if err := user.Notify(
			job.Owner,
			"Automation Failed",
			fmt.Sprintf("Automation '%s' was not executed because its referenced Homescript could not be found in the database due to an internal error, contact your administrator", job.Data.Name),
			user.NotificationLevelError,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}
	if !scriptExists {
		log.Error(fmt.Sprintf("Automation '%s' failed because its Homescript Id: '%s' is invalid", job.Data.Name, job.Data.HomescriptId))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation '%s' failed because its Homescript Id: '%s' is invalid. This indicates a bad configuration.", job.Data.Name, job.Data.HomescriptId),
		)
		if err := user.Notify(
			job.Owner,
			"Automation Failed",
			fmt.Sprintf("Automation '%s' was not executed because its referenced Homescript could not be found in the database, contact your administrator", job.Data.Name),
			user.NotificationLevelError,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}
	output, exitCode, _, err := homescript.HmsManager.RunById(
		job.Data.HomescriptId,
		job.Owner,
		make([]string, 0),
		false,
		make(map[string]string, 0),
		homescript.InitiatorAutomation,
		make(chan int),
	)
	if err != nil {
		log.Warn(fmt.Sprintf("Automation '%s' failed during the execution of Homescript: '%s', which terminated abnormally", job.Data.Name, job.Data.HomescriptId))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation '%s' failed during execution of Homescript '%s'. Error: %s", job.Data.Name, job.Data.HomescriptId, err.Error()),
		)
		if err := user.Notify(
			job.Owner,
			"Automation Failed",
			fmt.Sprintf("Automation '%s' failed during execution of Homescript '%s'. Error: %s", job.Data.Name, job.Data.HomescriptId, err.Error()),
			user.NotificationLevelError,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}
	event.Debug(
		"Automation Executed Successfully",
		fmt.Sprintf("Automation '%d' of user '%s' has executed successfully. HMS-Exit code: %d, HMS-Output: '%s'", id, job.Owner, exitCode, output),
	)
}
