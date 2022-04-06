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
	Id          uint                 `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Priority    NotificationPriority `json:"priority"`
	CreatedDate time.Time            `json:"createdDate"`
	DueDate     time.Time            `json:"dueDate"`
	Owner       string               `json:"owner"`
}

// Creates the table which contains reminders
func createReminderTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	reminder(
		Id INT AUTO_INCREMENT,
		Name VARCHAR(200),
		Description TEXT,
		Priority INT,
		CreatedDate DATETIME DEFAULT CURRENT_TIMESTAMP,
		DueDate DATETIME,
		Owner VARCHAR(20),
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
		Id, Name, Description, CreatedDate, DueDate, Owner, Priority
	)
	VALUES(DEFAULT, ?, ?, DEFAULT, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to create new reminder: ", err.Error())
		return 0, err
	}
	res, err := query.Exec(name, description, dueDate, owner, priority)
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
	Id, Name, Description, Priority, CreatedDate, DueDate, Owner
	FROM reminder
	WHERE Owner=?
	`)
	if err != nil {
		log.Error("Failed to get user reminders: preparing query failed: ", err.Error())
		return nil, err
	}
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to get user reminders: executing query failed: ", err.Error())
		return nil, err
	}
	reminders := make([]Reminder, 0)
	for res.Next() {
		var reminder Reminder
		var createdDate sql.NullTime
		var dueDate sql.NullTime
		if err := res.Scan(
			&reminder.Id,
			&reminder.Name,
			&reminder.Description,
			&reminder.Priority,
			&createdDate,
			&dueDate,
			&reminder.Owner,
		); err != nil {
			log.Error("Failed to get user reminders: scanning result failed: ", err.Error())
			return nil, err
		}
		if !createdDate.Valid || !dueDate.Valid {
			log.Error("Failed to get user reminders: some dates are invalid")
			return nil, fmt.Errorf("failed to get user reminders: invalid dates in result")
		}
		reminder.CreatedDate = createdDate.Time
		reminder.DueDate = dueDate.Time
		reminders = append(reminders, reminder)
	}
	return reminders, nil
}

func ModifyReminder(id uint, newName string, newDescription string, newDueDate time.Time) error {
	return nil
}

// Checks if a given reminder exists and is owned by a certain user
func DoesReminderExist(id uint, owner string) (bool, error) {
	query, err := db.Prepare(`
	SELECT
	Id
	FROM reminder
	WHERE Id=?
	AND Owner=?
	`)
	if err != nil {
		log.Error("Failed to check if reminder exists: preparing query failed: ", err.Error())
		return false, err
	}
	// Scan to `id` due to the id not being required
	if err := query.QueryRow(id, owner).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Error("Failed to check if reminder exists: executing query failed: ", err.Error())
		return false, err

	}
	return true, nil
}

// Deletes all reminders of a given user
func DeleteAllRemindersFromUser(username string) error {
	query, err := db.Prepare(`
	DELETE FROM
	reminder
	WHERE Owner=?
	`)
	if err != nil {
		log.Error("Failed to remove all reminders from user: preparing query failed: ", err.Error())
		return err
	}
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
	WHERE Owner=?
	AND Id=?
	`)
	if err != nil {
		log.Error("Deleting user reminder failed: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(owner, id); err != nil {
		log.Error("Deleting user reminder failed: executing query failed: ", err.Error())
		return err
	}
	return nil
}
