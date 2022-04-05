package scheduler

import (
	"fmt"

	"github.com/MikMuellerDev/smarthome/core/database"
)

// Used for listing personal schedules
type UserSchedule struct {
	Id             uint   `json:"id"`
	Name           string `json:"name"`
	Hour           uint   `json:"hour"`
	Minute         uint   `json:"minute"`
	NextRun        string `json:"nextRun"`
	HomescriptCode string `json:"homescriptCode"` // Will be executed if the scheduler runs the job
}

// Creates and starts a schedule based on the provided input data
func CreateNewSchedule(schedule database.Schedule) error {
	newScheduleId, err := database.CreateNewSchedule(schedule)
	if err != nil {
		log.Error("Failed to create new schedule: database failure: ", err.Error())
		return err
	}
	// Prepare the job for go-cron
	schedulerJob := scheduler.Every(1).Day().At(fmt.Sprintf("%02d:%02d", schedule.Hour, schedule.Minute))
	schedulerJob.Tag(fmt.Sprintf("%d", newScheduleId))
	schedulerJob.LimitRunsTo(1)
	if _, err := schedulerJob.Do(scheduleRunnerFunc, newScheduleId); err != nil {
		log.Error("Failed to create new schedule: could not register cron job: ", err.Error())
		return err
	}
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

// Modify an already set up schedule
func ModifyScheduleById(id uint, newSchedule database.Schedule) error {
	if err := database.ModifySchedule(id, database.ScheduleWithoudIdAndUsername{
		Name:           newSchedule.Name,
		Hour:           newSchedule.Hour,
		Minute:         newSchedule.Minute,
		HomescriptCode: newSchedule.HomescriptCode,
	}); err != nil {
		log.Error("Failed to modify schedule by id: ", err.Error())
		return err
	}
	if err := scheduler.RemoveByTag(fmt.Sprintf("%d", id)); err != nil {
		log.Error("Failed to modify schedule: could not abort schedule: ", err.Error())
		return err
	}
	// Prepare the job for go-cron
	schedulerJob := scheduler.Every(1).Day().At(fmt.Sprintf("%02d:%02d", newSchedule.Hour, newSchedule.Minute))
	schedulerJob.Tag(fmt.Sprintf("%d", id))
	schedulerJob.LimitRunsTo(1)
	if _, err := schedulerJob.Do(scheduleRunnerFunc, id); err != nil {
		log.Error("Failed to modify schedule: could not register cronjob after modification: ", err.Error())
		return err
	}
	log.Trace(fmt.Sprintf("Successfully added and setup schedule after modification'%d'", id))
	return nil
}

// Gets a schedule based on its id and its owner's username
func GetUserScheduleById(username string, id uint) (database.Schedule, bool, error) {
	schedules, err := database.GetUserSchedules(username)
	if err != nil {
		log.Error("Failed to get user schedule: database error")
		return database.Schedule{}, false, err
	}
	for _, schedule := range schedules {
		if schedule.Id == id {
			return schedule, true, nil
		}
	}
	return database.Schedule{}, false, nil
}
