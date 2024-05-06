package scheduler

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

type SchedulerManager struct {
	// This scheduler executes jobs only once, and then removes the given job
	scheduler *gocron.Scheduler
	hms       types.Manager
}

var Manager SchedulerManager

func InitManager(hms types.Manager) error {
	Manager = SchedulerManager{
		scheduler: gocron.NewScheduler(time.Local),
		hms:       hms,
	}
	Manager.scheduler.TagsUnique()
	if err := Manager.startSavedSchedules(); err != nil {
		return err
	}
	Manager.scheduler.StartAsync()
	return nil
}

// Retrieves saved schedules from the database and starts them
func (m SchedulerManager) startSavedSchedules() error {
	schedules, err := database.GetSchedules()
	if err != nil {
		log.Error("Failed to start schedules: database failure: ", err.Error())
		return err
	}
	for _, schedule := range schedules {
		// Prepare the job for go-cron
		schedulerJob := m.scheduler.Every(1).Day().At(
			fmt.Sprintf("%02d:%02d", schedule.Data.Hour, schedule.Data.Minute),
		)
		schedulerJob.Tag(fmt.Sprintf("%d", schedule.Id))
		schedulerJob.LimitRunsTo(1)
		if _, err := schedulerJob.Do(scheduleRunnerFunc, schedule.Id, &m); err != nil {
			log.Error("Failed to activates saved schedules: could not register cronjob: ", err.Error())
			return err
		}
		log.Trace(fmt.Sprintf("Successfully activated schedule '%d' of user '%s'", schedule.Id, schedule.Owner))
	}
	log.Debug("Successfully activated conventional scheduler")
	return nil
}
