package scheduler

import (
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/core/user"
)

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
		return
	}
	owner, found, err := database.GetUserByUsername(job.Owner)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to run schedule '%s': database error whilst retrieving user information: %s", job.Data.Name, err.Error()))
		return
	}
	if !found {
		log.Error(fmt.Sprintf("Owner %s of schedule %d does not exist", job.Owner, job.Id))
		return
	}
	if err := database.DeleteScheduleById(id); err != nil {
		log.Error("Executing schedule failed: could not remove schedule from database: ", err.Error())
		return
	}
	if !owner.SchedulerEnabled {
		log.Info(fmt.Sprintf("Not running schedule '%s' because user's schedules are currently disabled", job.Data.Name))
		if err := user.Notify(
			owner.Username,
			"Schedule Skipped",
			fmt.Sprintf("Schedule '%s' was discarded because your schedules are disabled", job.Data.Name),
			user.NotificationLevelWarn,
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
	log.Debug(fmt.Sprintf("Schedule '%d' is running", id))
	_, _, hmsErrors := homescript.HmsManager.Run(
		owner.Username,
		fmt.Sprintf("%d.hms", id),
		job.Data.HomescriptCode,
		false,
		make(map[string]string, 0),
		make([]string, 0),
		homescript.InitiatorScheduler,
	)
	if len(hmsErrors) > 0 {
		log.Error("Executing schedule's Homescript failed: ", hmsErrors[0].ErrorType)
		if err := user.Notify(
			owner.Username,
			"Schedule Failed",
			fmt.Sprintf("Schedule '%s' failed due to Homescript error: %s", job.Data.Name, hmsErrors[0].Message),
			user.NotificationLevelError,
		); err != nil {
			log.Error("Failed to notify user: ", err.Error())
			return
		}
		event.Error(
			"Schedule Failure",
			fmt.Sprintf("Schedule '%d' failed. Error: %s", id, hmsErrors[0].Message),
		)
		return
	}
	if err := user.Notify(
		job.Owner,
		"Schedule Executed Successfully",
		fmt.Sprintf("Schedule '%s' has been executed successfully", job.Data.Name),
		1,
	); err != nil {
		log.Error("Failed to notify user about failing schedule: ", err.Error())
		return
	}
	event.Info("Schedule Executed Successfully",
		fmt.Sprintf("Schedule '%s' of user '%s' has been executed successfully", job.Data.Name, job.Owner),
	)
}
