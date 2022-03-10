package database

import (
	"database/sql"
	"errors"
)

// Many notifications are always meant to address one user
// Will later be used in `core/user`

// Creates the notification table unless it exists, returns an error if the database fails
func CreateNotificationTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	notifications(
		Id INT AUTO_INCREMENT,
		Username VARCHAR(20),
		Name VARCHAR(100),
		Description TEXT,
		Date DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(Id),
		CONSTRAINT NotificationUsername
		FOREIGN KEY (Username)
		REFERENCES user(Username)
		)`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create notification table: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Adds a new notification to a user's `inbox`, can return an error if the database fails
func AddNotification(receiverUsername string, name string, description string) error {
	query, err := db.Prepare(`
	INSERT INTO
	notifications(
		Id,
		Username,
		Name,
		Description,
		Date
	)
	VALUES (DEFAULT, ?, ?, ?, DEFAULT)
	`)
	if err != nil {
		log.Error("Failed to add notification: preparing query failed: ", err.Error())
		return err
	}
	_, err = query.Exec(receiverUsername, name, description)
	if err != nil {
		log.Error("Failed to add notification: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// If the user requests to empty their notification area, all hist notifications will be deleted
func DeleteAllNotificationsFromUser(username string) error {
	query, err := db.Prepare(`
	DELETE FROM
	notifications
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to remove all user notifications: preparing query failed: ", err.Error())
		return err
	}
	if _, err = query.Exec(username); err != nil {
		log.Error("Failed to remove all user notifications: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes a given notification, can return an error
func DeleteNotificationById(notificationId uint) error {
	query, err := db.Prepare(`
	DELETE FROM
	notifications
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to delete notification by id: preparing query failed: ", err.Error())
		return err
	}
	_, err = query.Exec(notificationId)
	if err != nil {
		log.Error("Failed to delete notification by id: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Used for displaying the notification count for a given user
// Used in the frontend before the actual permissions are fetched
func GetUserNotificationCount(username string) (uint16, error) {
	query, err := db.Prepare(`
	SELECT COUNT(Id)
	AS Count
	FROM notifications
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to get notification count: preparing query failed: ", err.Error())
		return 0, err
	}
	var count uint16 = 0
	err = query.QueryRow(username).Scan(&count)
	if err != nil {
		log.Error("Failed to get notification count: executing query failed: ", err.Error())
		return 0, err
	}
	return count, nil
}

// Used when requesting the user's permissions in the frontend
// Returns a list containing the permissions of a given user
func GetUserNotifications(username string) ([]Notification, error) {
	query, err := db.Prepare(`
	SELECT Id, Name, Description, Date
	FROM notifications
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to get notifications: preparing query failed: ", err.Error())
		return []Notification{}, err
	}
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to get notifications: executing query failed: ", err.Error())
	}
	notifications := make([]Notification, 0)
	for res.Next() {
		var notificationItem Notification
		var notificationTime sql.NullTime
		err := res.Scan(
			&notificationItem.Id,
			&notificationItem.Name,
			&notificationItem.Description,
			&notificationTime,
		)
		if err != nil {
			log.Error()
			return []Notification{}, err
		}
		if !notificationTime.Valid {
			log.Error("Failed tp get notifications: notification time is not valid: critical failure")
			return []Notification{}, errors.New("critical error: notification date column contains null value")
		}
		notificationItem.Date = notificationTime.Time
		notifications = append(notifications, notificationItem)
	}
	return notifications, nil
}
