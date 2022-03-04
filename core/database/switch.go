package database

import (
	"fmt"
)

// Creates the table containing switches
// If the database fails, this function can return an error
func createSwitchTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS
	switch(
		Id VARCHAR(2) PRIMARY KEY,
		Name VARCHAR(30),
		Power BOOLEAN,
		RoomId VARCHAR(30),
		CONSTRAINT SwitchRoomId FOREIGN KEY (RoomId)
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
	CREATE TABLE IF NOT EXISTS
	hasSwitchPermission(
		Username VARCHAR(20),
		Switch VARCHAR(2),
		CONSTRAINT HasSwitchPermissionUsername FOREIGN KEY (Username)
		REFERENCES user(Username),
		CONSTRAINT HasSwitchPermissionSwitch FOREIGN KEY (Switch)
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
func CreateSwitch(Id string, Name string, RoomId string) error {
	query, err := db.Prepare(`
	INSERT INTO switch(Id, Name, Power, RoomId) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE Name=Values(Name)
	`)
	if err != nil {
		log.Error("Failed to add switch: preparing query failed: ", err.Error())
		return err
	}
	res, err := query.Exec(Id, Name, false, RoomId)
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
		log.Debug(fmt.Sprintf("Added switch `%s` with name `%s`", Id, Name))
	}
	return nil
}

// Returns a list of available switches with their attributes
func ListSwitches() ([]Switch, error) {
	query := `
	SELECT Id, Name, RoomId FROM switch
	`
	res, err := db.Query(query)
	if err != nil {
		log.Error("Could not list switches: failed to execute query: ", err.Error())
		return []Switch{}, err
	}
	switches := make([]Switch, 0)
	for res.Next() {
		var switchItem Switch
		if err := res.Scan(&switchItem.Id, &switchItem.Name, &switchItem.RoomId); err != nil {
			log.Error("Could not list switches: Failed to scan results: ", err.Error())
		}
		switches = append(switches, switchItem)
	}
	return switches, nil
}

// Same as `ListSwitches()` but takes a user sting as a filter
func ListUserSwitches(username string) ([]Switch, error) {
	query, err := db.Prepare(`
	SELECT Id, Name, RoomId FROM switch JOIN hasSwitchPermission ON hasSwitchPermission.Switch=switch.Id WHERE hasSwitchPermission.Username=?`)
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
		if err := res.Scan(&switchItem.Id, &switchItem.Name, &switchItem.RoomId); err != nil {
			log.Error("Could not list user switches: Failed to scan results: ", err.Error())
		}
		switches = append(switches, switchItem)
	}
	return switches, nil
}

// Adds a given switchId to a given user
// If this permission already resides inside the table, it is ignored and `no` error is returned
func AddUserSwitchPermission(username string, switchId string) error {
	userAlreadyHasPermission, err := UserHasSwitchPermission(username, switchId)
	if err != nil {
		log.Error("Failed to add permission: Could not validate the existence of switch: ", err.Error())
		return err
	}
	if userAlreadyHasPermission {
		return nil
	}
	query, err := db.Prepare(`
	INSERT INTO hasSwitchPermission(Username, Switch) VALUES(?,?)
	`)
	if err != nil {
		log.Error("Could not add switch permission to user: preparing query failed: ", err.Error())
		return err
	}
	_, err = query.Exec(username, switchId)
	if err != nil {
		log.Error("Failed to add switch permission to user: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns a list of strings which resemble switch permissions
func GetUserSwitchPermissions(username string) ([]string, error) {
	query, err := db.Prepare(`
	SELECT Switch FROM hasSwitchPermission WHERE Username=?
	`)
	if err != nil {
		log.Error("Could not list user switch permissions: failed to prepare query: ", err.Error())
		return make([]string, 0), err
	}
	res, err := query.Query(username)
	if err != nil {
		log.Error("Could not list user switch permissions: failed to execute query: ", err.Error())
		return make([]string, 0), err
	}
	permissions := make([]string, 0)
	for res.Next() {
		var permission string
		err := res.Scan(&permission)
		if err != nil {
			log.Error("Could get userSwitchPermissions. Failed to scan query: ", err.Error())
			return permissions, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

// Will return a boolean if a user has a switch permission
func UserHasSwitchPermission(username string, switchId string) (bool, error) {
	permissions, err := GetUserSwitchPermissions(username)
	if err != nil {
		log.Error("Failed to check for user permission: ", err.Error())
		return false, err
	}
	for _, permission := range permissions {
		if permission == switchId {
			return true, nil
		}
	}
	return false, nil
}

func SetPowerState(switchId string, isPoweredOn bool) (bool, error) {
	query, err := db.Prepare(`
	UPDATE switch SET Power=? WHERE Id=? 
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
	return true, nil
}
