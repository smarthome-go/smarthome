package database

import "database/sql"

type ServerConfig struct {
	AutomationEnabled bool `json:"automationEnabled"` // Sets the global state of the server's automation system
	LockDownMode      bool `json:"lockDownMode"`      // If enabled, the server is unable to change power states and will not allow power actions
}

// Creates the table that contains the server configuration
func CreateConfigTable() error {
	_, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	configuration(
		Id INT PRIMARY KEY,
		AutomationEnabled BOOLEAN DEFAULT TRUE,
		LockDownMode BOOLEAN DEFAULT FALSE
	)`)
	if err != nil {
		log.Error("Failed to create server configuration table: executing query failed: ", err.Error())
		return err
	}
	_, found, err := GetServerConfiguration()
	if err != nil {
		log.Error("Failed to create server configuration table: probing for present configuration failed: ", err.Error())
		return err
	}
	if !found {
		if _, err := db.Exec(`
		INSERT INTO
		configuration(Id, AutomationEnabled, LockDownMode)
		VALUES(0, TRUE, FALSE)
		`); err != nil {
			log.Error("Failed to create configuration: insert failed: executing query failed: ", err.Error())
			return err
		}
		log.Trace("Created new server configuration")
	}
	return nil
}

// Retrieves the servers configuration
func GetServerConfiguration() (ServerConfig, bool, error) {
	var config ServerConfig
	err := db.QueryRow(`
	SELECT
	AutomationEnabled, LockDownMode
	FROM configuration
	WHERE Id=0
	`).Scan(
		&config.AutomationEnabled,
		&config.LockDownMode,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Trace("No results for server configuration")
			return ServerConfig{}, false, nil
		}
		log.Error("Failed to retrieve server configuration: ", err.Error())
		return ServerConfig{}, false, err
	}
	return config, true, nil
}

// Updates the servers configuration
func SetServerConfiguration(config ServerConfig) error {
	query, err := db.Prepare(`
	UPDATE configuration
	SET
	AutomationEnabled=?,
	LockDownMode=?
	WHERE Id=0
	`)
	if err != nil {
		log.Error("Failed to update the servers configuration: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(
		config.AutomationEnabled,
		config.LockDownMode,
	); err != nil {
		log.Error("Failed to update the servers configuration: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Change the state of the automation system
func SetAutomationSystemActivation(enabled bool) error {
	query, err := db.Prepare(`
	UPDATE configuration
	SET
	AutomationEnabled=?
	WHERE Id=0
	`)
	if err != nil {
		log.Error("Failed to update the activation mode of automations: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(enabled); err != nil {
		log.Error("Failed to update the activation mode of automations: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// TODO: create server config / at startup
