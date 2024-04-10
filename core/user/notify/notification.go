package notify

import (
	"github.com/sirupsen/logrus"
	automationTypes "github.com/smarthome-go/smarthome/core/automation/types"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

type NotificationLevel uint8

const (
	NotificationLevelInfo  NotificationLevel = 1
	NotificationLevelWarn  NotificationLevel = 2
	NotificationLevelError NotificationLevel = 3
)

type Notification struct {
	ID          uint              `json:"id"`
	Priority    NotificationLevel `json:"priority"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Date        uint              `json:"date"` // Unix-Millis are used in this layer.
}

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

type NotificationManager struct {
	HmsManager types.Manager
	Automation automationTypes.AutomationManager
}

var Manager NotificationManager

func InitManager(hms types.Manager, automation automationTypes.AutomationManager) {
	Manager = NotificationManager{
		HmsManager: hms,
		Automation: automation,
	}
}

func (m NotificationManager) Notify(
	username string,
	title string,
	description string,
	level NotificationLevel,
	runHooks bool,
) (uint, error) {
	newID, err := database.AddNotification(username, title, description, uint8(level))
	if err != nil {
		log.Error("Failed to notify user: database failure: ", err.Error())
		return 0, err
	}

	// Run any notification hooks.
	if runHooks {
		notificationContext := types.NotificationContext{
			Id:          newID,
			Title:       title,
			Description: description,
			Level:       uint8(level),
		}
		go m.Automation.RunAllAutomationsWithTrigger(
			username,
			database.TriggerOnNotification,
			types.AutomationContext{
				NotificationContext: &notificationContext,
				MaximumHMSRuntime:   nil,
			},
		)
	}

	return newID, nil
}

func GetNotifications(username string) ([]Notification, error) {
	fromDB, err := database.GetUserNotifications(username)
	if err != nil {
		return nil, err
	}
	output := make([]Notification, 0)
	for _, notification := range fromDB {
		output = append(output, Notification{
			ID:          notification.Id,
			Priority:    NotificationLevel(notification.Priority),
			Name:        notification.Name,
			Description: notification.Description,
			Date:        uint(notification.Date.UnixMilli()),
		})
	}

	return output, nil
}
