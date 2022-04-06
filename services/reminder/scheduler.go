package reminder

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"

	"github.com/MikMuellerDev/smarthome/core/database"
)

var scheduler *gocron.Scheduler

func reminderRunner() {
	log.Trace("Checking for urgent reminders...")
	var err error
	var users []database.User
	var notificationsSent uint

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
		notificationsSent, err = SendUrgencyNotifications(user.Username)
		if err != nil {
			return
		}
	}
	log.Trace(fmt.Sprintf("Successfully sent %d notifications for reminders", notificationsSent))
}

func InitSchedule() error {
	scheduler = gocron.NewScheduler(time.Local)
	runner := scheduler.Every(time.Minute)
	if _, err := runner.Do(reminderRunner); err != nil {
		log.Error("Failed to setup notification runner: ", err.Error())
		return err
	}
	scheduler.StartAsync()
	return nil
}
