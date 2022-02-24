package database

import (
	"errors"
	"fmt"
)

// Used during <Init> of the database, only called once
// Creates the table containing <permissions> if it doesn't exists already
// Can return an error if the database fails
func createPermissionsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS
	hasPermission(
		Username VARCHAR(20),
		Permission VARCHAR(30),
		CONSTRAINT Username FOREIGN KEY (Username)
		REFERENCES users(Username)
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Could not create permissions table. Failed to execute query: ", err.Error())
	}
	return nil
}

// Adds a permission to a user, if database fails, then an error is returned
// Does not check for either username or permission validity, so additional checks should be completed beforehand
func AddUserPermission(username string, permission string) error {
	if !doesPermissionExist(permission) {
		log.Error("Will not add permission: Unknown permission type: ", permission)
		return errors.New("permission not found error: unknown permission type")
	}
	query, err := db.Prepare("INSERT INTO hasPermission(Username, Permission) VALUES(?,?) ON DUPLICATE KEY UPDATE Permission=VALUES(Permission)")
	if err != nil {
		log.Error("Could not add permission. Failed to prepare query: ", err.Error())
		return err
	}
	_, err = query.Exec(username, permission)
	if err != nil {
		log.Error("Could not add permission. Failed to execute query: ", err.Error())
		return err
	}
	log.Debug(fmt.Sprintf("Successfully added permission: `%s` to user: `%s`", permission, username))
	return nil
}



// Returns a list of permissions assigned to a given user, if it exists
func GetUserPermissions(username string) ([]string, error) {
	var permissions []string
	query, err := db.Prepare("SELECT Permission FROM hasPermission WHERE Username=?")
	if err != nil {
		log.Error("Could get user permissions. Failed to prepare query: ", err.Error())
		return permissions, err
	}
	res, err := query.Query(username)
	if err != nil {
		log.Error("Could get user permissions. Failed to execute query: ", err.Error())
		return permissions, nil
	}
	for res.Next() {
		var permission string
		err = res.Scan(&permission)
		if err != nil {
			log.Error("Could get user permissions. Failed to scan query: ", err.Error())
			return permissions, nil
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

// Checks the validity of a given permission string
func doesPermissionExist(permission string) bool {
	permissions := GetPermissions()
	for _, permissionItem := range permissions {
		if permissionItem == permission {
			return true
		}
	}
	return false
}
