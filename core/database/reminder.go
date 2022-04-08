package database

import (
	"database/sql"
	"fmt"
	"time"
)

type NotificationPriority uint

const (
	Low NotificationPriority = iota
	Normal
	Medium
	High
	Urgent
)

type Reminder struct {
	Id                uint                 `json:"id"`
	Name              string               `json:"name"`
	Description       string               `json:"description"`
	Priority          NotificationPriority `json:"priority"`
	CreatedDate       time.Time            `json:"createdDate"`
	DueDate           time.Time            `json:"dueDate"`
	Owner             string               `json:"owner"`
	UserWasNotified   bool                 `json:"userWasNotified"` // Saves if the ownere has been notified about the current urgency of the task
	UserWasNotifiedAt time.Time            `json:"userWasNotifiedAt"`
}

// Creates the table which contains reminders
func createReminderTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	reminder(
		Id INT AUTO_INCREMENT,
		Name TEXT,
		Description TEXT,
		Priority INT,
		CreatedDate DATETIME DEFAULT CURRENT_TIMESTAMP,
		DueDate DATETIME,
		Owner VARCHAR(20),
		UserWasNotified BOOL,
		UserWasNotifiedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(Id),
		FOREIGN KEY(Owner)
		REFERENCES user(Username)
	)
	`
	if _, err := db.Exec(query); err != nil {
		log.Error("Failed to create reminder table: preparing query failed: ", err)
		return err
	}
	return nil
}

// Creates a new reminder in the database
func CreateNewReminder(name string, description string, dueDate time.Time, owner string, priority NotificationPriority) (uint, error) {
	query, err := db.Prepare(`
	INSERT INTO
	reminder(
		Id, Name, Description, CreatedDate, DueDate, Owner, Priority, UserWasNotified, UserWasNotifiedAt
	)
	VALUES(DEFAULT, ?, ?, DEFAULT, ?, ?, ?, FALSE, ?)
	`)
	if err != nil {
		log.Error("Failed to create new reminder: ", err.Error())
		return 0, err
	}
	defer query.Close()
	res, err := query.Exec(name, description, dueDate, owner, priority, time.Date(1970, 0, 0, 0, 0, 0, 0, time.Local))
	if err != nil {
		log.Error("Failed to create new reminder: ", err.Error())
		return 0, err
	}
	newId, err := res.LastInsertId()
	if err != nil {
		log.Error("Failed to create reminder: could not retrieve its id: ", err.Error())
		return 0, err
	}
	return uint(newId), nil
}

// Returns a slice of reminders which were set up by a given user
func GetUserReminders(username string) ([]Reminder, error) {
	query, err := db.Prepare(`
	SELECT
	Id, Name, Description, Priority, CreatedDate, DueDate, Owner, UserWasNotified, UserWasNotifiedAt
	FROM reminder
	WHERE
	Owner=?
	`)
	if err != nil {
		log.Error("Failed to get user reminders: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to get user reminders: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	reminders := make([]Reminder, 0)
	for res.Next() {
		var reminder Reminder

		var createdDate sql.NullTime
		var dueDate sql.NullTime
		var userWasNotifiedAt sql.NullTime

		if err := res.Scan(
			&reminder.Id,
			&reminder.Name,
			&reminder.Description,
			&reminder.Priority,
			&createdDate,
			&dueDate,
			&reminder.Owner,
			&reminder.UserWasNotified,
			&userWasNotifiedAt,
		); err != nil {
			log.Error("Failed to get user reminders: scanning result failed: ", err.Error())
			return nil, err
		}

		if !createdDate.Valid || !dueDate.Valid || !userWasNotifiedAt.Valid {
			log.Error("Failed to get user reminders: some dates are invalid")
			return nil, fmt.Errorf("failed to get user reminders: invalid dates in result")
		}

		reminder.CreatedDate = createdDate.Time
		reminder.DueDate = dueDate.Time
		reminder.UserWasNotifiedAt = userWasNotifiedAt.Time

		reminders = append(reminders, reminder)
	}
	return reminders, nil
}

// Modifies a given reminder to possess the new metadata
func ModifyReminder(id uint, name string, description string, dueDate time.Time, priority NotificationPriority) error {
	query, err := db.Prepare(`
	UPDATE reminder
	SET
	Name=?,
	Description=?,
	DueDate=?,
	Priority=?
	WHERE
	Id=?
	`)
	if err != nil {
		log.Error("Failed to modify reminder: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(
		name,
		description,
		dueDate,
		priority,
		id,
	); err != nil {
		log.Error("Failed to modify reminder: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Modifies the reminders status its owner has been informed about urgency
func SetReminderUserWasNotified(id uint, wasNotified bool, wasNotifiedAt time.Time) error {
	query, err := db.Prepare(`
	UPDATE reminder
	SET
	UserWasNotified=?,
	UserWasNotifiedAt=?
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to update notification status of reminders: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(wasNotified, wasNotifiedAt, id); err != nil {
		log.Error("Failed to update notification status of reminders: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Given its id and owner, the function returns a reminder, if it was found and an error
func GetReminderById(id uint, owner string) (Reminder, bool, error) {
	query, err := db.Prepare(`
	SELECT
	Id, Name, Description, CreatedDate, DueDate, Priority, Owner, UserWasNotified, UserWasNotifiedAt
	FROM reminder
	WHERE
	Id=? AND Owner=?
	`)
	if err != nil {
		log.Error("Failed to check if reminder exists: preparing query failed: ", err.Error())
		return Reminder{}, false, err
	}
	defer query.Close()

	var reminder Reminder
	var createdDate sql.NullTime
	var dueDate sql.NullTime
	var userWasNotifiedAt sql.NullTime

	if err := query.QueryRow(id, owner).Scan(
		&reminder.Id,
		&reminder.Name,
		&reminder.Description,
		&createdDate,
		&dueDate,
		&reminder.Priority,
		&reminder.Owner,
		&reminder.UserWasNotified,
		&userWasNotifiedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return Reminder{}, false, nil
		}
		log.Error("Failed to check if reminder exists: executing query failed: ", err.Error())
		return Reminder{}, false, err

	}

	if !createdDate.Valid || !dueDate.Valid || !userWasNotifiedAt.Valid {
		log.Errorf("Failed to get user reminders: some dates are invalid: (%t, %t, %t)", createdDate.Valid, dueDate.Valid, userWasNotifiedAt.Valid)
		return Reminder{}, false, fmt.Errorf("failed to get user reminders: invalid dates in result")
	}

	reminder.CreatedDate = createdDate.Time
	reminder.DueDate = dueDate.Time
	reminder.UserWasNotifiedAt = userWasNotifiedAt.Time
	return reminder, true, nil
}

// Deletes all reminders of a given user
func DeleteAllRemindersFromUser(username string) error {
	query, err := db.Prepare(`
	DELETE FROM
	reminder
	WHERE
	Owner=?
	`)
	if err != nil {
		log.Error("Failed to remove all reminders from user: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(username); err != nil {
		log.Error("Failed to remove all")
	}
	return nil
}

// Delete a single reminder, for example if its task is finished
func DeleteUserReminderById(owner string, id uint) error {
	query, err := db.Prepare(`
	DELETE FROM
	reminder
	WHERE
	Owner=? AND Id=?
	`)
	if err != nil {
		log.Error("Deleting user reminder failed: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(owner, id); err != nil {
		log.Error("Deleting user reminder failed: executing query failed: ", err.Error())
		return err
	}
	return nil
}
