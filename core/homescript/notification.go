package homescript

import (
	"github.com/smarthome-go/smarthome/core/database"
)

type NotificationLevel uint8

const (
	NotificationLevelInfo  NotificationLevel = 1
	NotificationLevelWarn  NotificationLevel = 2
	NotificationLevelError NotificationLevel = 3
)

type Notification struct {
	Id          uint              `json:"id"`
	Priority    NotificationLevel `json:"priority"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Date        uint              `json:"date"` // Unix-Millis are used in this layer
}

func Notify(username string, title string, description string, level NotificationLevel, run_hooks bool) (uint, error) {
	newId, err := database.AddNotification(username, title, description, uint8(level))
	if err != nil {
		logger.Error("Failed to notify user: database failure: ", err.Error())
		return 0, err
	}

	// Run any notification hooks
	if run_hooks {
		notificationContext := NotificationContext{
			Id:          newId,
			Title:       title,
			Description: description,
			Level:       uint8(level),
		}
		go RunAllAutomationsWithTrigger(
			username,
			database.TriggerOnNotification,
			AutomationContext{NotificationContext: &notificationContext},
		)
	}

	return newId, nil
}

func GetNotifications(username string) ([]Notification, error) {
	fromDB, err := database.GetUserNotifications(username)
	if err != nil {
		return nil, err
	}
	output := make([]Notification, 0)
	for _, notification := range fromDB {
		output = append(output, Notification{
			Id:          notification.Id,
			Priority:    NotificationLevel(notification.Priority),
			Name:        notification.Name,
			Description: notification.Description,
			Date:        uint(notification.Date.UnixMilli()),
		})
	}

	return output, nil
}
