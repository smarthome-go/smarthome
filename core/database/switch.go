package database

import "fmt"

// Creates the table containing switches
// If the database fails, this function can return an error
func createSwitchTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS
	switch(
		Id VARCHAR(2) PRIMARY KEY,
		Name VARCHAR(30),
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
	INSERT INTO switch(Id, Name, RoomId) VALUES(?,?,?) ON DUPLICATE KEY UPDATE Name=Values(Name)
	`)
	if err != nil {
		log.Error("Failed to add switch: preparing query failed: ", err.Error())
		return err
	}
	res, err := query.Exec(Id, Name, RoomId)
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

func ListSwitches() ([]Switch, error) {
	query := `
	SELECT Id, Name, RoomId FROM switch
	`
	res, err := db.Query(query)
	if err != nil {
		log.Error("Could not list switches: failed to execute query: ", err.Error())
		return []Switch{}, err
	}

	var switches []Switch
	for res.Next() {
		var switchItem Switch
		if err := res.Scan(&switchItem.Id, &switchItem.Name, &switchItem.RoomId); err != nil {
			log.Error("Could not list switches: Failed to scan results: ", err.Error())
		}
		switches = append(switches, switchItem)
	}
	return switches, nil
}

func ListUserSwitches() ([]Switch, error) {
	query, err := db.Prepare(`
	SELECT Id, Name, RoomId FROM switch JOIN hasSwitchPermission ON hasSwitchPermission.Switch=switch.Id WHERE hasSwitchPermission.Username=?`)
}
