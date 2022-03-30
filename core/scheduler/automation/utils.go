package automation

import (
	"errors"
	"fmt"

	"github.com/MikMuellerDev/smarthome/core/database"
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
	Enabled         bool
	TimingMode      database.TimingMode
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
	enabled bool,
	timingMode database.TimingMode,
) error {
	// Generate a cron expression based on the input data
	// The `days` slice should not be longer than 7
	cronExpression, err := GenerateCronExpression(
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
			Enabled:        enabled,
			TimingMode:     timingMode,
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
	if enabled {
		user.Notify(
			owner,
			"Automation Added",
			fmt.Sprintf("Automation '%s' has been added", name),
			1,
		)
		log.Debug(fmt.Sprintf("Created new automation '%s' for user '%s. It will run %s", name, owner, cronDescription))
	} else {
		user.Notify(
			owner,
			"Inactive Automation Added",
			fmt.Sprintf("Automation '%s' has been added but is currently disabled", name),
			2,
		)
		log.Trace(fmt.Sprintf("Added automation '%d' which is currently disabled, not adding to scheduler", newAutomationId))
		return nil
	}
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		log.Error("Failed to setup new automation: could not retrieve server configuration due to database failure")
		return errors.New("failed to setup new automation: could not retrieve server configuration due to database failure")
	}
	if !serverConfig.AutomationEnabled { // If the automation scheduler is disabled, do not add the scheduler
		return nil

	}
	if timingMode != database.TimingNormal {
		// Add a dummy scheduler so that the modify function does not fail
		automationJob := scheduler.Cron(cronExpression)
		automationJob.Tag(fmt.Sprintf("%d", newAutomationId))
		automationJob.Do(func() {})
		// If the timing mode is set to either `sunrise` or `sunset`, do not activate the automation, update it instead
		return updateJobTime(newAutomationId, timingMode == database.TimingSunrise)
	}
	// Prepare the job for go-cron
	automationJob := scheduler.Cron(cronExpression)
	automationJob.Tag(fmt.Sprintf("%d", newAutomationId))
	automationJob.Do(automationRunnerFunc, newAutomationId)
	return nil
}

// Removes an automation from the database and prevents its further execution
// Does not check if the job exists, checks should be completed beforehand
func RemoveAutomation(automationId uint) error {
	previousAutomation, _, err := database.GetAutomationById(automationId)
	if err != nil {
		log.Error("Failed tp remove automation item: database failure: ", err.Error())
		return err
	}
	if err := database.DeleteAutomationById(automationId); err != nil {
		log.Error("Failed to remove automation item: database failure: ", err.Error())
		return err
	}
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		log.Error("Failed to remove automation: could not retrieve server configuration due to database failure")
		return errors.New("failed to remove automation: could not retrieve server configuration due to database failure")
	}
	if !previousAutomation.Enabled || !serverConfig.AutomationEnabled { // A disabled automation cannot be removed from the scheduler, so return here
		log.Trace(fmt.Sprintf("Removed already disabled automation '%d'", automationId))
		return nil
	}
	if err := scheduler.RemoveByTag(fmt.Sprintf("%d", automationId)); err != nil {
		log.Error("Failed to remove automation item: could not stop job: ", err.Error())
		return err
	}
	log.Trace(fmt.Sprintf("Deactivated and removed automation '%d'", automationId))
	user.Notify(
		previousAutomation.Owner,
		"Removed Automation",
		fmt.Sprintf("The Automation '%s' has been removed from the system", previousAutomation.Name),
		1,
	)
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
				Enabled:         automation.Enabled,
				TimingMode:      automation.TimingMode,
			},
		)
	}
	return automations, nil
}

// Returns the automation, if it exists and an error
// Returns an automation given its id and the owners username
func GetUserAutomationById(username string, automationId uint) (Automation, bool, error) {
	automationsTemp, err := database.GetUserAutomations(username)
	if err != nil {
		log.Error("Failed to get user automation by id: database failure: ", err.Error())
		return Automation{}, false, err
	}
	for _, automation := range automationsTemp {
		if automation.Id != automationId {
			continue // Skip any automations which don't match
		}
		cronDescription, err := generateHumanReadableCronExpression(automation.CronExpression)
		if err != nil {
			log.Error("Failed to get user automation by id: could not generate cron description: ", err.Error())
			return Automation{}, false, err
		}
		return Automation{
			Id:              automation.Id,
			Name:            automation.Name,
			Description:     automation.Description,
			CronExpression:  automation.CronExpression,
			CronDescription: cronDescription,
			HomescriptId:    automation.HomescriptId,
			Owner:           automation.Owner,
			Enabled:         automation.Enabled,
			TimingMode:      automation.TimingMode,
		}, true, nil
	}
	return Automation{}, false, nil
}

// Changes the metadata of a given automation and then restarts it so it uses the updated values
func ModifyAutomationById(automationId uint, newAutomation database.AutomationWithoutIdAndUsername) error {
	if !IsValidCronExpression(newAutomation.CronExpression) {
		log.Error("Failed to modify automation: invalid cron expression provided")
		return errors.New("failed to modify automation: invalid cron expression provided")
	}
	automationBefore, _, err := database.GetAutomationById(automationId)
	if err != nil {
		log.Error("Failed to modify automation by id: could not get previous state due to database failure: ", err.Error())
		return err
	}
	if err := database.ModifyAutomation(automationId, newAutomation); err != nil {
		log.Error("Failed to modify automation by id: database failure during modification: ", err.Error())
		return err
	}
	if automationBefore.Enabled { // If the automation was enabled before it was modified, remove it
		// After the metadata has been changed, restart the scheduler
		if err := scheduler.RemoveByTag(fmt.Sprintf("%d", automationId)); err != nil {
			log.Error("Failed to remove automation item: could not stop job: ", err.Error())
			return err
		}
	}
	if newAutomation.Enabled {
		// Restart the scheduler after the old one was disabled
		// Only add the scheduler if it is enabled in the new version
		automationJob := scheduler.Cron(newAutomation.CronExpression)
		automationJob.Tag(fmt.Sprintf("%d", automationId))
		automationJob.Do(automationRunnerFunc, automationId)
		log.Debug(fmt.Sprintf("Automation %d has been modified and restarted", automationId))

		if !automationBefore.Enabled {
			log.Trace(fmt.Sprintf("Automation with id %d has been activated", automationId))
			user.Notify(
				automationBefore.Owner,
				"Automation Activated",
				fmt.Sprintf("Automation '%s' has been activated", newAutomation.Name),
				1,
			)
		}
	} else {
		if automationBefore.Enabled {
			log.Trace(fmt.Sprintf("Automation with id %d has been disabled", automationId))
			user.Notify(
				automationBefore.Owner,
				"Automation Temporarely Disabled",
				fmt.Sprintf("Automation '%s' has been disabled", automationBefore.Name),
				2,
			)
		}
		log.Debug(fmt.Sprintf("Automation %d has been modified but not started due to being disabled", automationId))
	}
	return nil
}
