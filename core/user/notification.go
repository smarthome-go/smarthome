package user

import "github.com/MikMuellerDev/smarthome/core/database"

type NotificationLevel uint8

const (
	NotificationLevelInfo  NotificationLevel = 1
	NotificationLevelWarn  NotificationLevel = 2
	NotificationLevelError NotificationLevel = 3
)

func Notify(username string, title string, description string, level NotificationLevel) error {
	if err := database.AddNotification(username, title, description, uint8(level)); err != nil {
		log.Error("Failed to notify user: database failure: ", err.Error())
		return err
	}
	return nil
}
