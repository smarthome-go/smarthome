package scheduler

import (
	"fmt"

	"github.com/MikMuellerDev/smarthome/core/database"
)

// Creates and starts a schedule based on the provided input data
func CreateNewSchedule(schedule database.Schedule) error {
	newScheduleId, err := database.CreateNewSchedule(schedule)
	if err != nil {
		log.Error("Failed to create new schedule: database failure: ", err.Error())
		return err
	}
	// Prepare the job for go-cron
	automationJob := scheduler.Every(1).Day().At(fmt.Sprintf("%02d:%02d", schedule.Hour, schedule.Minute))
	automationJob.Tag(fmt.Sprintf("%d", newScheduleId))
	automationJob.LimitRunsTo(1)
	automationJob.Do(scheduleRunnerFunc, newScheduleId)
	log.Trace(fmt.Sprintf("Successfully added and setup schedule '%d'", newScheduleId))
	return nil
}

// Aborts and deletes a schedule based on its id
func RemoveScheduleById(id uint) error {
	if err := database.DeleteScheduleById(id); err != nil {
		log.Error("Failed to remove schedule: could not delete schedule from database: ", err.Error())
		return err
	}
	if err := scheduler.RemoveByTag(fmt.Sprintf("%d", id)); err != nil {
		log.Error("Failed to remove schedule: could not abort schedule: ", err.Error())
		return err
	}
	log.Trace(fmt.Sprintf("Successfully removed and aborted schedule '%d'", id))
	return nil
}

// Modify an already set up sc