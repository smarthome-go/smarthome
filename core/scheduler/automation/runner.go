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
		if err := user.Notify(
			job.Owner,
			"Automation Failed",
			fmt.Sprintf("Automation with id: '%d' was not executed because it could not be found in the database due to an internal error, contact your administrator", id),
			user.NotificationLevelError,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}
	if !jobFound {
		log.Error(fmt.Sprintf("Automation with id: '%d' could not be executed: Id not found in database", id))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation with id: '%d' could not be executed because it could not be found in the database", id),
		)
		if err := user.Notify(
			job.Owner,
			"Automation Failed",
			fmt.Sprintf("Automation with id: '%d' was not executed because it could not be found in the database, contact your administrator", id),
			user.NotificationLevelError,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}
	// Notify and reminf the user about the disabled automation
	if !job.Enabled {
		log.Info(fmt.Sprintf("Automation '%s' was not executed because it is deactivated", job.Name))
		if err := user.Notify(
			job.Owner,
			"Automation Skipped",
			fmt.Sprintf("Automation '%s' was not executed because it is currently disabled. If you want to disable this automation completely, delete it. If the automation should be executed next time, enable it.", job.Name),
			user.NotificationLevelWarn,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}
	// If the timing mode is set to either 'sunrise' or 'sunset', a new time with according cron-expression should be generated
	if job.TimingMode != database.TimingNormal {
		if err := updateJobTime(id, job.TimingMode == database.TimingSunrise); err != nil {
			log.Error("Failed to run automation: could not update next launch time: ", err.Error())
			event.Error(
				"Automation Failed",
				fmt.Sprintf("Automation '%s' failed because its next launch time could not be adjusted: %s", job.Name, err.Error()),
			)
			if err := user.Notify(
				job.Owner,
				"Automation Failed",
				fmt.Sprintf("Automation '%s' was not executed because the next time it should run could not be determined. This is caused by the automations timing mode which is currently set to '%s'", job.Name, job.TimingMode),
				user.NotificationLevelError,
			); err != nil {
				log.Error("Failed to notify user: ", err.Error())
				return
			}
			return
		}
	}
	log.Debug(fmt.Sprintf("Automation '%d' is running", id))
	_, scriptExists, err := database.GetUserHomescriptById(job.HomescriptId, job.Owner)
	if err != nil {
		log.Error(fmt.Sprintf("Automation '%s' failed because its Homescript Id could not be retrieved from the database: %s", job.Name, err.Error()))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation '%s' could not be executed because it s Homescript Id could not be retrieved from the database: %s", job.Name, err.Error()),
		)
		if err := user.Notify(
			job.Owner,
			"Automation Failed",
			fmt.Sprintf("Automation '%s' was not executed because its referenced Homescript could not be found in the database due to an internal error, contact your administrator", job.Name),
			user.NotificationLevelError,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}
	if !scriptExists {
		log.Error(fmt.Sprintf("Automation '%s' failed because its Homescript Id: '%s' is invalid", job.Name, job.HomescriptId))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation '%s' failed because its Homescript Id: '%s' is invalid. This indicates a bad configuration.", job.Name, job.HomescriptId),
		)
		if err := user.Notify(
			job.Owner,
			"Automation Failed",
			fmt.Sprintf("Automation '%s' was not executed because its referenced Homescript could not be found in the database, contact your administrator", job.Name),
			user.NotificationLevelError,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}
	output, exitCode, err := homescript.RunById(job.Owner, job.HomescriptId)
	if err != nil {
		log.Warn(fmt.Sprintf("Automation '%s' failed during the execution of Homescript: '%s', which terminated abnormally", job.Name, job.HomescriptId))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation '%s' failed during execution of the referenced Homescript (id: '%s'). Error: %s", job.Name, job.HomescriptId, err.Error()),
		)
		if err := user.Notify(
			job.Owner,
			"Automation Failed",
			fmt.Sprintf("Automation '%s' failed during execution of the referenced Homescript (id: '%s'). Homescript-error: %s", job.Name, job.HomescriptId, err.Error()),
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
