package database

import "database/sql"

// Contains the database backend for static automation

type Automation struct {
	Id             uint   `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	CronExpression string `json:"cronExpression"`
	HomescriptId   string `json:"homescriptId"`
	Owner          string `json:"owner"`
	Enabled        bool   `json:"enabled"`
}

type AutomationWithoutIdAndUsername struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	CronExpression string `json:"cronExpression"`
	HomescriptId   string `json:"homescriptId"`
	Enabled        bool   `json:"enabled"`
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
		Id, Name, Description, CronExpression, HomescriptId, Owner, Enabled
	)	
	VALUES(DEFAULT, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to create new automation: preparing query failed: ", err.Error())
	}
	res, err := query.Exec(
		automation.Name,
		automation.Description,
		automation.CronExpression,
		automation.HomescriptId,
		automation.Owner,
		automation.Enabled,
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
	Id, Name, Description, CronExpression, HomescriptId, Owner, Enabled
	FROM automation
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Could not get automation by id: preparing query failed: ", err.Error())
		return Automation{}, false, err
	}
	var automation Automation
	err = query.QueryRow(id).Scan(
		&automation.Id,
		&automation.Name,
		&automation.Description,
		&automation.CronExpression,
		&automation.HomescriptId,
		&automation.Owner,
		&automation.Enabled,
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
	Id, Name, Description, CronExpression, HomescriptId, Owner, Enabled
	FROM automation
	WHERE Owner=?
	`)
	if err != nil {
		log.Error("Failed to list user automations: preparing query failed: ", err.Error())
		return nil, err
	}
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to list user automations: executing query failed: ", err.Error())
		return nil, err
	}
	automations := make([]Automation, 0)
	for res.Next() {
		var automation Automation
		if err := res.Scan(
			&automation.Id,
			&automation.Name,
			&automation.Description,
			&automation.CronExpression,
			&automation.HomescriptId,
			&automation.Owner,
			&automation.Enabled,
		); err != nil {
			log.Error("Failed to list user automations: scanning for results failed: ", err.Error())
		}
		automations = append(automations, automation)
	}
	defer res.Close()
	return automations, nil
}

// Returns a list with automations of all users
// Used for activating persistent automations when the server starts
func GetAutomations() ([]Automation, error) {
	res, err := db.Query(`
	SELECT
	Id, Name, Description, CronExpression, HomescriptId, Owner, Enabled
	FROM automation
	`)
	if err != nil {
		log.Error("Failed to list all automations: executing query failed: ", err.Error())
		return nil, err
	}
	automations := make([]Automation, 0)
	for res.Next() {
		var automation Automation
		if err := res.Scan(
			&automation.Id,
			&automation.Name,
			&automation.Description,
			&automation.CronExpression,
			&automation.HomescriptId,
			&automation.Owner,
			&automation.Enabled,
		); err != nil {
			log.Error("Failed to list all automations: scanning for results failed: ", err.Error())
		}
		automations = append(automations, automation)
	}
	defer res.Close()
	return automations, nil
}

// Modifies the metadata of a given automation item
// Does not validate the provided metadata
func ModifyAutomation(id uint, newItem AutomationWithoutIdAndUsername) error {
	query, err := db.Prepare(`
	UPDATE automation
	SET
	Name=?,
	Description=?,
	CronExpression=?,
	HomescriptId=?,
	Enabled=?
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to modify automation: preparing query failed: ", err.Error())
		return err
	}
	_, err = query.Exec(
		newItem.Name,
		newItem.Description,
		newItem.CronExpression,
		newItem.HomescriptId,
		newItem.Enabled,
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
	if _, err := query.Exec(id); err != nil {
		log.Error("Failed to delete automation by Id: executing query failed: ", err.Error())
		return err
	}
	return nil
}