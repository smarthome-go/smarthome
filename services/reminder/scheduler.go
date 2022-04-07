package reminder

import (
	"time"

	"github.com/go-co-op/gocron"

	"github.com/MikMuellerDev/smarthome/core/database"
)

var scheduler *gocron.Scheduler

func reminderRunner() {
	var err error
	var users []database.User

	defer func(error) {
		if err != nil {
			log.Error("Failed to send notifications for reminders: ", err.Error())
		}
	}(err)

	users, err = database.ListUsers()
	if err != nil {
		return
	}

	for _, user := range users {
		err = SendUrgencyNotifications(user.Username)
		if err != nil {
			return
		}
	}
}

func InitSchedule() error {
	scheduler = gocron.NewScheduler(time.Local)
	runner := scheduler.Every(time.Hour)
	if _, err := runner.Do(reminderRunner); err != nil {
		log.Error("Failed to setup notification runner: ", err.Error())
		return err
	}
	runner.StartImmediately()
	scheduler.StartAsync()
	return nil
}
