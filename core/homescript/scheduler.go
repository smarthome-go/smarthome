package homescript

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"

	"github.com/smarthome-go/smarthome/core/database"
)

// This scheduleScheduler is executed only once, then disabled the job it should run
var scheduleScheduler *gocron.Scheduler

func InitScheduler() error {
	scheduleScheduler = gocron.NewScheduler(time.Local)
	scheduleScheduler.TagsUnique()
	if err := startSavedSchedules(); err != nil {
		return err
	}
	scheduleScheduler.StartAsync()
	return nil
}

// Retrieves saved schedules from the database and starts them
func startSavedSchedules() error {
	schedules, err := database.GetSchedules()
	if err != nil {
		logger.Error("Failed to start schedules: database failure: ", err.Error())
		return err
	}
	for _, schedule := range schedules {
		// Prepare the job for go-cron
		schedulerJob := scheduleScheduler.Every(1).Day().At(fmt.Sprintf("%02d:%02d", schedule.Data.Hour, schedule.Data.Minute))
		schedulerJob.Tag(fmt.Sprintf("%d", schedule.Id))
		schedulerJob.LimitRunsTo(1)
		if _, err := schedulerJob.Do(scheduleRunnerFunc, schedule.Id); err != nil {
			logger.Error("Failed to activates saved schedules: could not register cronjob: ", err.Error())
			return err
		}
		logger.Trace(fmt.Sprintf("Successfully activated schedule '%d' of user '%s'", schedule.Id, schedule.Owner))
	}
	logger.Debug("Successfully activated conventional scheduler")
	return nil
}
