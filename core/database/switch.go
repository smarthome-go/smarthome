package database

import "fmt"

// Creates the table containing switches
// If the database fails, this function can return an error
func createSwitchTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS
	switch(
		Id VARCHAR(2) PRIMARY KEY,
		Name VARCHAR(30)
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

func CreateSwitch(Id string, Name string) error {
	query, err := db.Prepare(`
	INSERT INTO switch(Id, Name) VALUES(?,?) ON DUPLICATE KEY UPDATE Name=Values(Name)
	`)
	if err != nil {
		log.Error("Failed to add switch: preparing query failed: ", err.Error())
		return err
	}
	res, err := query.Exec(Id, Name)
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
