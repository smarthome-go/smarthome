package database

import (
	"database/sql"
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
		Id VARCHAR(20) PRIMARY KEY,
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
		RoomId=VALUES(RoomId),
		Watts=VALUES(Watts)
	`)
	if err != nil {
		log.Error("Failed to add switch: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
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
	return nil
}

// Modifies the metadata of a given switch
func ModifySwitch(id string, name string, watts uint16) error {
	query, err := db.Prepare(`
	UPDATE switch
	SET
		Name=?,
		Watts=?
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to modify switch: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(name, watts, id); err != nil {
		log.Error("Failed to modify switch: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Delete a given switch after all data which depends on this switch has been deleted
func DeleteSwitch(switchId string) error {
	if err := RemoveSwitchFromPermissions(switchId); err != nil {
		log.Error("Failed to remove switch: dependencies could not be removed: ", err.Error())
		return err
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
	defer query.Close()
	if _, err = query.Exec(switchId); err != nil {
		log.Error("Failed to remove switch: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes all switches from an arbitrary room
// TODO: move to business layer
func DeleteRoomSwitches(roomId string) error {
	switches, err := ListSwitches()
	if err != nil {
		return err
	}
	for _, switchItem := range switches {
		if err := DeleteSwitch(switchItem.Id); err != nil {
			return err
		}
	}
	return nil
}

// Returns a list of available switches with their attributes
func ListSwitches() ([]Switch, error) {
	res, err := db.Query(`SELECT Id, Name, RoomId, Watts FROM switch`)
	if err != nil {
		log.Error("Could not list switches: failed to execute query: ", err.Error())
		return nil, err
	}
	defer res.Close()
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
			return nil, err
		}
		switches = append(switches, switchItem)
	}
	return switches, nil
}

// Same as `ListSwitches()` but takes a user sting as a filter
func ListUserSwitchesQuery(username string) ([]Switch, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		Name,
		RoomId,
		Power,
		Watts
	FROM switch
	JOIN hasSwitchPermission
	ON hasSwitchPermission.Switch=switch.Id
	WHERE hasSwitchPermission.Username=?`,
	)
	if err != nil {
		log.Error("Could not list user switches: preparing query failed.", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Could not list user switches: executing query failed: ", err.Error())
		return nil, err
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
			return nil, err
		}
		switches = append(switches, switchItem)
	}
	return switches, nil
}

func ListUserSwitches(username string) ([]Switch, error) {
	hasPermissionToAllSwitches, err := UserHasPermission(username, PermissionModifyRooms)
	if err != nil {
		return nil, err
	}
	if hasPermissionToAllSwitches {
		return ListSwitches()
	}
	return ListUserSwitchesQuery(username)
}

// Returns an arbitrary switch given its id
func GetSwitchById(id string) (Switch, bool, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		Name,
		RoomId,
		Power,
		Watts
	FROM switch
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to get switch by id: preparing query failed: ", err.Error())
		return Switch{}, false, err
	}
	var switchItem Switch
	if err := query.QueryRow(id).Scan(
		&switchItem.Id,
		&switchItem.Name,
		&switchItem.RoomId,
		&switchItem.PowerOn,
		&switchItem.Watts,
	); err != nil {
		if err == sql.ErrNoRows {
			return Switch{}, false, nil
		}
		log.Error("Failed to get switch by id: scanning results failed: ", err.Error())
		return Switch{}, false, err
	}
	return switchItem, true, nil
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
	defer query.Close()
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
		return nil, err
	}
	defer res.Close()
	powerStates := make([]PowerState, 0)
	for res.Next() {
		var powerState PowerState
		err := res.Scan(&powerState.Switch, &powerState.PowerOn)
		if err != nil {
			log.Error("Failed to list powerstates: failed to scan query: ", err.Error())
			return nil, err
		}
		powerStates = append(powerStates, powerState)
	}
	return powerStates, nil
}
