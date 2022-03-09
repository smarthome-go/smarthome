package database

import (
	"errors"
	"fmt"
)

// Used during <Init> of the database, only called once
// Creates the table containing <permissions> if it doesn't exists already
// Can return an error if the database fails
func createHasPermissionTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS
	hasPermission(
		Username VARCHAR(20),
		Permission VARCHAR(30),
		CONSTRAINT HasPermissionUsername FOREIGN KEY (Username)
		REFERENCES user(Username),
		CONSTRAINT HasPermissionPermission FOREIGN KEY (Permission)
		REFERENCES permission(Permission)
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Could not create hasPermission table: Executing query failed: ", err.Error())
	}
	return nil
}

// Used during <init> of the database, only called once
// May return an error if the database fails
func createPermissionTable() error {
	query := `
  CREATE TABLE
  IF NOT EXISTS
  permission(
	  Permission VARCHAR(30) PRIMARY KEY,
	  Name VARCHAR(100),
	  Description text
	)
  `
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Could not create permissions table: Executing query failed: ", err.Error())
	}
	return nil
}

// Creates permissions defined in `schemas.go` and inserts them into the permissions table
func initializePermissions() error {
	query, err := db.Prepare(`
	INSERT INTO
	permission(Permission, Name, Description)
	VALUES(?, ?, ?)
	ON DUPLICATE KEY UPDATE
	Name=VALUES(Name)`)
	if err != nil {
		log.Error("Failed to create permission: preparing query failed: ", err.Error())
		return err
	}
	permissions := GetPermissions()
	for _, permission := range permissions {
		res, err := query.Exec(permission.Permission, permission.Name, permission.Description)
		if err != nil {
			log.Error("Failed to create permission: executing query failed: ", err.Error())
			return err
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			log.Error("Failed to obtain rows affected: ", err.Error())
			return err
		}
		if rowsAffected > 0 {
			log.Debug("Inserted new permission into permissions table: ", permission.Permission)
		}
	}
	defer query.Close()
	return nil
}

// Adds a permission to a user, if database fails, then an error is returned
// Does not check for username so additional checks should be completed beforehand
func AddUserPermission(username string, permission string) (bool, error) {
	if !doesPermissionExist(permission) {
		log.Warn("Will not add permission: Unknown permission type: ", permission)
		// do not change error message: used in api
		return false, errors.New("permission not found error: unknown permission type")
	}
	alreadyHasPermission, err := UserHasPermission(username, permission)
	if err != nil {
		return false, err
	}
	if alreadyHasPermission {
		return false, nil
	}
	query, err := db.Prepare("INSERT INTO hasPermission(Username, Permission) VALUES(?,?) ON DUPLICATE KEY UPDATE Permission=VALUES(Permission)")
	if err != nil {
		log.Error("Could not add permission. Failed to prepare query: ", err.Error())
		return false, err
	}
	_, err = query.Exec(username, permission)
	if err != nil {
		log.Error("Could not add permission. Failed to execute query: ", err.Error())
		return false, err
	}
	log.Debug(fmt.Sprintf("Successfully added permission: `%s` to user: `%s`", permission, username))
	defer query.Close()
	return true, nil
}

// Attempts to remove a provided permission from a provided user
// Fails if permission does not exist or if the database fails
// Warns and returns `false` for the `modified` boolean the user does not have the permission
func RemoveUserPermission(username string, permission string) (bool, error) {
	if !doesPermissionExist(permission) {
		log.Warn("Will not remove permission: Unknown permission type: ", permission)
		return false, errors.New("permission not found error: unknown permission type")
	}
	hasPermission, err := UserHasPermission(username, permission)
	if err != nil {
		return false, err
	}
	if !hasPermission {
		log.Warn("Will not remove abundant permission: User does not have requested permission: ", permission)
		return false, nil
	}
	query, err := db.Prepare("DELETE FROM hasPermission WHERE username=? AND Permission=?")
	if err != nil {
		log.Error("Could not remove permission: Failed to prepare query: ", err.Error())
		return false, err
	}
	_, err = query.Exec(username, permission)
	if err != nil {
		log.Error("Failed to remove permission: Failed to execute query: ", err.Error())
		return false, err
	}
	log.Debug(fmt.Sprintf("Successfully removed permission: `%s` from user: `%s`", permission, username))
	defer query.Close()
	return true, nil
}

// Removes all permissions of a given user, used when deleting a user in order to prevent foreign key failure
// Does not validate username, additional checks required, returns an error if the database fails
func RemoveAllPermissionsOfUser(username string) error {
	query, err := db.Prepare(`DELETE FROM hasPermission WHERE Username=?`)
	if err != nil {
		log.Error("Could not delete all permissions of user: preparing query failed: ", err.Error())
		return err
	}
	if _, err = query.Exec(username); err != nil {
		log.Error("Could not delete all permissions of user: executing query failed: ", err.Error())
		return err
	}
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
	defer query.Close()
	return permissions, nil
}

// Checks the validity of a given permission string
func doesPermissionExist(permission string) bool {
	permissions := GetPermissions()
	for _, permissionItem := range permissions {
		if permissionItem.Permission == permission {
			return true
		}
	}
	return false
}

// Checks if a provided user is in possession of a provided permission, can return an error, if the database fails
func UserHasPermission(username string, permission string) (bool, error) {
	existentPermissions, err := GetUserPermissions(username)
	if err != nil {
		log.Error("Checking user permissions failed: Could not retrive permissions: ", err.Error())
		return false, err
	}
	for _, permissionItem := range existentPermissions {
		if permissionItem == "*" || permissionItem == permission {
			if !doesPermissionExist(permission) {
				log.Warn(fmt.Sprintf("failed to check permission `%s` for user `%s`: permission does not exist", permission, username))
				return false, nil
			}
			return true, nil
		}
	}
	if !doesPermissionExist(permission) {
		log.Warn(fmt.Sprintf("failed to check permission `%s` for user `%s`: permission does not exist", permission, username))
	}
	return false, nil
}
