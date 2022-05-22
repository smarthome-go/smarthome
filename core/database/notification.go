package database

import (
	"database/sql"
	"errors"
	"time"
)

// Many notifications are always meant to address one user
// Will later be used in `core/user`

// User notification
type Notification struct {
	Id          uint      `json:"id"`
	Priority    uint8     `json:"priority"` // Includes 1: info, 2: warning, 3: alert
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	// Username is left out due to not being required in the service layer
}

// Creates the notification table unless it exists, returns an error if the database fails
func createNotificationTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	notifications(
		Id INT AUTO_INCREMENT,
		Username VARCHAR(20),
		Priority INT,
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
func AddNotification(receiverUsername string, name string, description string, priority uint8) error {
	if priority > 3 || priority < 1 {
		log.Error("Invalid Priority range")
		return errors.New("failed to send notification: invalid priority range")
	}
	query, err := db.Prepare(`
	INSERT INTO
	notifications(
		Id,
		Username,
		Priority,
		Name,
		Description,
		Date
	)
	VALUES (DEFAULT, ?, ?, ?, ?, DEFAULT)
	`)
	if err != nil {
		log.Error("Failed to add notification: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(receiverUsername, priority, name, description)
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
	defer query.Close()
	if _, err = query.Exec(username); err != nil {
		log.Error("Failed to remove all user notifications: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes a given notification, can return an error
func DeleteNotificationFromUserById(notificationId uint, username string) error {
	query, err := db.Prepare(`
	DELETE FROM
	notifications
	WHERE Id=?
	AND Username=?
	`)
	if err != nil {
		log.Error("Failed to delete notification by id: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(notificationId, username)
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
	SELECT
		COUNT(Id)
		AS Count
	FROM notifications
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to get notification count: preparing query failed: ", err.Error())
		return 0, err
	}
	defer query.Close()
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
	SELECT
		Id,
		Priority,
		Name,
		Description,
		Date
	FROM notifications
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to get notifications: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to get notifications: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	notifications := make([]Notification, 0)
	for res.Next() {
		var notificationItem Notification
		var notificationTime sql.NullTime
		err := res.Scan(
			&notificationItem.Id,
			&notificationItem.Priority,
			&notificationItem.Name,
			&notificationItem.Description,
			&notificationTime,
		)
		if err != nil {
			log.Error()
			return nil, err
		}
		if !notificationTime.Valid {
			log.Error("Failed to get user notifications: notification time is not valid: critical failure")
			return nil, errors.New("failed to get user notifications: notification date column contains null value")
		}
		notificationItem.Date = notificationTime.Time
		notifications = append(notifications, notificationItem)
	}
	return notifications, nil
}
