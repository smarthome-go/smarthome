package reminder

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/user"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Send a fitting notification which informs the reminders owner of the remaining time to complete the specified task
func sendDueInReminder(username string, name string, daysLeft uint, level user.NotificationLevel) error {
	dayText := "days"
	if daysLeft == 1 {
		dayText = "day"
	}
	if err := user.Notify(
		username,
		fmt.Sprintf("Task '%s' is due in %d %s", name, daysLeft, dayText),
		fmt.Sprintf("You have %d %s left to complete the task '%s'.",
			daysLeft,
			dayText,
			name,
		),
		level,
	); err != nil {
		return err
	}
	// Mark that the user has been notified and update it to the current time
	return database.SetReminderUserWasNotified(true, time.Now().Local())
}

// Will be executed by a scheduler in order to send periodic notifications to the user
func SendUrgencyNotifications(username string) (uint, error) {
	var notificationsSent uint

	reminders, err := database.GetUserReminders(username)
	if err != nil {
		return 0, err
	}

	now := time.Now()

	for _, reminder := range reminders {
		// The due date will likely be in the future
		remainingDays := reminder.DueDate.Sub(now)
		dayText := "days"
		if remainingDays == -1 || remainingDays == 1 {
			dayText = "day"
		}

		// If the last notification was sent in the last 24 hours, skip the current notification
		if now.Sub(reminder.UserWasNotifiedAt).Hours() < 24 {
			continue
		}

		if remainingDays < 0 && reminder.Priority > database.Low {
			if err := user.Notify(
				username,
				fmt.Sprintf("Reminder '%s' Is Overdue", reminder.Name),
				fmt.Sprintf("The task '%s' was supposed to be finished on %s. You are behind schedule by %d %s.",
					reminder.Name,
					reminder.DueDate.Format(time.Kitchen),
					remainingDays*-1,
					dayText,
				),
				user.NotificationLevelError,
			); err != nil {
				return 0, err
			}
			// Mark that the user has been notified and update it to the current time
			if err := database.SetReminderUserWasNotified(true, time.Now().Local()); err != nil {
				return 0, err
			}
			notificationsSent++
			continue // Continue to the next reminder
		}

		switch reminder.Priority {
		case database.Low:
			continue // The low priority will not trigger any notification
		case database.Normal:
			if remainingDays < 2 {
				if err := sendDueInReminder(username, reminder.Name, uint(remainingDays), user.NotificationLevelInfo); err != nil {
					log.Error("Failed to send notification for reminder: ", err.Error())
					return 0, err
				}
				notificationsSent++
			}
			continue
		case database.Medium:
			if remainingDays < 3 {
				if err := sendDueInReminder(username, reminder.Name, uint(remainingDays), user.NotificationLevelInfo); err != nil {
					log.Error("Failed to send notification for reminder: ", err.Error())
					return 0, err
				}
				notificationsSent++
			}
			continue
		case database.High:
			if remainingDays < 4 {
				if err := sendDueInReminder(username, reminder.Name, uint(remainingDays), user.NotificationLevelWarn); err != nil {
					log.Error("Failed to send notification for reminder: ", err.Error())
					return 0, err
				}
				notificationsSent++
			}
			continue
		case database.Urgent:
			if remainingDays < 5 {
				if err := sendDueInReminder(username, reminder.Name, uint(remainingDays), user.NotificationLevelError); err != nil {
					log.Error("Failed to send notification for reminder: ", err.Error())
					return 0, err
				}
				notificationsSent++
			}
			continue
		}
	}
	return notificationsSent, nil
}
