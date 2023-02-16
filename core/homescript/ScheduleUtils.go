package homescript

import (
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
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
func CreateNewSchedule(data database.ScheduleData, owner string) (uint, error) {
	newScheduleId, err := database.CreateNewSchedule(owner, data)
	if err != nil {
		log.Error("Failed to create new schedule: database failure: ", err.Error())
		return 0, err
	}

	// Prepare the job for go-cron
	schedulerJob := scheduleScheduler.Every(1).Day().At(fmt.Sprintf("%02d:%02d", data.Hour, data.Minute))
	schedulerJob.Tag(fmt.Sprintf("%d", newScheduleId))
	schedulerJob.LimitRunsTo(1)
	if _, err := schedulerJob.Do(scheduleRunnerFunc, newScheduleId); err != nil {
		log.Error("Failed to create new schedule: could not register cron job: ", err.Error())
		return 0, err
	}
	event.Debug("Schedule Created", fmt.Sprintf("%s created Schedule `%s` (ID: %d)", owner, data.Name, newScheduleId))
	log.Trace(fmt.Sprintf("Successfully added and setup schedule '%d'", newScheduleId))
	return newScheduleId, nil
}

// Aborts and deletes a schedule based on its id
func RemoveScheduleById(id uint) error {
	if err := database.DeleteScheduleById(id); err != nil {
		log.Error("Failed to remove schedule: could not delete schedule from database: ", err.Error())
		return err
	}
	if err := scheduleScheduler.RemoveByTag(fmt.Sprintf("%d", id)); err != nil {
		log.Error("Failed to remove schedule: could not abort schedule: ", err.Error())
		return err
	}
	log.Trace(fmt.Sprintf("Successfully removed and aborted schedule '%d'", id))
	event.Debug("Schedule Removed", fmt.Sprintf("Schedule %d was removed from the system", id))
	return nil
}

// Modify an already set up schedule
// After the modification was performed, the schedule is restarted
func ModifyScheduleById(id uint, newSchedule database.ScheduleData) error {
	if err := database.ModifySchedule(id, newSchedule); err != nil {
		log.Error("Failed to modify schedule by id: ", err.Error())
		return err
	}
	if err := scheduleScheduler.RemoveByTag(fmt.Sprintf("%d", id)); err != nil {
		log.Error("Failed to modify schedule: could not abort schedule: ", err.Error())
		return err
	}
	// Prepare the job for go-cron
	schedulerJob := scheduleScheduler.Every(1).Day().At(fmt.Sprintf("%02d:%02d", newSchedule.Hour, newSchedule.Minute))
	schedulerJob.Tag(fmt.Sprintf("%d", id))
	schedulerJob.LimitRunsTo(1)
	if _, err := schedulerJob.Do(scheduleRunnerFunc, id); err != nil {
		log.Error("Failed to modify schedule: could not register cronjob after modification: ", err.Error())
		return err
	}
	log.Trace(fmt.Sprintf("Successfully added and setup schedule after modification: '%d'", id))
	event.Debug("Schedule Modified", fmt.Sprintf("Schedule %d was modified: new time: %d:%d ", newSchedule.Hour, newSchedule.Minute, id))
	return nil
}

// Gets a schedule based on its id and its owner's username
func GetUserScheduleById(username string, id uint) (database.Schedule, bool, error) {
	schedules, err := database.GetUserSchedules(username)
	if err != nil {
		return database.Schedule{}, false, err
	}
	for _, schedule := range schedules {
		if schedule.Id == id {
			return schedule, true, nil
		}
	}
	return database.Schedule{}, false, nil
}
