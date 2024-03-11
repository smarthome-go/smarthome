package automation

import (
	"errors"
	"fmt"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

type Automation struct {
	Id                     uint                       `json:"id"`
	Name                   string                     `json:"name"`
	Description            string                     `json:"description"`
	CronDescription        *string                    `json:"cronDescription"`
	HomescriptId           string                     `json:"homescriptId"`
	Owner                  string                     `json:"owner"`
	Enabled                bool                       `json:"enabled"`
	DisableOnce            bool                       `json:"disableOnce"`
	Trigger                database.AutomationTrigger `json:"trigger"`
	TriggerCronExpression  *string                    `json:"triggerCronExpression"`
	TriggerIntervalSeconds *uint                      `json:"triggerInterval"`
}

// Creates a new automation which an according database entry
// Sets up the scheduler based on the provided hour, minute, and days of the week on which the automation should run
func (m AutomationManager) CreateNewAutomation(
	name string,
	description string,
	homescriptId string,
	owner string,
	enabled bool,
	hour *uint,
	minute *uint,
	days *[]uint8,
	trigger database.AutomationTrigger,
	triggerIntervalSeconds *uint,
) (uint, error) {
	// Generate a cron expression based on the input data if using the cron trigger
	var TriggerCronExpression *string = nil
	if trigger == database.TriggerCron {
		// The `days` slice should not contain more than 7 elements
		cronExpression, err := GenerateCronExpression(
			uint8(*hour),
			uint8(*minute),
			*days,
		)
		if err != nil {
			log.Error("Could not create automation: failed to generate cron expression: unexpected input: ", err.Error())
			return 0, err
		}
		TriggerCronExpression = &cronExpression
	}

	// Insert the automation into the database
	automationData := database.Automation{
		Owner: owner,
		Data: database.AutomationData{
			Name:                   name,
			Description:            description,
			HomescriptId:           homescriptId,
			Enabled:                enabled,
			DisableOnce:            false,
			Trigger:                trigger,
			TriggerCronExpression:  TriggerCronExpression,
			TriggerIntervalSeconds: triggerIntervalSeconds,
		},
	}
	newAutomationId, err := database.CreateNewAutomation(automationData)
	if err != nil {
		log.Error("Could not create automation: database failure: ", err.Error())
		return 0, err
	}

	// Retrieve the server config in order to determine if the automation system is enabled
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		log.Error("Failed to setup new automation: could not retrieve server configuration due to database failure")
		return 0, errors.New("failed to setup new automation: could not retrieve server configuration due to database failure")
	}

	if err := m.RegisterAutomation(newAutomationId, automationData.Data, serverConfig); err != nil {
		return 0, fmt.Errorf("Could not create new automation: registering job failed: %s", err.Error())
	}

	log.Debug(fmt.Sprintf("Created new automation '%s' for user '%s' with trigger '%v'.", name, owner, trigger))
	event.Debug("Automation Created", fmt.Sprintf("%s created a new automation (name: `%s`, trigger `%s`)", owner, name, trigger))
	return newAutomationId, nil
}

// Removes an automation from the database and prevents its further execution
func (m AutomationManager) RemoveAutomation(automationId uint) error {
	// Get current automation
	thisAutomation, exists, err := database.GetAutomationById(automationId)
	if err != nil {
		log.Error("Failed to remove automation: database failure: ", err.Error())
		return err
	}
	if !exists {
		log.Error(fmt.Sprintf("Failed to remove automation: no such id ('%d') is currently registered", automationId))
		return fmt.Errorf("failed to remove automation: id '%d' is not a currently active automation", automationId)
	}

	// Unregister automation
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		log.Error("Failed to remove automation: could not retrieve server configuration due to database failure")
		return errors.New("failed to remove automation: could not retrieve server configuration due to database failure")
	}
	if err := m.UnregisterAutomation(automationId, thisAutomation.Data, serverConfig); err != nil {
		return fmt.Errorf("failed to remove automation: could not unregister job: %s", err.Error())
	}

	// Delete automation
	if err := database.DeleteAutomationById(automationId); err != nil {
		log.Error("Failed to remove automation: database failure: ", err.Error())
		return err
	}

	log.Trace(fmt.Sprintf("Deactivated and removed automation. id: '%d'", automationId))
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
	for _, automationItem := range automationsTemp {
		var cronDescription *string

		if automationItem.Data.TriggerCronExpression != nil {
			cronDescriptionTemp, err := GenerateHumanReadableCronExpression(*automationItem.Data.TriggerCronExpression)
			if err != nil {
				log.Error("Failed to list automations of user: could not generate cron description: ", err.Error())
				return nil, err
			}
			cronDescription = &cronDescriptionTemp
		}

		automations = append(automations,
			Automation{
				Id:                     automationItem.Id,
				Name:                   automationItem.Data.Name,
				Description:            automationItem.Data.Description,
				HomescriptId:           automationItem.Data.HomescriptId,
				Owner:                  automationItem.Owner,
				Enabled:                automationItem.Data.Enabled,
				DisableOnce:            automationItem.Data.DisableOnce,
				Trigger:                automationItem.Data.Trigger,
				TriggerCronExpression:  automationItem.Data.TriggerCronExpression,
				CronDescription:        cronDescription,
				TriggerIntervalSeconds: automationItem.Data.TriggerIntervalSeconds,
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
	for _, automationItem := range automationsTemp {
		if automationItem.Id != automationId {
			continue // Skip any automations which don't match
		}

		var cronDescription *string
		if automationItem.Data.TriggerCronExpression != nil {
			cronDescriptionTemp, err := GenerateHumanReadableCronExpression(*automationItem.Data.TriggerCronExpression)
			if err != nil {
				log.Error("Failed to get user automation by id: could not generate cron description: ", err.Error())
				return Automation{}, false, err
			}
			cronDescription = &cronDescriptionTemp
		}

		return Automation{
			Id:                     automationItem.Id,
			Name:                   automationItem.Data.Name,
			Description:            automationItem.Data.Description,
			HomescriptId:           automationItem.Data.HomescriptId,
			Owner:                  automationItem.Owner,
			Enabled:                automationItem.Data.Enabled,
			DisableOnce:            automationItem.Data.DisableOnce,
			Trigger:                automationItem.Data.Trigger,
			TriggerCronExpression:  automationItem.Data.TriggerCronExpression,
			CronDescription:        cronDescription,
			TriggerIntervalSeconds: automationItem.Data.TriggerIntervalSeconds,
		}, true, nil
	}
	return Automation{}, false, nil
}

func (m AutomationManager) UnregisterAutomation(automationId uint, data database.AutomationData, config database.ServerConfig) error {
	switch data.Trigger {
	case database.TriggerCron, database.TriggerSunrise, database.TriggerSunset, database.TriggerInterval:
		// If the automation and the automation system are enabled, remove the underlying-job
		if data.Enabled && config.AutomationEnabled {
			// After the metadata has been changed, restart the scheduler
			if err := m.automationScheduler.RemoveByTag(fmt.Sprint(automationId)); err != nil {
				log.Error("Failed to unregister automation item: could not stop cron job: ", err.Error())
				return err
			}
		}
	case database.TriggerOnLogin, database.TriggerOnLogout, database.TriggerOnNotification, database.TriggerOnShutdown, database.TriggerOnBoot:
		// ignore these, they do not need to be unregistered
	default:
		panic("not implemented")
	}

	event.Debug("Deactivated Automation", fmt.Sprintf("Successfully deactivated automation '%s' (%d)", data.Name, automationId))
	log.Debug(fmt.Sprintf("Successfully deactivated automation '%s' (%d)", data.Name, automationId))
	return nil
}

func (m AutomationManager) RegisterAutomation(automationId uint, data database.AutomationData, config database.ServerConfig) error {
	// If the automation is disabled or the entire subsystem is shutdown, do not register this automation
	if !data.Enabled || !config.AutomationEnabled {
		event.Trace("Automation Skipped", fmt.Sprintf("Automation `%d` was not started", automationId))
		log.Debug(fmt.Sprintf("Skipping activation of automation '%d': automation (system) is disabled", automationId))
		return nil
	}

	switch data.Trigger {
	case database.TriggerCron, database.TriggerSunrise, database.TriggerSunset:
		newCronExpression := data.TriggerCronExpression

		// The cron expression must be updated if using suntimes
		if data.Trigger == database.TriggerSunrise || data.Trigger == database.TriggerSunset {
			// Calculate both the sunrise and sunset time
			sunRise, sunSet := CalculateSunRiseSet(config.Latitude, config.Longitude)

			// Select the time which is desired
			var finalTime SunTime
			if data.Trigger == database.TriggerSunrise {
				finalTime = sunRise
			} else {
				finalTime = sunSet
			}

			// Generate a cron expression from the sun time
			cronExpression, err := GenerateCronExpression(uint8(finalTime.Hour), uint8(finalTime.Minute), []uint8{0, 1, 2, 3, 4, 5, 6})
			if err != nil {
				return err
			}
			newCronExpression = &cronExpression
		}

		automationJob := m.automationScheduler.Cron(*newCronExpression)
		automationJob.Tag(fmt.Sprint(automationId))
		if _, err := automationJob.Do(AutomationRunnerFunc, automationId, types.AutomationContext{}); err != nil {
			log.Error("Failed to start automation, registering cron job failed: ", err.Error())
			return err
		}
	case database.TriggerInterval:
		automationJob := m.automationScheduler.Every(time.Second * time.Duration(*data.TriggerIntervalSeconds))
		automationJob.Tag(fmt.Sprint(automationId))
		if _, err := automationJob.Do(AutomationRunnerFunc, automationId, types.AutomationContext{}); err != nil {
			log.Error("Failed to start automation, registering cron job failed: ", err.Error())
			return err
		}
	case database.TriggerOnLogin, database.TriggerOnLogout, database.TriggerOnNotification, database.TriggerOnShutdown, database.TriggerOnBoot:
		// ignore these, they are triggered externally
	default:
		panic("not implemented")
	}

	event.Debug("Automation Activated", fmt.Sprintf("Successfully activated automation '%s' (%d)", data.Name, automationId))
	log.Debug(fmt.Sprintf("Successfully activated automation '%s' (%d)", data.Name, automationId))
	return nil
}

// Changes the metadata of a given automation, then restarts it so that it uses the updated values such as execution time
// Is also used after an automation with non-normal timing has been added
func (m AutomationManager) ModifyAutomationById(automationId uint, newAutomation database.AutomationData) error {
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil {
		return err
	}
	if !found {
		return errors.New("could not retrieve server configuration")
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

	if err := m.UnregisterAutomation(automationId, automationBefore.Data, serverConfig); err != nil {
		return fmt.Errorf("failed to modify automation: could not unregister automation: %s", err.Error())
	}

	if err := m.RegisterAutomation(automationId, newAutomation, serverConfig); err != nil {
		return fmt.Errorf("failed to modify automation: could not register new automation: %s", err.Error())
	}

	if err := database.ModifyAutomation(automationId, newAutomation); err != nil {
		log.Error("Failed to modify automation by id: database failure during modification: ", err.Error())
		return err
	}

	event.Debug("Automation Modified", fmt.Sprintf("Automation %d was modified", automationBefore.Id))
	return nil
}

// Given a jobId and whether sunrise or sunset should is activated, the next execution time is modified
// Used when an automation with non-normal Timing-Mode is executed in order to update its next start time
func (m AutomationManager) UpdateJobTime(id uint, config database.ServerConfig) error {
	// Retrieve the current job in order to get its current cron-expression (for the days)
	job, found, err := database.GetAutomationById(id)
	if err != nil {
		return fmt.Errorf("Could not update launch time: database failure: %s", err.Error())
	}
	if !found {
		return errors.New("Could not update launch time: invalid id supplied")
	}

	if err := m.UnregisterAutomation(id, job.Data, config); err != nil {
		return fmt.Errorf("Could not update launch time: unregistering failed: %s", err.Error())
	}

	if err := m.RegisterAutomation(id, job.Data, config); err != nil {
		return fmt.Errorf("Could not update launch time: registering failed: %s", err.Error())
	}

	log.Trace(fmt.Sprintf("Successfully updated the next execution time of automation '%d'", id))
	return nil
}
