package scheduler

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
)

// This scheduler is executed only once, then disabled the job it should run
var scheduler *gocron.Scheduler

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func Init() error {
	scheduler = gocron.NewScheduler(time.Local)
	scheduler.TagsUnique()
	if err := startSavedSchedules(); err != nil {
		return err
	}
	scheduler.StartAsync()
	return nil
}

// Retrieves saved schedules from the database and starts them
func startSavedSchedules() error {
	schedules, err := database.GetSchedules()
	if err != nil {
		log.Error("Failed to start schedules: database failure: ", err.Error())
		return err
	}
	for _, schedule := range schedules {
		// Prepare the job for go-cron
		schedulerJob := scheduler.Every(1).Day().At(fmt.Sprintf("%02d:%02d", schedule.Data.Hour, schedule.Data.Minute))
		schedulerJob.Tag(fmt.Sprintf("%d", schedule.Id))
		schedulerJob.LimitRunsTo(1)
		if _, err := schedulerJob.Do(scheduleRunnerFunc, schedule.Id); err != nil {
			log.Error("Failed to activates saved schedules: could not register cronjob: ", err.Error())
			return err
		}
		log.Trace(fmt.Sprintf("Successfully activated schedule '%d' of user '%s'", schedule.Id, schedule.Owner))
	}
	log.Debug("Successfully activated conventional scheduler")
	return nil
}
