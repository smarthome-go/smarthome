package database

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
		AutomationEnabled BOOLEAN DEFAULT TRUE,
		LockDownMode BOOLEAN DEFAULT FALSE
	)`)
	if err != nil {
		log.Error("Failed to create server configuration table: executing query failed: ", err.Error())
		return err
	}
	if _, err := db.Exec(`
	INSERT INTO
	configuration(AutomationEnabled, LockDownMode)
	VALUES(TRUE, FALSE)
	`); err != nil {
		log.Error("Failed to create configuration: insert failed: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Retrieves the servers configuration
func GetServerConfiguration() (ServerConfig, error) {
	var config ServerConfig
	err := db.QueryRow(`
	SELECT
	AutomationEnabled, LockDownMode
	FROM configuration
	`).Scan(
		&config.AutomationEnabled,
		&config.AutomationEnabled,
	)
	if err != nil {
		log.Error("Failed to retrieve server configuration: ", err.Error())
		return ServerConfig{}, err
	}
	return config, nil
}

// Updates the servers configuration
func SetServerConfiguration(config ServerConfig) error {
	query, err := db.Prepare(`
	UPDATE configuration
	SET
	AutomationEnabled=?,
	LockDownMode=?
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

// TODO: create server config / at startup
