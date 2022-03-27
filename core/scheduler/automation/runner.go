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
func automationRunnerFunc(automationId uint) {
	job, jobFound, err := database.GetAutomationById(automationId)
	if err != nil {
		log.Error(fmt.Sprintf("Automation with the Id '%d' could not be executed: database failure: %s", automationId, err.Error()))
		event.Error(
			"Automation Failure",
			fmt.Sprintf("Automation with the Id: '%d' could not be executed due to database failure: %s", automationId, err.Error()),
		)
		return
	}
	if !jobFound {
		log.Error(fmt.Sprintf("Automation with the Id: '%d' could not be executed: Id not found in database", automationId))
		event.Error(
			"Automation Failure",
			fmt.Sprintf("Automation with the Id: '%d' could not be executed because it could not be found in the database", automationId),
		)
		return
	}
	fmt.Printf("automation '%d' is running.\n", automationId)

	_, scriptExists, err := database.GetUserHomescriptById(job.HomescriptId, job.Owner)
	if err != nil {
		log.Error(fmt.Sprintf("Automation with Id: '%d' failed because its Homescript Id could not be retrieved from the database: %s", automationId, err.Error()))
		event.Error(
			"Automation Failure",
			fmt.Sprintf("Automation with the Id: '%d' could not be executed because it s Homescript Id could not be retrieved from the database: %s", automationId, err.Error()),
		)
		return
	}
	if !scriptExists {
		log.Error(fmt.Sprintf("Automation with Id: '%d' failed because its Homescript Id: '%s' is invalid", automationId, job.HomescriptId))
		event.Error(
			"Automation Failure",
			fmt.Sprintf("Automation with the Id: '%d' failed because its Homescript Id: '%s' is invalid. This indicates a bad configuration.", automationId, job.HomescriptId),
		)
		user.Notify(
			job.Owner,
			"Automation Failure",
			fmt.Sprintf("Your automation with the Id: '%d' failed because its Homescript Id: '%s' is invalid. Contact your administrator.", automationId, job.HomescriptId),
			user.NotificationLevelError,
		)
		return
	}
	output, exitCode, err := homescript.RunById(job.Owner, job.HomescriptId)
	if err != nil {
		log.Warn(fmt.Sprintf("Automation with Id: '%d' failed because the Homescript: '%s' could not be executed", automationId, job.HomescriptId))
		event.Error(
			"Automation Failure",
			fmt.Sprintf("Automation with Id: '%d' failed. Error: %s", automationId, err.Error()),
		)
		user.Notify(
			job.Owner,
			"Automation Failure",
			fmt.Sprintf("Automation with Id: '%d' failed. Error: %s", automationId, err.Error()),
			user.NotificationLevelError,
		)
		return
	}
	event.Debug(
		"Automation Finished Successfully",
		fmt.Sprintf("Automation '%d' of user '%s' has executed successfully. Exit code: %d, Output: '%s'", automationId, job.Owner, exitCode, output),
	)
}
