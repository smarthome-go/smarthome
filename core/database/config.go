package database

import (
	"database/sql"
)

type ServerConfig struct {
	// Sets the global state of the server's automation system
	AutomationEnabled bool `json:"automationEnabled"`
	// If enabled, the server is unable to change power states and will not allow power actions
	LockDownMode bool `json:"lockDownMode"`
	// Specifies the OpenWeatherMap API key, if left empty, open weather will be disabled
	OpenWeatherMapApiKey string `json:"openWeatherMapApiKey"`
	// Specifies the physical location of the Smarthome server
	Latitude float32 `json:"latitude"`
	// Latitude and longitude are being used for calculating the sunset / sunrise times and for OpenWeatherMap's weather service
	Longitude float32    `json:"longitude"`
	Mqtt      MqttConfig `json:"mqtt"`
}

type MqttConfig struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     uint16 `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Creates the table that contains the server configuration
func createConfigTable() error {
	_, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	configuration(
		Id						INT PRIMARY KEY,
		AutomationEnabled		BOOLEAN DEFAULT TRUE,
		LockDownMode BOOLEAN	DEFAULT FALSE,
		-- Begin Weather
		OpenWeatherMapApiKey	VARCHAR(64) DEFAULT "",
		Latitude FLOAT(32)		DEFAULT 0.0,
		Longitude FLOAT(32)		DEFAULT 0.0,
		-- Begin MQTT
		MQTTEnabled				BOOLEAN,
		MQTTHost				TEXT,
		MQTTPort				SMALLINT,
		MQTTUsername			TEXT,
		MQTTPassword			TEXT
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
			LockDownMode,
			MQTTEnabled,
			MQTTHost,
			MQTTPort,
			MQTTUsername,
			MQTTPassword
		)
		VALUES(0, TRUE, FALSE, FALSE, "host", 1883, "username", "")
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
	if err := db.QueryRow(`
	SELECT
		AutomationEnabled,
		LockDownMode,
		OpenWeatherMapApiKey,
		Latitude,
		Longitude,
		MQTTEnabled,
		MQTTHost,
		MQTTPort,
		MQTTUsername,
		MQTTPassword
	FROM configuration
	WHERE Id=0
	`).Scan(
		&config.AutomationEnabled,
		&config.LockDownMode,
		&config.OpenWeatherMapApiKey,
		&config.Latitude,
		&config.Longitude,
		&config.Mqtt.Enabled,
		&config.Mqtt.Host,
		&config.Mqtt.Port,
		&config.Mqtt.Username,
		&config.Mqtt.Password,
	); err != nil {
		if err == sql.ErrNoRows {
			log.Warn("No server configuration present")
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
		OpenWeatherMapApiKey=?,
		Latitude=?,
		Longitude=?,
		-- MQTT section
		MQTTEnabled=?,
		MQTTHost=?,
		MQTTPort=?,
		MQTTUsername=?,
		MQTTPassword=?
	WHERE Id=0
	`)
	if err != nil {
		log.Error("Failed to update the servers configuration: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(
		config.AutomationEnabled,
		config.LockDownMode,
		config.OpenWeatherMapApiKey,
		config.Latitude,
		config.Longitude,
		config.Mqtt.Enabled,
		config.Mqtt.Host,
		config.Mqtt.Port,
		config.Mqtt.Username,
		config.Mqtt.Password,
	); err != nil {
		log.Error("Failed to update the servers configuration: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Changes the state of the lock-down mode
func SetLockDownModeEnabled(enabled bool) error {
	query, err := db.Prepare(`
	UPDATE configuration
	SET
		LockDownMode=?
	WHERE Id=0
	`)
	if err != nil {
		log.Error("Failed to update lock-down mode: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(enabled); err != nil {
		log.Error("Failed to update lock-down mode: executing query failed: ", err.Error())
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
	defer query.Close()
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
	defer query.Close()
	if _, err := query.Exec(lat, lon); err != nil {
		log.Error("Failed to update the servers location: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Changes the server's Open Weather Map API Key
func UpdateOpenWeatherMapApiKey(newKey string) error {
	query, err := db.Prepare(`
	UPDATE configuration
	SET
		OpenWeatherMapApiKey=?
	WHERE Id=0
	`)
	if err != nil {
		log.Error("Failed to update the servers OpenWeatherMap API Key: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(newKey); err != nil {
		log.Error("Failed to update the servers OpenWeatherMap API Key: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Changes the server's MQTT broker attributes
func UpdateMqttConfig(settings MqttConfig) error {
	query, err := db.Prepare(`
	UPDATE configuration
	SET
		MQTTEnabled=?,
		MQTTHost=?,
		MQTTPort=?,
		MQTTUsername=?,
		MQTTPassword=?
	WHERE Id=0
	`)
	if err != nil {
		log.Error("Failed to update the servers MQTT settings: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(
		settings.Enabled,
		settings.Host,
		settings.Port,
		settings.Username,
		settings.Password,
	); err != nil {
		log.Error("Failed to update the servers MQTT config: executing query failed: ", err.Error())
		return err
	}
	return nil
}
