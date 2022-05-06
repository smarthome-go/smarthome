package reminder

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/user"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Send a fitting notification which informs the reminders owner of the remaining time to complete the specified task
func sendDueInReminder(id uint, username string, name string, daysLeft uint, dueDate time.Time, level user.NotificationLevel) error {
	daysLeft++ // The current day should be added to the days the user has left to complete the task
	dayText := "days"
	if daysLeft == 1 {
		dayText = "day"
	}
	if err := user.Notify(
		username,
		fmt.Sprintf("Task is due on %s", dueDate.Format("Monday, 2.1.2006")),
		fmt.Sprintf("You have %d %s left to complete the task '%s'",
			daysLeft,
			dayText,
			name,
		),
		level,
	); err != nil {
		return err
	}
	log.Trace(fmt.Sprintf("Successfully sent reminder notification for '%d' to user '%s'", id, username))
	// Mark that the user has been notified and update it to the current time
	return database.SetReminderUserWasNotified(id, true, time.Now().Local())
}

// Will be executed by a scheduler in order to send periodic notifications to the user
func SendUrgencyNotifications(username string) error {
	reminders, err := database.GetUserReminders(username)
	if err != nil {
		return err
	}

	now := time.Now()

	for _, reminder := range reminders {
		// The due date will likely be in the future
		remainingDays := int(reminder.DueDate.Sub(now).Hours() / 24)
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
					reminder.DueDate.Format("Monday, January 2 2006"),
					remainingDays*-1,
					dayText,
				),
				user.NotificationLevelError,
			); err != nil {
				return err
			}
			// Mark that the user has been notified and update it to the current time
			if err := database.SetReminderUserWasNotified(reminder.Id, true, time.Now().Local()); err != nil {
				return err
			}
			continue // Continue to the next reminder
		}

		switch reminder.Priority {
		case database.Low:
			continue // The low priority will not trigger any notification
		case database.Normal:
			if remainingDays < 2 {
				if err := sendDueInReminder(reminder.Id, username, reminder.Name, uint(remainingDays), reminder.DueDate, user.NotificationLevelInfo); err != nil {
					log.Error("Failed to send notification for reminder: ", err.Error())
					return err
				}
			}
			continue
		case database.Medium:
			if remainingDays < 3 {
				if err := sendDueInReminder(reminder.Id, username, reminder.Name, uint(remainingDays), reminder.DueDate, user.NotificationLevelInfo); err != nil {
					log.Error("Failed to send notification for reminder: ", err.Error())
					return err
				}
			}
			continue
		case database.High:
			if remainingDays < 4 {
				if err := sendDueInReminder(reminder.Id, username, reminder.Name, uint(remainingDays), reminder.DueDate, user.NotificationLevelWarn); err != nil {
					log.Error("Failed to send notification for reminder: ", err.Error())
					return err
				}
			}
			continue
		case database.Urgent:
			if remainingDays < 5 {
				if err := sendDueInReminder(reminder.Id, username, reminder.Name, uint(remainingDays), reminder.DueDate, user.NotificationLevelError); err != nil {
					log.Error("Failed to send notification for reminder: ", err.Error())
					return err
				}
			}
			continue
		default:
			log.Warn(fmt.Sprintf("Invalid priority for reminder: '%d': priority: %d", reminder.Id, reminder.Priority))
			continue
		}
	}
	return nil
}
