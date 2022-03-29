package automation

import (
	"fmt"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/core/homescript"
	"github.com/MikMuellerDev/smarthome/core/user"
)

// Is called when the scheduler executes the given automation
// The automationRunnerFunc automatically tries to fetch the needed configuration from the provided id
// Error handling in this function works by logging to the internal event system and notifying the user about their automation's failure
func automationRunnerFunc(id uint) {
	job, jobFound, err := database.GetAutomationById(id)
	if err != nil {
		log.Error(fmt.Sprintf("Automation with the Id '%d' could not be executed: database failure: %s", id, err.Error()))
		event.Error(
			"Automation Failure",
			fmt.Sprintf("Automation with the Id: '%d' could not be executed due to database failure: %s", id, err.Error()),
		)
		return
	}
	if !jobFound {
		log.Error(fmt.Sprintf("Automation with the Id: '%d' could not be executed: Id not found in database", id))
		event.Error(
			"Automation Failure",
			fmt.Sprintf("Automation with the Id: '%d' could not be executed because it could not be found in the database", id),
		)
		return
	}
	if !job.Enabled {
		log.Info(fmt.Sprintf("Automations %d canceled because it is deactivated", id))
		user.Notify(
			job.Owner,
			"Automation Suspended",
			fmt.Sprintf("Automation '%s' was not executed because it is currently disabled. If you want to disable this automation completely, delete it. If the automation should run, enable it.", job.Name),
			user.NotificationLevelWarn,
		)
		return
	}
	log.Debug(fmt.Sprintf("Automation '%d' is running.\n", id))
	_, scriptExists, err := database.GetUserHomescriptById(job.HomescriptId, job.Owner)
	if err != nil {
		log.Error(fmt.Sprintf("Automation with Id: '%d' failed because its Homescript Id could not be retrieved from the database: %s", id, err.Error()))
		event.Error(
			"Automation Failure",
			fmt.Sprintf("Automation with the Id: '%d' could not be executed because it s Homescript Id could not be retrieved from the database: %s", id, err.Error()),
		)
		return
	}
	if !scriptExists {
		log.Error(fmt.Sprintf("Automation with Id: '%d' failed because its Homescript Id: '%s' is invalid", id, job.HomescriptId))
		event.Error(
			"Automation Failure",
			fmt.Sprintf("Automation with the Id: '%d' failed because its Homescript Id: '%s' is invalid. This indicates a bad configuration.", id, job.HomescriptId),
		)
		user.Notify(
			job.Owner,
			"Automation Failure",
			fmt.Sprintf("Automation '%s' failed because its Homescript Id: '%s' is invalid. Contact your administrator.", job.Name, job.HomescriptId),
			user.NotificationLevelError,
		)
		return
	}
	output, exitCode, err := homescript.RunById(job.Owner, job.HomescriptId)
	if err != nil {
		log.Warn(fmt.Sprintf("Automation with Id: '%d' failed because the Homescript: '%s' could not be executed", id, job.HomescriptId))
		event.Error(
			"Automation Failure",
			fmt.Sprintf("Automation with Id: '%d' failed. Error: %s", id, err.Error()),
		)
		user.Notify(
			job.Owner,
			"Automation Failure",
			fmt.Sprintf("Automation '%s' failed during execution. Error: %s", job.Name, err.Error()),
			user.NotificationLevelError,
		)
		return
	}
	event.Debug(
		"Automation Finished Successfully",
		fmt.Sprintf("Automation '%d' of user '%s' has executed successfully. Exit code: %d, Output: '%s'", id, job.Owner, exitCode, output),
	)
}
