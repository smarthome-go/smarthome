package database

import (
	"database/sql"
)

type ServerConfig struct {
	AutomationEnabled bool    `json:"automationEnabled"` // Sets the global state of the server's automation system
	LockDownMode      bool    `json:"lockDownMode"`      // If enabled, the server is unable to change power states and will not allow power actions
	Latitude          float32 `json:"latitude"`          // Used for calculating the sunset / sunrise and for openweathermap
	Longitude         float32 `json:"longitude"`
}

// Creates the table that contains the server configuration
func createConfigTable() error {
	_, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	configuration(
		Id INT PRIMARY KEY,
		AutomationEnabled BOOLEAN DEFAULT TRUE,
		LockDownMode BOOLEAN DEFAULT FALSE,
		Latitude FLOAT(32) DEFAULT 0.0,
		Longitude FLOAT(32) DeFAULT 0.0
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
		configuration(
			Id,
			AutomationEnabled,
			LockDownMode
		)
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
		AutomationEnabled,
		LockDownMode,
		Latitude,
		Longitude
	FROM configuration
	WHERE Id=0
	`).Scan(
		&config.AutomationEnabled,
		&config.LockDownMode,
		&config.Latitude,
		&config.Longitude,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Trace("No server configuration present")
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
		LockDownMode=?,
		Latitude=?,
		Longitude=?
	WHERE Id=0
	`)
	if err != nil {
		log.Error("Failed to update the servers configuration: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(
		config.AutomationEnabled,
		config.LockDownMode,
		config.Latitude,
		config.Longitude,
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

// Changes the location of the server
func UpdateLocation(lat float32, lon float32) error {
	query, err := db.Prepare(`
	UPDATE configuration
	SET
		Latitude=?,
		Longitude=?
	WHERE Id=0
	`)
	if err != nil {
		log.Error("Failed to update the servers location: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(lat, lon); err != nil {
		log.Error("Failed to update the servers location: executing query failed: ", err.Error())
		return err
	}
	return nil
}
