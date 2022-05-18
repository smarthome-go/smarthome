package database

import "database/sql"

// Contains the database backend for static automation

type TimingMode string

const (
	TimingNormal  TimingMode = "normal"  // Will not change, automation will always execute based on this time
	TimingSunrise TimingMode = "sunrise" // Uses the local time for sunrise, each run of a set automation will update the actual time and regenerate a cron expression
	TimingSunset  TimingMode = "sunset"  // Same as above, just for sunset
)

type Automation struct {
	Id    uint           `json:"id"`
	Owner string         `json:"owner"`
	Data  AutomationData `json:"data"`
}

type AutomationData struct {
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	CronExpression string     `json:"cronExpression"`
	HomescriptId   string     `json:"homescriptId"`
	Enabled        bool       `json:"enabled"`
	TimingMode     TimingMode `json:"timingMode"`
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
		TimingMode ENUM('normal', 'sunrise', 'sunset'),
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

// Creates a new automation item, does not check the validity of the user or the homescript Id
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
		TimingMode
	)	
	VALUES(DEFAULT, ?, ?, ?, ?, ?, ?, ?)
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
		automation.Data.TimingMode,
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
// If the id does not match a struct, a `false`` is returned
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
		TimingMode
	FROM automation
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Could not get automation by id: preparing query failed: ", err.Error())
		return Automation{}, false, err
	}
	defer query.Close()
	var automation Automation
	err = query.QueryRow(id).Scan(
		&automation.Id,
		&automation.Data.Name,
		&automation.Data.Description,
		&automation.Data.CronExpression,
		&automation.Data.HomescriptId,
		&automation.Owner,
		&automation.Data.Enabled,
		&automation.Data.TimingMode,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Automation{}, false, nil
		}
		return Automation{}, false, err
	}
	return automation, true, nil
}

// Returns a list containing automations of a given user
// Does not check the validity of the user
func GetUserAutomations(username string) ([]Automation, error) {
	query, err := db.Prepare(`
	SELECT
	Id, Name, Description, CronExpression, HomescriptId, Owner, Enabled, TimingMode
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
		if err := res.Scan(
			&automation.Id,
			&automation.Data.Name,
			&automation.Data.Description,
			&automation.Data.CronExpression,
			&automation.Data.HomescriptId,
			&automation.Owner,
			&automation.Data.Enabled,
			&automation.Data.TimingMode,
		); err != nil {
			log.Error("Failed to list user automations: scanning for results failed: ", err.Error())
			return nil, err
		}
		automations = append(automations, automation)
	}
	return automations, nil
}

// Returns a list with automations of all users
// Used for activating persistent automations when the server starts
func GetAutomations() ([]Automation, error) {
	res, err := db.Query(`
	SELECT
	Id, Name, Description, CronExpression, HomescriptId, Owner, Enabled, TimingMode
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
		if err := res.Scan(
			&automation.Id,
			&automation.Data.Name,
			&automation.Data.Description,
			&automation.Data.CronExpression,
			&automation.Data.HomescriptId,
			&automation.Owner,
			&automation.Data.Enabled,
			&automation.Data.TimingMode,
		); err != nil {
			log.Error("Failed to list all automations: scanning for results failed: ", err.Error())
			return nil, err
		}
		automations = append(automations, automation)
	}
	return automations, nil
}

// Modifies the metadata of a given automation item
// Does not validate the provided metadata
func ModifyAutomation(id uint, newItem AutomationData) error {
	query, err := db.Prepare(`
	UPDATE automation
	SET
	Name=?,
	Description=?,
	CronExpression=?,
	HomescriptId=?,
	Enabled=?,
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
		newItem.TimingMode,
		id,
	)
	if err != nil {
		log.Error("Failed to modify automation: executing query failed: ", err.Error())
		return err

	}
	return nil
}

// Deletes an automation item given its Id
// Does not validate the validity of the provided Id
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
