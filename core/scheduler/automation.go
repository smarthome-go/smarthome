package scheduler

import (
	"fmt"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/core/homescript"
	"github.com/MikMuellerDev/smarthome/core/user"
)

type Automation struct {
	Id              uint
	Name            string
	Description     string
	CronExpression  string
	CronDescription string
	HomescriptId    string
	Owner           string
}

// Creates a new automation item
// Sets up the scheduler based on the provided hour, minute, and days of the week on which the automation should run
func CreateNewAutomation(
	name string,
	description string,
	hour uint8,
	minute uint8,
	days []uint8,
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
		log.Error("Could not create automation: failed to generate cron expression: unexpected input: ", err.Error())
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
		log.Error("Could not create automation: database failure: ", err.Error())
		return err

	}
	cronDescription, err := generateHumanReadableCronExpression(cronExpression)
	if err != nil {
		log.Error("Could not create automation: failed to generate human readable string: ", err.Error())
		return err
	}
	log.Debug(fmt.Sprintf("Created new automation '%s' for user '%s. It will run %s", name, owner, cronDescription))

	// Prepare a job for go-cron
	automationJob := scheduler.Cron(cronExpression)
	automationJob.Tag(fmt.Sprintf("%d", newAutomationId))
	automationJob.Do(automationRunnerFunc, newAutomationId)
	return nil
}

// Removes an automation from the database and prevents its further execution
// Does not check if the job exists, checks should be completed beforehand
func RemoveAutomation(automationId uint) error {
	if err := database.DeleteAutomationById(automationId); err != nil {
		log.Error("Failed to remove automation item: database failure: ", err.Error())
		return err
	}
	if err := scheduler.RemoveByTag(fmt.Sprintf("%d", automationId)); err != nil {
		log.Error("Failed to remove automation item: could not stop job: ", err.Error())
		return err
	}
	return nil
}

// Returns a slice of automations which a given user has set up
// Does not check the validity of the user
func GetUserAutomations(username string) ([]Automation, error) {
	automations := make([]Automation, 0)
	automationsTemp, err := database.GetUserAutomations(username)
	if err != nil {
		log.Error("Failed to list automations of user: database failure: ", err.Error())
		return nil, err
	}
	for _, automation := range automationsTemp {
		cronDescription, err := generateHumanReadableCronExpression(automation.CronExpression)
		if err != nil {
			log.Error("Failed to list automations of user: could not generate cron description: ", err.Error())
			return nil, err
		}
		automations = append(automations,
			Automation{
				Id:              automation.Id,
				Name:            automation.Name,
				Description:     automation.Description,
				CronExpression:  automation.CronExpression,
				CronDescription: cronDescription,
				HomescriptId:    automation.HomescriptId,
				Owner:           automation.Owner,
			},
		)
	}
	return automations, nil
}

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
