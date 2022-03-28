package database

import (
	"fmt"
)

// Identified by a Switch Id, has a name and belongs to a room
type Switch struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	RoomId  string `json:"roomId"`
	PowerOn bool   `json:"powerOn"`
	Watts   uint16 `json:"watts"`
}

//Contains the switch id and a matching boolean
// Used when requesting global power states
type PowerState struct {
	Switch  string `json:"switch"`
	PowerOn bool   `json:"powerOn"`
}

// Creates the table containing switches
// If the database fails, this function returns an error
func createSwitchTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	switch(
		Id VARCHAR(2) PRIMARY KEY,
		Name VARCHAR(30),
		Power BOOLEAN DEFAULT FALSE,
		RoomId VARCHAR(30),
		Watts INT,
		CONSTRAINT SwitchRoomId
		FOREIGN KEY (RoomId)
		REFERENCES room(Id)
	) 
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create switch Table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Stores the n:m relation between the user and their switch-permissions
func createHasSwitchPermissionTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	hasSwitchPermission(
		Username VARCHAR(20),
		Switch VARCHAR(2),
		CONSTRAINT HasSwitchPermissionUsername
		FOREIGN KEY (Username)
		REFERENCES user(Username),
		CONSTRAINT HasSwitchPermissionSwitch
		FOREIGN KEY (Switch)
		REFERENCES switch(Id)
	)
	`
	_, err := db.Query(query)
	if err != nil {
		log.Error("Failed to create hasSwitchPermissionTable: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Creates a new switch
// Will return an error if the database fails
func CreateSwitch(id string, name string, roomId string, watts uint16) error {
	query, err := db.Prepare(`
	INSERT INTO
	switch(
		Id, Name, Power, RoomId, Watts
	)
	VALUES(?, ?, DEFAULT, ?, ?)
	ON DUPLICATE KEY
	UPDATE
	Name=VALUES(Name),
	Power=VALUES(Power),
	RoomId=VALUES(RoomId),
	Watts=VALUES(Watts)
	`)
	if err != nil {
		log.Error("Failed to add switch: preparing query failed: ", err.Error())
		return err
	}
	res, err := query.Exec(id, name, roomId, watts)
	if err != nil {
		log.Error("Failed to add switch: executing query failed: ", err.Error())
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("Could not get result of createSwitch: obtaining rowsAffected failed: ", err.Error())
		return err
	}
	if rowsAffected > 0 {
		log.Debug(fmt.Sprintf("Added switch `%s` with name `%s`", id, name))
	}
	defer query.Close()
	return nil
}

// Delete a given switch after all data which depends on this switch has been deleted
func DeleteSwitch(switchId string) error {
	if err := RemoveSwitchFromPermissions(switchId); err != nil {
		log.Error("Failed to remove switch: dependencies could not be removed: ", err.Error())
	}
	query, err := db.Prepare(`
	DELETE FROM
	switch
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to remove switch: preparing query failed: ", err.Error())
		return err
	}
	if _, err = query.Exec(switchId); err != nil {
		log.Error("Failed to remove switch: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns a list of available switches with their attributes
func ListSwitches() ([]Switch, error) {
	query := `
	SELECT Id, Name, RoomId, Watts FROM switch
	`
	res, err := db.Query(query)
	if err != nil {
		log.Error("Could not list switches: failed to execute query: ", err.Error())
		return []Switch{}, err
	}
	switches := make([]Switch, 0)
	for res.Next() {
		var switchItem Switch
		if err := res.Scan(
			&switchItem.Id,
			&switchItem.Name,
			&switchItem.RoomId,
			&switchItem.Watts,
		); err != nil {
			log.Error("Could not list switches: Failed to scan results: ", err.Error())
		}
		switches = append(switches, switchItem)
	}
	return switches, nil
}

// Same as `ListSwitches()` but takes a user sting as a filter
func ListUserSwitches(username string) ([]Switch, error) {
	query, err := db.Prepare(`
	SELECT Id, Name, RoomId, Power, Watts
	FROM switch
	JOIN hasSwitchPermission
	ON hasSwitchPermission.Switch=switch.Id
	WHERE hasSwitchPermission.Username=?`,
	)
	if err != nil {
		log.Error("Could not list user switches: preparing query failed.", err.Error())
		return []Switch{}, err
	}
	res, err := query.Query(username)
	if err != nil {
		log.Error("Could not list user switches: executing query failed: ", err.Error())
		return []Switch{}, err
	}
	switches := make([]Switch, 0)
	for res.Next() {
		var switchItem Switch
		if err := res.Scan(
			&switchItem.Id,
			&switchItem.Name,
			&switchItem.RoomId,
			&switchItem.PowerOn,
			&switchItem.Watts,
		); err != nil {
			log.Error("Could not list user switches: Failed to scan results: ", err.Error())
		}
		switches = append(switches, switchItem)
	}
	defer query.Close()
	return switches, nil
}

// Used when marking a power state of a switch
// Does not check the validity of the switch Id
// The returned boolean indicates if the power state had changed
func SetPowerState(switchId string, isPoweredOn bool) (bool, error) {
	query, err := db.Prepare(`
	UPDATE switch
	SET Power=?
	WHERE Id=? 
	`)
	if err != nil {
		log.Error("Could not alter power state: preparing query failed: ", err.Error())
		return false, err
	}
	res, err := query.Exec(isPoweredOn, switchId)
	if err != nil {
		log.Error("Could not alter power state: executing query failed: ", err.Error())
		return false, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("Could not evaluate outcome of `SetPowerState`: Reading RowsAffected failed: ", err.Error())
		return false, err
	}
	if rowsAffected == 0 {
		return false, nil
	}
	defer query.Close()
	return true, nil
}

// Returns a list of PowerStates
// Can return a database error
func GetPowerStates() ([]PowerState, error) {
	res, err := db.Query(`
	SELECT 
	Id, Power
	FROM switch
	`)
	if err != nil {
		log.Error("Failed to list powerstates: failed to execute query: ", err.Error())
	}
	powerStates := make([]PowerState, 0)
	for res.Next() {
		var powerState PowerState
		err := res.Scan(&powerState.Switch, &powerState.PowerOn)
		if err != nil {
			log.Error("Failed to list powerstates: failed to scan query: ", err.Error())
			return []PowerState{}, err
		}
		powerStates = append(powerStates, powerState)
	}
	return powerStates, nil
}

// Returns the power state of a given switch as a boolean
// Does not check if the switch exists.
// If the switch does not exist, an error is returned
func GetPowerStateOfSwitch(switchId string) (bool, error) {
	query, err := db.Prepare(`
	SELECT Power
	FROM switch
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to get switch power state: preparing query failed: ", err.Error())
		return false, err
	}
	var powerState bool
	err = query.QueryRow(switchId).Scan(&powerState)
	if err != nil {
		log.Error("Failed to get switch power state: executing query failed: ", err.Error())
	}
	return powerState, err
}

// Returns (exists, error), err when the database fails
func DoesSwitchExist(switchId string) (bool, error) {
	switches, err := ListSwitches()
	if err != nil {
		log.Error("Cold not validate existence of switch: fatabase failure: ", err.Error())
		return false, err
	}
	for _, switchItem := range switches {
		if switchItem.Id == switchId {
			return true, nil
		}
	}
	return false, nil
}
