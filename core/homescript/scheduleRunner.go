package homescript

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
)

const SCHEDULE_MAXIMUM_HOMESCRIPT_RUNTIME = time.Minute * 10

// Executes a given scheduler
// If the user's schedulers are currently disabled
// the job runner will still be executed and remove the current scheduler but without running the homescript
// Error handling works in a similar way to the runner of the automation system
func scheduleRunnerFunc(id uint) {
	job, jobFound, err := database.GetScheduleById(id)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to run schedule '%s': database failure whilst retrieving job information: %s", job.Data.Name, err.Error()))
		return
	}
	if !jobFound {
		log.Error(fmt.Sprintf("Failed to run schedule '%s': no metadata saved in the database: %s", job.Data.Name, err.Error()))
		// Abort this schedule to avoid future errors
		if err := scheduleScheduler.RemoveByTag(fmt.Sprintf("%d", id)); err != nil {
			log.Error("Failed to remove dangling schedule: could not abort schedule: ", err.Error())
			return
		}
		log.Info(fmt.Sprintf("Successfully aborted dangling schedule: %d", id))
		return
	}
	// Delete the schedule from the database
	if err := database.DeleteScheduleById(id); err != nil {
		log.Error("Removing schedule failed: could not remove schedule from database: ", err.Error())
		return
	}
	// Check if the user has blocked their automations & schedules
	owner, found, err := database.GetUserByUsername(job.Owner)
	if err != nil {
		log.Error("Schedule failed because owner user could not be determined")
		return
	}
	if !found {
		log.Warn("Schedule failed because owner user does not exist anymore, skipping execution...")
		return
	}
	if !owner.SchedulerEnabled {
		log.Debug(fmt.Sprintf("Schedule '%s' was not executed because its owner has disabled their schedules & automations", job.Data.Name))
		return
	}
	if !owner.SchedulerEnabled {
		log.Debug(fmt.Sprintf("Not running schedule '%s' because user's schedules are currently disabled", job.Data.Name))
		if _, err := Notify(
			owner.Username,
			"Schedule Skipped",
			fmt.Sprintf("Schedule '%s' was discarded because your schedules are disabled", job.Data.Name),
			NotificationLevelWarn,
			true,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		event.Debug(
			"Schedule Skipped",
			fmt.Sprintf("Schedule '%s' has been skipped", job.Data.Name),
		)
		return
	}
	log.Debug(fmt.Sprintf("Schedule '%s' (%d) is executing...", job.Data.Name, id))
	switch job.Data.TargetMode {
	case database.ScheduleTargetModeCode:
		ctx, cancel := context.WithTimeout(context.Background(), SCHEDULE_MAXIMUM_HOMESCRIPT_RUNTIME)

		res, _, err := HmsManager.Run(
			HMS_PROGRAM_KIND_NORMAL,
			nil,
			owner.Username,
			nil,
			job.Data.HomescriptCode,
			InitiatorSchedule,
			ctx,
			cancel,
			nil,
			nil,
			&bytes.Buffer{},
			nil,
			// Do not use any custom runner func.
			nil,
			nil,
		)

		if err != nil {
			log.Error("Executing schedule's Homescript failed: ", err.Error())
			if _, err := Notify(
				owner.Username,
				"Schedule Failed",
				fmt.Sprintf("Schedule '%s' failed due to Homescript error: %s", job.Data.Name, err.Error()),
				NotificationLevelError,
				true,
			); err != nil {
				log.Error("Failed to notify user: ", err.Error())
				return
			}
			event.Error(
				"Schedule Failure",
				fmt.Sprintf("Schedule '%d' failed. Error: %s", id, err.Error()),
			)
			return
		}

		if !res.Success {
			log.Error("Executing schedule's Homescript failed: ", res.Errors[0])
			if _, err := Notify(
				owner.Username,
				"Schedule Failed",
				fmt.Sprintf("Schedule '%s' failed due to Homescript error: %s", job.Data.Name, res.Errors[0]),
				NotificationLevelError,
				true,
			); err != nil {
				log.Error("Failed to notify user: ", err.Error())
				return
			}
			event.Error(
				"Schedule Failure",
				fmt.Sprintf("Schedule '%d' failed. Error: %s", id, res.Errors[0]),
			)
			return
		}
	case database.ScheduleTargetModeHMS:
		ctx, cancel := context.WithTimeout(context.Background(), SCHEDULE_MAXIMUM_HOMESCRIPT_RUNTIME)

		res, _, err := HmsManager.RunById(
			HMS_PROGRAM_KIND_NORMAL,
			nil,
			job.Data.HomescriptTargetId,
			owner.Username,
			InitiatorSchedule,
			ctx,
			cancel,
			nil,
			nil,
			&bytes.Buffer{},
			nil,
			nil,
		)

		if err != nil {
			log.Error("Executing schedule's Homescript failed: ", err.Error())
			if _, err := Notify(
				owner.Username,
				"Schedule Failed",
				fmt.Sprintf("Schedule '%s' failed due to Homescript system error: %s", job.Data.Name, err.Error()),
				NotificationLevelError,
				true,
			); err != nil {
				log.Error("Failed to notify user: ", err.Error())
				return
			}
			event.Error(
				"Schedule Failure",
				fmt.Sprintf("Schedule '%d' failed. Error: %s", id, err.Error()),
			)
			return
		}
		if !res.Success {
			log.Error("Executing schedule's Homescript failed: ", res.Errors[0])
			if _, err := Notify(
				owner.Username,
				"Schedule Failed",
				fmt.Sprintf("Schedule '%s' failed due to Homescript execution error: %s", job.Data.Name, res.Errors[0]),
				NotificationLevelError,
				true,
			); err != nil {
				log.Error("Failed to notify user: ", err.Error())
				return
			}
			event.Error(
				"Schedule Failure",
				fmt.Sprintf("Schedule '%d' failed. Error: %s", id, res.Errors[0]),
			)
			return
		}
	case database.ScheduleTargetModeDevices:
		for _, switchJob := range job.Data.SwitchJobs {
			// Validate if the user still has permission to perform this power job
			hasPermission, err := database.UserHasDevicePermission(job.Owner, switchJob.DeviceId)
			if err != nil {
				log.Errorf("Schedule '%d' failed. Error: %s", id, err.Error())
				return
			}

			if !hasPermission {
				log.Warn("Executing schedule's switch jobs failed: user now lacks permission to use switch")
				if _, err := Notify(
					owner.Username,
					"Schedule Failed",
					fmt.Sprintf("Schedule '%s' failed due to lacking switch permissions: you now lack permission to use %s", job.Data.Name, switchJob.DeviceId),
					NotificationLevelError,
					true,
				); err != nil {
					log.Error("Failed to notify user: ", err.Error())
					return
				}
				event.Warn(
					"Schedule Failure",
					fmt.Sprintf("Schedule '%d' failed due to changed permissions.", id),
				)
			}

			switchData, found, err := database.GetDeviceById(switchJob.DeviceId)
			if err != nil {
				log.Errorf("Schedule '%d' failed. Error: %s", id, err.Error())
				return
			}

			if !found {
				log.Errorf("Schedule '%d' is being executed even though a switch was removed. Error: %s", id, err.Error())
				return
			}

			_, hmsErrs, err := InvokeDriverSetPower(
				switchJob.DeviceId,
				switchData.VendorId,
				switchData.ModelId,
				DriverActionPower{State: switchJob.PowerOn},
			)

			if err != nil {
				log.Errorf("Schedule '%d' failed. Error: %s", id, err.Error())
				return
			}

			if hmsErrs != nil {
				if _, err := Notify(
					owner.Username,
					"Schedule Failed",
					fmt.Sprintf("Schedule '%s' failed due to Homescript execution error: %s", job.Data.Name, hmsErrs[0]),
					NotificationLevelError,
					true,
				); err != nil {
					log.Error("Failed to notify user: ", err.Error())
					return
				}
				event.Error(
					"Schedule Failure",
					fmt.Sprintf("Schedule '%d' failed. Error: %s", id, hmsErrs[0]),
				)
				return
			}
		}
	default:
		log.Error("Unimplemented schedule mode")
	}
	event.Info("Schedule Executed Successfully",
		fmt.Sprintf("Schedule '%s' of user '%s' has been executed successfully", job.Data.Name, job.Owner),
	)
}
