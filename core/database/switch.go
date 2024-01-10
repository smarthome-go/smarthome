package database

import (
	"database/sql"
	"fmt"
)

// Identified by a Switch Id, has a name and belongs to a room
type Switch struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	RoomId   string `json:"roomId"`
	PowerOn  bool   `json:"powerOn"`
	Watts    uint16 `json:"watts"`
	VendorId string `json:"vendorId"`
	ModelId  string `json:"modelId"`
}

// Contains the switch id and a matching boolean which indicates whether the switch is on or off
// Additionally, the power draw is also included (mainly used for taking periodic snapshots of the power statistics)
// Used when requesting global power states
type PowerState struct {
	Switch  string `json:"switch"`
	PowerOn bool   `json:"powerOn"`
	Watts   uint16 `json:"watts"`
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

// Creates the table containing optional target hardware ids
// If the database fails, this function returns an error
func createSwitchTargetNodeTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	switchTargetNode(
		SwitchId VARCHAR(20) PRIMARY KEY,
		NodeUrl VARCHAR(50),
		FOREIGN KEY (SwitchId)
		REFERENCES switch(Id),
		FOREIGN KEY (NodeUrl)
		REFERENCES hardware(Url)
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create switch target node Table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Creates a new switch
// Will return an error if the database fails
func CreateDevice(
	id string,
	name string,
	roomId string,
	watts uint16,
	driverVendorId string,
	driverModelId string,
) error {
	query, err := db.Prepare(`
	INSERT INTO
	switch(
		Id,
		Name,
		Power,
		RoomId,
		Watts,
		DriverVendorId,
		DriverModelId
	)
	VALUES(?, ?, DEFAULT, ?, ?, ?, ?)
	ON DUPLICATE KEY
		UPDATE
		Name=VALUES(Name),
		RoomId=VALUES(RoomId),
		Watts=VALUES(Watts),
		DriverVendorId=VALUES(DriverVendorId),
		DriverModelId=VALUES(DriverModelId)
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

	// TODO: handle drivers

	// create target node entry if required
	// if targetNode != nil {
	// 	if err := setSwitchTargetNode(id, *targetNode); err != nil {
	// 		return err
	// 	}
	// }

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
func ModifySwitch(id string, name string, watts uint16, targetNode *string) error {
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

	defer query.Close()

	if _, err := query.Exec(name, watts, id); err != nil {
		log.Error("Failed to modify switch: executing query failed: ", err.Error())
		return err
	}

	if targetNode == nil {
		return removeSwitchTargetNode(id)
	} else {
		return setSwitchTargetNode(id, *targetNode)
	}
}

func removeSwitchTargetNode(switchId string) error {
	query, err := db.Prepare(`
		DELETE FROM switchTargetNode
			WHERE SwitchId=?
		`)
	if err != nil {
		log.Error("Failed to modify switch target node: preparing deletion query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(switchId); err != nil {
		log.Error("Failed to modify switch target node: executing deletion query failed: ", err.Error())
		return err
	}
	return nil
}

func setSwitchTargetNode(switchId string, nodeUrl string) error {
	query, err := db.Prepare(`
		INSERT INTO switchTargetNode(SwitchId, NodeUrl)
		VALUES(?, ?)
		ON DUPLICATE KEY
			UPDATE NodeUrl=VALUES(NodeUrl)
		`)
	if err != nil {
		log.Error("Failed to modify switch target node: preparing insertion query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(switchId, nodeUrl); err != nil {
		log.Error("Failed to modify switch target node: executing insertion query failed: ", err.Error())
		return err
	}
	return nil
}

// Delete a given switch after all data which depends on this switch has been deleted
func DeleteSwitch(switchId string) error {
	if err := RemoveSwitchFromPermissions(switchId); err != nil {
		return err
	}
	if err := removeSwitchTargetNode(switchId); err != nil {
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
// Before deleting the switch, dependencies like switch-permissions are deleted
func DeleteRoomSwitches(roomId string) error {
	switches, err := ListSwitches()
	if err != nil {
		return err
	}
	for _, switchItem := range switches {
		if switchItem.RoomId != roomId {
			continue
		}
		if err := DeleteSwitch(switchItem.Id); err != nil {
			return err
		}
	}
	return nil
}

// Returns a list of all available switches with their attributes
func ListSwitches() ([]Switch, error) {
	res, err := db.Query(`
	SELECT
		switch.Id,
		switch.Name,
		switch.Power,
		switch.RoomId,
		switch.Watts,
		switchTargetNode.NodeUrl
	FROM switch
	LEFT JOIN switchTargetNode
		ON switchTargetNode.SwitchId = switch.Id
	`)
	if err != nil {
		log.Error("Could not list switches: failed to execute query: ", err.Error())
		return nil, err
	}
	defer res.Close()
	switches := make([]Switch, 0)
	for res.Next() {
		var switchItem Switch
		var switchTargetNode sql.NullString
		if err := res.Scan(
			&switchItem.Id,
			&switchItem.Name,
			&switchItem.PowerOn,
			&switchItem.RoomId,
			&switchItem.Watts,
			&switchTargetNode,
		); err != nil {
			log.Error("Could not list switches: Failed to scan results: ", err.Error())
			return nil, err
		}
		if switchTargetNode.Valid {
			switchItem.TargetNode = &switchTargetNode.String
		}
		switches = append(switches, switchItem)
	}
	return switches, nil
}

// Like `ListSwitches()` but takes a user string as a filter
// Only returns switches which are contained in the switch-permission table with the given user
func ListUserSwitchesQuery(username string) ([]Switch, error) {
	query, err := db.Prepare(`
	SELECT
		switch.Id,
		switch.Name,
		switch.RoomId,
		switch.Power,
		switch.Watts,
		switchTargetNode.NodeUrl
	FROM switch
	JOIN hasSwitchPermission
		ON hasSwitchPermission.Switch=switch.Id
	LEFT JOIN switchTargetNode
		ON switchTargetNode.switchId = switch.Id
	WHERE hasSwitchPermission.Username=?`,
	)
	if err != nil {
		log.Error("Could not list user switches: preparing query failed: ", err.Error())
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
		var switchTargetNode sql.NullString

		if err := res.Scan(
			&switchItem.Id,
			&switchItem.Name,
			&switchItem.RoomId,
			&switchItem.PowerOn,
			&switchItem.Watts,
			&switchTargetNode,
		); err != nil {
			log.Error("Could not list user switches: Failed to scan results: ", err.Error())
			return nil, err
		}

		if switchTargetNode.Valid {
			switchItem.TargetNode = &switchTargetNode.String
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
		switch.Id,
		switch.Name,
		switch.RoomId,
		switch.Power,
		switch.Watts,
		switchTargetNode.NodeUrl
	FROM switch
	LEFT JOIN switchTargetNode
		ON switchTargetNode.SwitchId = switch.Id
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to get switch by id: preparing query failed: ", err.Error())
		return Switch{}, false, err
	}
	var switchItem Switch
	var switchTargetNode sql.NullString
	if err := query.QueryRow(id).Scan(
		&switchItem.Id,
		&switchItem.Name,
		&switchItem.RoomId,
		&switchItem.PowerOn,
		&switchItem.Watts,
		&switchTargetNode,
	); err != nil {
		if err == sql.ErrNoRows {
			return Switch{}, false, nil
		}
		log.Error("Failed to get switch by id: scanning results failed: ", err.Error())
		return Switch{}, false, err
	}

	if switchTargetNode.Valid {
		switchItem.TargetNode = &switchTargetNode.String
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
		Id,
		Power,
		Watts
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
		err := res.Scan(
			&powerState.Switch,
			&powerState.PowerOn,
			&powerState.Watts,
		)
		if err != nil {
			log.Error("Failed to list powerstates: failed to scan query: ", err.Error())
			return nil, err
		}
		powerStates = append(powerStates, powerState)
	}
	return powerStates, nil
}
