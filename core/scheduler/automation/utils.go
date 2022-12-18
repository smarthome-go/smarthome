package automation

import (
	"errors"
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/user"
)

type Automation struct {
	Id              uint                `json:"id"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	CronExpression  string              `json:"cronExpression"`
	CronDescription string              `json:"cronDescription"`
	HomescriptId    string              `json:"homescriptId"`
	Owner           string              `json:"owner"`
	Enabled         bool                `json:"enabled"`
	DisableOnce     bool                `json:"disableOnce"`
	TimingMode      database.TimingMode `json:"timingMode"`
}

// Creates a new automation which an according database entry
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
) (uint, error) {
	// Generate a cron expression based on the input data
	// The `days` slice should not contain more than 7 elements
	cronExpression, err := GenerateCronExpression(
		hour,
		minute,
		days,
	)
	if err != nil {
		log.Error("Could not create automation: failed to generate cron expression: unexpected input: ", err.Error())
		return 0, err
	}
	// Insert the automation into the database
	newAutomationId, err := database.CreateNewAutomation(
		database.Automation{
			Owner: owner,
			Data: database.AutomationData{
				Name:           name,
				Description:    description,
				CronExpression: cronExpression,
				HomescriptId:   homescriptId,
				Enabled:        enabled,
				DisableOnce:    false,
				TimingMode:     timingMode,
			},
		},
	)
	if err != nil {
		log.Error("Could not create automation: database failure: ", err.Error())
		return 0, err
	}
	if enabled {
		log.Debug(fmt.Sprintf("Created new automation '%s' for user '%s'. Cron-Expression: `%s`", name, owner, cronExpression))
	} else {
		if err := user.Notify(
			owner,
			"Inactive Automation Added",
			fmt.Sprintf("Automation '%s' has been added to the system, however it is currently disabled and thus will not be executed %s", name, description),
			2,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return 0, err
		}
		log.Trace(fmt.Sprintf("Added automation '%d' which is currently disabled, skipping registration to scheduler", newAutomationId))
		return newAutomationId, nil
	}
	// Retrieve the server config in order to determine if the automation system is enabled
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		log.Error("Failed to setup new automation: could not retrieve server configuration due to database failure")
		return 0, errors.New("failed to setup new automation: could not retrieve server configuration due to database failure")
	}
	if !serverConfig.AutomationEnabled { // If the automation scheduler is disabled, do not add the scheduler
		return newAutomationId, nil
	}
	if timingMode != database.TimingNormal {
		// Add a dummy scheduler which does nothing in order to prevent the modify function from failing
		automationJob := scheduler.Cron(cronExpression)
		automationJob.Tag(fmt.Sprintf("%d", newAutomationId))
		if _, err := automationJob.Do(func() {}); err != nil {
			log.Error("Failed to register dummy cron job: ", err.Error())
			return 0, err
		}
		// If the timing mode is set to either `sunrise` or `sunset`, do not activate the automation, update it instead
		return newAutomationId, updateJobTime(newAutomationId, timingMode == database.TimingSunrise)
	}
	// Otherwise, register a cron job for the automation
	automationJob := scheduler.Cron(cronExpression)
	automationJob.Tag(fmt.Sprintf("%d", newAutomationId))
	if _, err = automationJob.Do(automationRunnerFunc, newAutomationId); err != nil {
		log.Error("Failed to register cron job: ", err.Error())
		return 0, err
	}
	event.Debug("Automation Created", fmt.Sprintf("%s created a new automation (Name: %s, At %d:%d)", owner, name, hour, minute))
	return newAutomationId, nil
}

// Removes an automation from the database and prevents its further execution
func RemoveAutomation(automationId uint) error {
	previousAutomation, exists, err := database.GetAutomationById(automationId)
	if err != nil {
		log.Error("Failed to remove automation: database failure: ", err.Error())
		return err
	}
	if !exists {
		log.Error(fmt.Sprintf("Failed to remove automation: no such id ('%d') is currently registered", automationId))
		return fmt.Errorf("failed to remove automation: id '%d' is not a currently active automation", automationId)
	}
	if err := database.DeleteAutomationById(automationId); err != nil {
		log.Error("Failed to remove automation: database failure: ", err.Error())
		return err
	}
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		log.Error("Failed to remove automation: could not retrieve server configuration due to database failure")
		return errors.New("failed to remove automation: could not retrieve server configuration due to database failure")
	}
	if !previousAutomation.Data.Enabled || !serverConfig.AutomationEnabled { // A disabled automation cannot be removed from the scheduler, so return here
		log.Trace(fmt.Sprintf("Removed an already disabled automation id: '%d'", automationId))
		return nil
	}
	if err := scheduler.RemoveByTag(fmt.Sprintf("%d", automationId)); err != nil {
		log.Error("Failed to remove automation item: could not stop cron job: ", err.Error())
		return err
	}
	log.Trace(fmt.Sprintf("Deactivated and removed automation. id: '%d'", automationId))
	if err := user.Notify(
		previousAutomation.Owner,
		"Removed Automation",
		fmt.Sprintf("The Automation '%s' has been successfully removed from the system and will not execute again", previousAutomation.Data.Name),
		1,
	); err != nil {
		log.Error("Failed to notify user: ", err.Error())
		return err
	}
	event.Debug("Automation Removed", fmt.Sprintf("Automation %d was removed from the system", automationId))
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
		cronDescription, err := generateHumanReadableCronExpression(automation.Data.CronExpression)
		if err != nil {
			log.Error("Failed to list automations of user: could not generate cron description: ", err.Error())
			return nil, err
		}
		automations = append(automations,
			Automation{
				Id:              automation.Id,
				Name:            automation.Data.Name,
				Description:     automation.Data.Description,
				CronExpression:  automation.Data.CronExpression,
				CronDescription: cronDescription,
				HomescriptId:    automation.Data.HomescriptId,
				Owner:           automation.Owner,
				Enabled:         automation.Data.Enabled,
				DisableOnce:     automation.Data.DisableOnce,
				TimingMode:      automation.Data.TimingMode,
			},
		)
	}
	return automations, nil
}

// Given an username and id, it returns a matching automation, whether it exists and an error
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
		cronDescription, err := generateHumanReadableCronExpression(automation.Data.CronExpression)
		if err != nil {
			log.Error("Failed to get user automation by id: could not generate cron description: ", err.Error())
			return Automation{}, false, err
		}
		return Automation{
			Id:              automation.Id,
			Name:            automation.Data.Name,
			Description:     automation.Data.Description,
			CronExpression:  automation.Data.CronExpression,
			CronDescription: cronDescription,
			HomescriptId:    automation.Data.HomescriptId,
			Owner:           automation.Owner,
			Enabled:         automation.Data.Enabled,
			DisableOnce:     automation.Data.DisableOnce,
			TimingMode:      automation.Data.TimingMode,
		}, true, nil
	}
	return Automation{}, false, nil
}

// Changes the metadata of a given automation, then restarts it so that it uses the updated values such as execution time
// Is also used after an automation with non-normal timing has been added
func ModifyAutomationById(automationId uint, newAutomation database.AutomationData) error {
	if !IsValidCronExpression(newAutomation.CronExpression) {
		log.Error("Failed to modify automation: invalid cron expression provided")
		return errors.New("failed to modify automation: invalid cron expression provided")
	}
	automationBefore, exists, err := database.GetAutomationById(automationId)
	if err != nil {
		log.Error("Failed to modify automation by id: could not get previous state due to database failure: ", err.Error())
		return err
	}
	if !exists {
		log.Error("Failed to modify automation by id: could not get previous automation: not found")
		return fmt.Errorf("failed to modify automation by id: could not get previous automation: not found")
	}
	// Update sunset / sunrise times if changed
	if newAutomation.TimingMode != database.TimingNormal {
		// Obtain the server's configuration in order to determine the latitude and longitude
		config, found, err := database.GetServerConfiguration()
		if err != nil || !found {
			log.Error("Failed to update job launch time: could not obtain the server's configuration")
			return errors.New("could not update launch time: failed to obtain server config")
		}
		// Calculate both the sunrise and sunset time
		sunRise, sunSet := CalculateSunRiseSet(config.Latitude, config.Longitude)
		// Select the time which is desired
		var finalTime SunTime
		if newAutomation.TimingMode == database.TimingSunrise {
			finalTime = sunRise
		} else {
			finalTime = sunSet
		}
		// Extract the days from the cron-expression
		days, err := getDaysFromCronExpression(newAutomation.CronExpression)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to extract days from cron-expression '%s': Error: %s", newAutomation.CronExpression, err))
			return err
		}
		cronExpression, err := GenerateCronExpression(uint8(finalTime.Hour), uint8(finalTime.Minute), days)
		if err != nil {
			return err
		}
		newAutomation.CronExpression = cronExpression
	}
	if err := database.ModifyAutomation(automationId, newAutomation); err != nil {
		log.Error("Failed to modify automation by id: database failure during modification: ", err.Error())
		return err
	}
	// If the automation was enabled before it was modified, remove it from the cron jobs
	// Only attempt to remove it if the server's automation system is enabled
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil {
		return err
	}
	if !found {
		return errors.New("could not retrieve server configuration")
	}
	if automationBefore.Data.Enabled && serverConfig.AutomationEnabled {
		// After the metadata has been changed, restart the scheduler
		if err := scheduler.RemoveByTag(fmt.Sprintf("%d", automationId)); err != nil {
			log.Error("Failed to remove automation item: could not stop cron job: ", err.Error())
			return err
		}
	}
	// Restart the scheduler after the old one was disabled
	// Only add the scheduler if it is enabled in the new version and the server's automation system is active
	if newAutomation.Enabled && serverConfig.AutomationEnabled {
		automationJob := scheduler.Cron(newAutomation.CronExpression)
		automationJob.Tag(fmt.Sprintf("%d", automationId))
		if _, err := automationJob.Do(automationRunnerFunc, automationId); err != nil {
			log.Error("Failed to modify automation, registering cron job failed: ", err.Error())
			return err
		}
		log.Debug(fmt.Sprintf("Automation %d has been modified and restarted", automationId))
		if !automationBefore.Data.Enabled {
			log.Trace(fmt.Sprintf("Automation with id %d has been activated", automationId))
			if err := user.Notify(
				automationBefore.Owner,
				"Automation Activated",
				fmt.Sprintf("Automation '%s' has been activated", newAutomation.Name),
				1,
			); err != nil {
				log.Error("Failed to notify user: ", err.Error())
				return err
			}
		}
	} else {
		if automationBefore.Data.Enabled {
			log.Trace(fmt.Sprintf("Automation with id %d has been disabled", automationId))
			if err := user.Notify(
				automationBefore.Owner,
				"Automation Temporarily Disabled",
				fmt.Sprintf("Automation '%s' has been disabled", automationBefore.Data.Name),
				2,
			); err != nil {
				log.Error("Failed to notify user: ", err.Error())
				return err
			}
		}
		log.Debug(fmt.Sprintf("Automation %d has been modified and disabled", automationId))
	}
	event.Debug("Automation Modified", fmt.Sprintf("Automation %d was modified", automationBefore.Id))
	return nil
}
