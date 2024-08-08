package automation

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/homescript/types"
	"github.com/smarthome-go/smarthome/core/user/notify"
)

// Is called when the scheduler executes the given automation
// The AutomationRunnerFunc automatically tries to fetch the required configuration from the provided id
// Error handling is accomplished by logging to the internal event system and notifying the user about their automations failure
func AutomationRunnerFunc(id uint, automationCtx types.ExecutionContextAutomation) {
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
		if err := Manager.automationScheduler.RemoveByTag(fmt.Sprintf("%d", id)); err != nil {
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

	// Check if the automation's next run (this run) is disabled
	if job.Data.DisableOnce {
		// Re-enable the automation again
		if err := Manager.ModifyAutomationById(job.Id, database.AutomationData{
			Name:                   job.Data.Name,
			Description:            job.Data.Description,
			HomescriptId:           job.Data.HomescriptId,
			Enabled:                job.Data.Enabled,
			DisableOnce:            false,
			Trigger:                job.Data.Trigger,
			TriggerCronExpression:  job.Data.TriggerCronExpression,
			TriggerIntervalSeconds: job.Data.TriggerIntervalSeconds,
		}); err != nil {
			event.Error("Could not re-enable automation", fmt.Sprintf("Could not re-enable automation `%s`: %s", job.Data.Name, err.Error()))
			return
		}
		// Notify the user
		log.Info(fmt.Sprintf("Automation '%s' was skipped once", job.Data.Name))
		if _, err := notify.Manager.Notify(
			job.Owner,
			"Automation Skipped Once",
			fmt.Sprintf("Automation `%s` was skipped once. It will run regularely the next time.", job.Data.Name),
			notify.NotificationLevelInfo,
			false,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}

	// Update the execution time of the automation
	if err := database.UpdateAutomationLastRunTime(job.Id); err != nil {
		log.Error(fmt.Sprintf("Could not update `lastRun` of automation with ID `%d`: %s", job.Id, err.Error()))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation '%s' failed because its last run time could not be adjusted: %s", job.Data.Name, err.Error()),
		)
		return
	}

	// If the timing mode is set to either 'sunrise' or 'sunset', a new time with according cron-expression should be generated
	if job.Data.Trigger == database.TriggerSunrise || job.Data.Trigger == database.TriggerSunset {
		serverConfig, found, err := database.GetServerConfiguration()
		if err != nil || !found {
			log.Fatal("Could not retrieve server configuration")
			os.Exit(1)
		}

		// TODO: is this safe?
		if time.Since(*job.Data.LastRun).Minutes() < 5 {
			event.Trace("Geological Time Automation Skipped", fmt.Sprintf("The automation `%s` with ID `%d` was skipped due to cooldown.", job.Data.Name, job.Id))
			return
		}

		if err := Manager.UpdateJobTime(id, serverConfig); err != nil {
			log.Error("Failed to run automation: could not update next launch time: ", err.Error())
			event.Error(
				"Automation Failed",
				fmt.Sprintf("Automation '%s' failed because its next launch time could not be adjusted: %s", job.Data.Name, err.Error()),
			)
			if _, err := notify.Manager.Notify(
				job.Owner,
				"Automation Failed",
				fmt.Sprintf("Automation `%s` was not executed because the next time it should run could not be determined. This is caused by the automations trigger which is currently set to `%s`", job.Data.Name, job.Data.Trigger),
				notify.NotificationLevelError,
				false,
			); err != nil {
				log.Error("Failed to notify user: ", err.Error())
				return
			}
			return
		}
	}

	log.Debug(fmt.Sprintf("Automation '%d' is running", id))
	_, scriptExists, err := Manager.Hms.GetPersonalScriptById(job.Data.HomescriptId, job.Owner)
	if err != nil {
		log.Error(fmt.Sprintf("Automation '%s' failed because its Homescript Id could not be retrieved from the database: %s", job.Data.Name, err.Error()))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation '%s' could not be executed because it s Homescript Id could not be retrieved from the database: %s", job.Data.Name, err.Error()),
		)
		if _, err := notify.Manager.Notify(
			job.Owner,
			"Automation Failed",
			fmt.Sprintf("Automation `%s` was not executed because its referenced Homescript could not be found in the database due to an internal error, contact your administrator", job.Data.Name),
			notify.NotificationLevelError,
			false,
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
		if _, err := notify.Manager.Notify(
			job.Owner,
			"Automation Failed",
			fmt.Sprintf("Automation `%s` was not executed because its referenced Homescript could not be found in the database, contact your administrator", job.Data.Name),
			notify.NotificationLevelError,
			false,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}

	// var initiator types.HomescriptInitiator
	// switch job.Data.Trigger {
	// case database.TriggerOnNotification:
	// 	initiator = types.InitiatorAutomationOnNotify
	// default:
	// 	initiator = types.InitiatorAutomation
	// }

	// NOTE: If the automation context includes a maximum runtime, kill the script if it exceeds this timeout
	ctx, cancel := context.WithCancel(context.Background())
	if automationCtx.Inner.MaximumHMSRuntime != nil {
		ctx, cancel = context.WithTimeout(context.Background(), *automationCtx.Inner.MaximumHMSRuntime)
	}

	res, err := Manager.Hms.RunUserScriptTweakable(
		job.Data.HomescriptId,
		job.Owner,
		nil,
		types.Cancelation{
			Context:    ctx,
			CancelFunc: cancel,
		},
		&bytes.Buffer{},
		nil,
		false,
		&automationCtx.Inner,
	)

	if err != nil {
		log.Warn(fmt.Sprintf("Automation '%s' failed during the execution of Homescript: '%s', which terminated abnormally", job.Data.Name, job.Data.HomescriptId))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation '%s' failed during execution of Homescript '%s'. Error: %s", job.Data.Name, job.Data.HomescriptId, err.Error()),
		)
		if _, err := notify.Manager.Notify(
			job.Owner,
			"Automation Failed",
			fmt.Sprintf("Automation `%s` failed during execution of Homescript `%s`.\nError:\n```\n%s\n```", job.Data.Name, job.Data.HomescriptId, strings.ReplaceAll(err.Error(), "`", "\\`")),
			notify.NotificationLevelError,
			false,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}

	if res.Errors.ContainsError {
		log.Warn(fmt.Sprintf("Automation '%s' failed during the execution of Homescript: '%s', which terminated abnormally", job.Data.Name, job.Data.HomescriptId))
		event.Error(
			"Automation Failed",
			fmt.Sprintf("Automation '%s' failed during execution of Homescript '%s'. Error: %s", job.Data.Name, job.Data.HomescriptId, res.Errors.Diagnostics[0]),
		)

		if _, err := notify.Manager.Notify(
			job.Owner,
			"Automation Failed",
			fmt.Sprintf(
				"Automation '**%s**' failed during execution of Homescript '**%s**'.\n```\n%s\n```",
				job.Data.Name,
				job.Data.HomescriptId,
				strings.ReplaceAll(res.Errors.Diagnostics[0].String(),
					"`",
					"\\`",
				)),
			notify.NotificationLevelError,
			false,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		return
	}

	event.Debug(
		"Automation Executed Successfully",
		fmt.Sprintf("Automation `%s` (%d) of user '%s' has executed successfully.", job.Data.Name, id, job.Owner),
	)
}
