package reminder

import (
	"time"

	"github.com/go-co-op/gocron"

	"github.com/smarthome-go/smarthome/core/database"
)

var scheduler *gocron.Scheduler

func reminderRunner() {
	var err error
	var users []database.User

	users, err = database.ListUsers()
	if err != nil {
		log.Error("Failed to send notifications for reminders: cannot get user list: ", err.Error())
		return
	}

	for _, user := range users {
		if err := SendUrgencyNotifications(user.Username); err != nil {
			log.Error("Failed to send notifications for reminders: ", err.Error())
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
