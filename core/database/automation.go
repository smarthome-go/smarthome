// Contains the database backend for static automation
package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Represents the timing mode on which the automation operates
type TimingMode string

const (
	// Will not change, an automation will always execute based on this time
	TimingNormal TimingMode = "normal"
	// Uses the time of local sunrise
	// => Each run of a set automation will update the underlyingk time
	// => Each run will update and regenerate a cron-expression
	TimingSunrise TimingMode = "sunrise"
	// Same as above, just uses the time of local sunset
	TimingSunset TimingMode = "sunset"
)

type Automation struct {
	// The ID is automatically generated
	Id    uint           `json:"id"`
	Owner string         `json:"owner"`
	Data  AutomationData `json:"data"`
}

type AutomationData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// Saves the underlying cron-expression to wrap the time and days of execution
	CronExpression string `json:"cronExpression"`
	// Specifies which Homescript is to be executed when the automation runs
	HomescriptId string `json:"homescriptId"`
	Enabled      bool   `json:"enabled"`
	// Wont run the automation the next time it would
	DisableOnce bool       `json:"disableOnce"`
	TimingMode  TimingMode `json:"timingMode"`
	LastRun     time.Time  `json:"lastRun"`
}

// Creates a new table containing the automation jobs
func createAutomationTable() error {
	_, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	automation(
		Id INT AUTO_INCREMENT,
		Name VARCHAR(30),
		Description TEXT,
		CronExpression VARCHAR(100),
		HomescriptId VARCHAR(30),
		Owner VARCHAR(20),
		Enabled BOOL,
		DisableOnce BOOL,
		TimingMode ENUM(
			'normal',
			'sunrise',
			'sunset'
		),
		LastRun DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(Id),
		FOREIGN KEY (HomescriptId)
		REFERENCES homescript(Id),
		FOREIGN KEY (Owner)
		REFERENCES user(Username)
	)
	`)
	if err != nil {
		log.Error("Failed to create automation table: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Creates a new automation item using the raw data provided
func CreateNewAutomation(automation Automation) (uint, error) {
	query, err := db.Prepare(`
	INSERT INTO
	automation(
		Id,
		Name,
		Description,
		CronExpression,
		HomescriptId,
		Owner,
		Enabled,
		DisableOnce,
		TimingMode,
		LastRun
	)
	VALUES(DEFAULT, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to create new automation: preparing query failed: ", err.Error())
		return 0, err
	}
	defer query.Close()
	res, err := query.Exec(
		automation.Data.Name,
		automation.Data.Description,
		automation.Data.CronExpression,
		automation.Data.HomescriptId,
		automation.Owner,
		automation.Data.Enabled,
		automation.Data.DisableOnce,
		automation.Data.TimingMode,
		time.Unix(0, 0),
	)
	if err != nil {
		log.Error("Failed to create new automation: executing query failed: ", err.Error())
		return 0, err
	}
	newId, err := res.LastInsertId()
	if err != nil {
		log.Error("Failed to create new automation: retrieving last inserted id failed: ", err.Error())
		return 0, err
	}
	return uint(newId), nil
}

// Returns a Automation struct which matches the given Id
// If the id does not match a struct, an empty struct and `false` (for found) is returned
func GetAutomationById(id uint) (Automation, bool, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		Name,
		Description,
		CronExpression,
		HomescriptId,
		Owner,
		Enabled,
		DisableOnce,
		TimingMode,
		LastRun
	FROM automation
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Could not get automation by id: preparing query failed: ", err.Error())
		return Automation{}, false, err
	}
	defer query.Close()
	var automation Automation
	var lastRun sql.NullTime
	if err := query.QueryRow(id).Scan(
		&automation.Id,
		&automation.Data.Name,
		&automation.Data.Description,
		&automation.Data.CronExpression,
		&automation.Data.HomescriptId,
		&automation.Owner,
		&automation.Data.Enabled,
		&automation.Data.DisableOnce,
		&automation.Data.TimingMode,
		&lastRun,
	); err != nil {
		if err == sql.ErrNoRows {
			return Automation{}, false, nil
		}
		return Automation{}, false, err
	}

	if !lastRun.Valid {
		log.Error("Invalid time column when getting automation")
		return Automation{}, false, fmt.Errorf("invalid time column when scanning logs")
	} else {
		automation.Data.LastRun = lastRun.Time
	}

	return automation, true, nil
}

// Returns a list containing automations of a given user (must be valid)
// An invalid user will yield an empty list
func GetUserAutomations(username string) ([]Automation, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		Name,
		Description,
		CronExpression,
		HomescriptId,
		Owner,
		Enabled,
		DisableOnce,
		TimingMode,
		LastRun
	FROM automation
	WHERE Owner=?
	`)
	if err != nil {
		log.Error("Failed to list user automations: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to list user automations: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	automations := make([]Automation, 0)
	for res.Next() {
		var automation Automation
		var lastRun sql.NullTime
		if err := res.Scan(
			&automation.Id,
			&automation.Data.Name,
			&automation.Data.Description,
			&automation.Data.CronExpression,
			&automation.Data.HomescriptId,
			&automation.Owner,
			&automation.Data.Enabled,
			&automation.Data.DisableOnce,
			&automation.Data.TimingMode,
			&lastRun,
		); err != nil {
			log.Error("Failed to list user automations: scanning for results failed: ", err.Error())
			return nil, err
		}

		if !lastRun.Valid {
			log.Error("Invalid time column when getting automation")
			return nil, fmt.Errorf("invalid time column when scanning logs")
		} else {
			automation.Data.LastRun = lastRun.Time
		}

		automations = append(automations, automation)
	}
	return automations, nil
}

// Returns a slice with automations of all users
// Used for activating persistent automations when the server starts
func GetAutomations() ([]Automation, error) {
	res, err := db.Query(`
	SELECT
		Id,
		Name,
		Description,
		CronExpression,
		HomescriptId,
		Owner,
		Enabled,
		DisableOnce,
		TimingMode,
		LastRun
	FROM automation
	`)
	if err != nil {
		log.Error("Failed to list all automations: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	automations := make([]Automation, 0)
	for res.Next() {
		var automation Automation
		var lastRun sql.NullTime

		if err := res.Scan(
			&automation.Id,
			&automation.Data.Name,
			&automation.Data.Description,
			&automation.Data.CronExpression,
			&automation.Data.HomescriptId,
			&automation.Owner,
			&automation.Data.Enabled,
			&automation.Data.DisableOnce,
			&automation.Data.TimingMode,
			&lastRun,
		); err != nil {
			log.Error("Failed to list all automations: scanning for results failed: ", err.Error())
			return nil, err
		}

		if !lastRun.Valid {
			log.Error("Invalid time column when getting automation")
			return nil, fmt.Errorf("invalid time column when scanning logs")
		} else {
			automation.Data.LastRun = lastRun.Time
		}

		automations = append(automations, automation)
	}
	return automations, nil
}

// Modifies the metadata of a given automation item given its raw id
// Does not validate the provided metadata
// If an invalid id is specified, no data will be modified
func ModifyAutomation(id uint, newItem AutomationData) error {
	query, err := db.Prepare(`
	UPDATE automation
	SET
		Name=?,
		Description=?,
		CronExpression=?,
		HomescriptId=?,
		Enabled=?,
		DisableOnce=?,
		TimingMode=?
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to modify automation: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(
		newItem.Name,
		newItem.Description,
		newItem.CronExpression,
		newItem.HomescriptId,
		newItem.Enabled,
		newItem.DisableOnce,
		newItem.TimingMode,
		id,
	)
	if err != nil {
		log.Error("Failed to modify automation: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Sets the last run timestamp of the given automation to now.
func UpdateAutomationExecuteTime(id uint) error {
	query, err := db.Prepare(`
	UPDATE automation
	SET
		LastRun=?
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to modify automation `lastRun` timestamp: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(
		time.Now(),
		id,
	)
	if err != nil {
		log.Error("Failed to modify automation `lastRun` timestamp: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes an automation item given its Id
// Does not validate the validity of the provided Id
// If an invalid id is specified, nothing will be deleted
func DeleteAutomationById(id uint) error {
	query, err := db.Prepare(`
	DELETE FROM
	automation
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to delete automation by Id: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(id); err != nil {
		log.Error("Failed to delete automation by Id: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes all automations from a given user
// Used when deleting a user
// An invalid username will lead to no deletions
func DeleteAllAutomationsFromUser(username string) error {
	query, err := db.Prepare(`
	DELETE FROM
	automation
	WHERE Owner=?
	`)
	if err != nil {
		log.Error("Failed to delete all automations from user: preparing query failed", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(username); err != nil {
		log.Error("Failed to delete all automations from user: executing query failed", err.Error())
		return err
	}
	return nil
}
