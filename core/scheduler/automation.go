package scheduler

import (
	"fmt"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/core/homescript"
	"github.com/MikMuellerDev/smarthome/core/user"
)

// Creates a new automation item
func CreateNewAutomation(
	name string,
	description string,
	hour uint8,
	minute uint8,
	days []Day,
	homescriptId string,
	owner string,
) error {
	// Generate a cron expression based on the input data
	// The `days` slice should not be longer than 7
	cronExpression, err := generateCronExpression(
		hour,
		minute,
		days,
	)
	if err != nil {
		log.Error("Failed to generate cron expression: unexpected input: ", err.Error())
		return err
	}
	// Insert the automation into the database
	newAutomationId, err := database.CreateNewAutomation(
		database.Automation{
			Name:           name,
			Description:    description,
			CronExpression: cronExpression,
			HomescriptId:   homescriptId,
			Owner:          owner,
		},
	)
	if err != nil {
		log.Error("Could not create a new automation: database failure: ", err.Error())
		return err
	}
	log.Debug(fmt.Sprintf("Created new automation '%s' for user '%s", name, owner))

	// Prepare a job for go-cron
	automationJob := scheduler.Cron(cronExpression)
	automationJob.Tag(fmt.Sprintf("%d", newAutomationId))

	automationJob.Do(automationRunnerFunc, newAutomationId)
	return nil
}

// Is called when the scheduler executes the given automation
// The automationRunnerFunc automatically tries to fetch the needed configuration from the provided id
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
