package database

import "fmt"

// Stores the n:m relation between the user and their switch-permissions
func createHasSwitchPermissionTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	hasSwitchPermission(
		Username VARCHAR(20),
		Switch VARCHAR(20),
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

// Adds a given switchId to a given user
// The existence of the switch should be validated beforehand
// If this permission already resides inside the table, it is ignored and modified=false, error=nil is returned
// TODO: Remove useless check if user already has permission
func AddUserSwitchPermission(username string, switchId string) (bool, error) {
	userAlreadyHasPermission, err := UserHasSwitchPermission(username, switchId)
	if err != nil {
		log.Error("Failed to add permission: Could not validate the preexistence of a switchPermission: ", err.Error())
		return false, err
	}
	if userAlreadyHasPermission {
		return false, nil
	}
	query, err := db.Prepare(`
	INSERT INTO
	hasSwitchPermission(
		Username,
		Switch
	)
	VALUES(?,?)
	`)
	if err != nil {
		log.Error("Could not add switch permission to user: preparing query failed: ", err.Error())
		return false, err
	}
	defer query.Close()
	_, err = query.Exec(username, switchId)
	if err != nil {
		log.Error("Failed to add switch permission to user: executing query failed: ", err.Error())
		return false, err
	}
	defer query.Close()
	return true, nil
}

// TODO: Remove useless check if user already has permission
// TODO: check naming consistency of `ADD / CREATE` and `DELETE / REMOVE`
// Removes a switch permission from a user, but does not delete if from the switch permission list
func RemoveUserSwitchPermission(username string, switchId string) (bool, error) {
	userHasPermission, err := UserHasSwitchPermission(username, switchId)
	if err != nil {
		log.Error("Failed to remove permission from user: failed to check if user has permission")
		return false, err
	}
	if !userHasPermission {
		return false, nil
	}
	query, err := db.Prepare(`DELETE FROM hasSwitchPermission WHERE Username=? AND Switch=?`)
	if err != nil {
		log.Error("Failed to remove switch permission from user: failed to prepare query: ", err.Error())
		return false, err
	}
	defer query.Close()
	if _, err = query.Exec(username, switchId); err != nil {
		log.Error("Failed to remove switch permission from user: executing query failed: ", err.Error())
		return false, nil
	}
	return true, nil
}

// Deletes all occurrences of a given switch, used if a certain switch is deleted completely
func RemoveSwitchFromPermissions(switchId string) error {
	query, err := db.Prepare(`
	DELETE FROM
	hasSwitchPermission
	WHERE Switch=?
	`)
	if err != nil {
		log.Error("Failed to remove switch completely from switch permissions: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err = query.Exec(switchId); err != nil {
		log.Error("Failed to remove switch completely from switch permissions: executing query failed: ", err.Error())
		return err
	}
	log.Debug(fmt.Sprintf("Completely removed switch %s from switch permissions", switchId))
	return nil
}

// Removes all switch permission of a given user, used when deleing a user
// Does not validate the existence of said user
func RemoveAllSwitchPermissionsOfUser(username string) error {
	query, err := db.Prepare(`
	DELETE FROM
	hasSwitchPermission
	WHERE Username=?`)
	if err != nil {
		log.Error("Failed to remove all switch permissions of user: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(username); err != nil {
		log.Error("Failed to remove all switch permissions of user: executing query failed: ", err.Error())
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
	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Could not list user switch permissions: failed to execute query: ", err.Error())
		return make([]string, 0), err
	}
	defer res.Close()
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
// TODO: Replace with QueryRow
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
