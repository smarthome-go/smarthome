package database

import "testing"

func TestCreateGetConfig(t *testing.T) {
	// 	Delete possible table to be sure it does not exists
	if _, err := db.Exec(`DROP TABLE IF EXISTS configuration`); err != nil {
		t.Error(err.Error())
		return
	}
	if err := createConfigTable(); err != nil {
		t.Error(err.Error())
		return
	}
	// Query for config after its creation
	config, exists, err := GetServerConfiguration()
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Error("Configuration does not exists after creation")
		return
	}
	if !config.AutomationEnabled || config.LockDownMode || config.Latitude != 0.0 || config.Longitude != 0.0 {
		t.Errorf("Invalid configuration after creation: got: %v", config)
		return
	}
}

func TestSetConfig(t *testing.T) {
	configNew := ServerConfig{
		AutomationEnabled: false,
		LockDownMode:      true,
		Latitude:          42.42,
		Longitude:         42.42,
	}
	if err := SetServerConfiguration(configNew); err != nil {
		t.Error(err.Error())
		return
	}
	// Query for config after its modification
	config, exists, err := GetServerConfiguration()
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Error("Configuration does not exists after modification")
		return
	}
	if config.AutomationEnabled != config.AutomationEnabled ||
		config.LockDownMode != config.LockDownMode ||
		config.Latitude != config.Latitude ||
		config.Longitude != config.Longitude {
		t.Errorf("Configuration was not modified: want: %v got: %v", configNew, config)
		return
	}
}

func TestSetAutomationSystemStatus(t *testing.T) {
	configBefore, exists, err := GetServerConfiguration()
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Error("Configuration does not exists")
		return
	}
	if err := SetAutomationSystemActivation(!configBefore.AutomationEnabled); err != nil {
		t.Error(err.Error())
		return
	}
	configAfter, exists, err := GetServerConfiguration()
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Error("Configuration does not exists after modification")
		return
	}
	if configAfter.AutomationEnabled == configBefore.AutomationEnabled {
		t.Errorf("Automation status did not change after modification: want: %t got %t", !configBefore.AutomationEnabled, configAfter.AutomationEnabled)
		return
	}
}

func TestUpdateLocation(t *testing.T) {
	var wantLat float32 = 11.11
	var wantLon float32 = 11.11

	if err := UpdateLocation(wantLat, wantLon); err != nil {
		t.Error(err.Error())
		return
	}
	configAfter, exists, err := GetServerConfiguration()
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Error("Configuration does not exists after modification")
		return
	}
	if configAfter.Latitude != wantLat || configAfter.Longitude != wantLon {
		t.Errorf("Geolocation did not change: want: (%f | %f) got: (%f | %f)",
			wantLat, wantLon, configAfter.Latitude, configAfter.Longitude,
		)
		return
	}
}
