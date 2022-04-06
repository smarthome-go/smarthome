package database

import "time"

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

func GetUserReminders(username string) (Reminder, bool, error) {
	return Reminder{}, false, nil
}
