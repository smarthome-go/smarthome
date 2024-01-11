package database

import (
	"database/sql"
)

// Stores the n:m relation between the user and their device-permissions
func createHasDevicePermissionTable() error {
	_, err := db.Query(`
	CREATE TABLE
	IF NOT EXISTS
	hasDevicePermission(
		Username    VARCHAR(20),
		Device      VARCHAR(20),
		FOREIGN KEY (Username)
		REFERENCES user(Username),
		FOREIGN KEY (Device)
		REFERENCES device(Id)
	)`)
	if err != nil {
		log.Error("Failed to create device permissions table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Adds a given device ID to a given user
// The existence of the device should be validated beforehand
// If this permission already resides inside the table, it is ignored and modified=false, error=nil is returned
// TODO: Remove useless check if user already has permission
func AddUserDevicePermission(username string, deviceId string) (bool, error) {
	userAlreadyHasPermission, err := UserHasDevicePermission(username, deviceId)
	if err != nil {
		log.Error("Failed to add permission: Could not validate the preexistence of a device permission: ", err.Error())
		return false, err
	}
	if userAlreadyHasPermission {
		return false, nil
	}

	query, err := db.Prepare(`
	INSERT INTO
	hasDevicePermission(
		Username,
		Device
	)
	VALUES(?, ?)
	`)
	if err != nil {
		log.Error("Could not add device permission to user: preparing query failed: ", err.Error())
		return false, err
	}
	defer query.Close()
	if _, err = query.Exec(username, deviceId); err != nil {
		log.Error("Failed to add device permission to user: executing query failed: ", err.Error())
		return false, err
	}
	return true, nil
}

// TODO: Remove useless check if user already has permission
// TODO: check naming consistency of `ADD / CREATE` and `DELETE / REMOVE`
// Removes a device permission from a user, but does not delete if from the device permission list
func RemoveUserDevicePermission(username string, deviceId string) (bool, error) {
	userHasPermission, err := UserHasDevicePermission(username, deviceId)
	if err != nil {
		log.Error("Failed to remove device permission from user: failed to check if user has permission")
		return false, err
	}
	if !userHasPermission {
		return false, nil
	}
	query, err := db.Prepare(`
	DELETE FROM
	hasDevicePermission
	WHERE Username=? AND Device=?
	`)
	if err != nil {
		log.Error("Failed to remove device permission from user: failed to prepare query: ", err.Error())
		return false, err
	}
	defer query.Close()
	if _, err = query.Exec(username, deviceId); err != nil {
		log.Error("Failed to remove device permission from user: executing query failed: ", err.Error())
		return false, nil
	}
	return true, nil
}

// Deletes all occurrences of a given device, used if a certain device is deleted
func RemoveDeviceFromPermissions(deviceId string) error {
	query, err := db.Prepare(`
	DELETE FROM
	hasDevicePermission
	WHERE Device=?
	`)
	if err != nil {
		log.Error("Failed to remove device completely from device permissions: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err = query.Exec(deviceId); err != nil {
		log.Error("Failed to remove device completely from device permissions: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Removes all device permissions of a given user, used when deleting a user
// Does not validate the existence of said user
func RemoveAllDevicePermissionsOfUser(username string) error {
	query, err := db.Prepare(`
	DELETE FROM
	hasDevicePermission
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to remove all device permissions of user: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(username); err != nil {
		log.Error("Failed to remove all device permissions of user: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns a list of strings which resemble device permissions
func GetUserDevicePermissions(username string) ([]string, error) {
	query, err := db.Prepare(`
	SELECT Device
	FROM hasDevicePermission
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Could not list user device permissions: failed to prepare query: ", err.Error())
		return make([]string, 0), err
	}
	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Could not list user device permissions: failed to execute query: ", err.Error())
		return make([]string, 0), err
	}
	defer res.Close()
	permissions := make([]string, 0)
	for res.Next() {
		var permission string
		err := res.Scan(&permission)
		if err != nil {
			log.Error("Could get user device permissions. Failed to scan query: ", err.Error())
			return permissions, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

// Used in uas has device permission
func UserHasDevicePermissionQuery(username string, deviceId string) (bool, error) {
	query, err := db.Prepare(`
	SELECT
		Device
	FROM hasDevicePermission
	WHERE Username=? AND Device=?
	`)
	if err != nil {
		log.Error("Failed to test user device permission: preparing query failed: ", err.Error())
		return false, err
	}
	if err := query.QueryRow(username, deviceId).Scan(&deviceId); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Error("Failed to test user device permission: executing query failed: ", err.Error())
		return false, err
	}
	return true, nil
}

// Returns a boolean if a user has a device permission
func UserHasDevicePermission(username string, deviceId string) (bool, error) {
	hasPermission, err := UserHasDevicePermissionQuery(username, deviceId)
	if err != nil {
		return false, err
	}
	if hasPermission {
		return true, nil
	}
	// If there is no matching permission, check for the '* | modifyRooms' permissions
	return UserHasPermission(username, PermissionModifyRooms)
}
