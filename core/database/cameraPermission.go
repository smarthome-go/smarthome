package database

import "database/sql"

// Stores the n:m relation between the user and their camera-permissions
func createHasCameraPermissionsTable() error {
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	hasCameraPermission(
		Username VARCHAR(20),
		Camera   VARCHAR(50),
		FOREIGN KEY (Username)
		REFERENCES user(Username),
		FOREIGN KEY (Camera)
		REFERENCES camera(Id)
	)`); err != nil {
		log.Error("Failed to create hasCameraPermissionsTable: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns the camera-permissions of an arbitrary user in the form of a slice of strings
func GetUserCameraPermissions(username string) ([]string, error) {
	query, err := db.Prepare(`
	SELECT
		Camera
	FROM hasCameraPermission
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Could not list user camera permissions: failed to prepare query: ", err.Error())
		return make([]string, 0), err
	}
	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Could not list user switch camera: failed to execute query: ", err.Error())
		return make([]string, 0), err
	}
	defer res.Close()
	permissions := make([]string, 0)
	for res.Next() {
		var permission string
		err := res.Scan(&permission)
		if err != nil {
			log.Error("Could get userCameraPermissions. Failed to scan query: ", err.Error())
			return permissions, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

// Adds a given cameraId to an arbitrary user
// The existence of the camera and the user should be validated beforehand
func AddUserCameraPermission(username string, cameraId string) (modified bool, err error) {
	query, err := db.Prepare(`
	INSERT INTO
	hasCameraPermission(
		Username,
		Camera
	)
	VALUES(?, ?)
	`)
	if err != nil {
		log.Error("Failed to add camera permission to user: preparing query failed: ", err.Error())
		return false, err
	}
	defer query.Close()
	res, err := query.Exec(username, cameraId)
	if err != nil {
		log.Error("Failed to add camera permission to user: executing query failed: ", err.Error())
		return false, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("Failed to add camera permission to user: failed to retrieve rows affected count: ", err.Error())
		return false, err
	}
	return rowsAffected == 1, nil
}

// Removes a camera permission of an arbitrary user
func RemoveUserCameraPermission(username string, cameraId string) (modified bool, err error) {
	query, err := db.Prepare(`
	DELETE FROM
	hasCameraPermission
	WHERE Username=? AND Camera=?
	`)
	if err != nil {
		log.Error("Failed to remove user camera permission: preparing query failed: ", err.Error())
		return false, err
	}
	defer query.Close()
	res, err := query.Exec(username, cameraId)
	if err != nil {
		log.Error("Failed to remove user camera permission: executing query failed: ", err.Error())
		return false, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("Failed to add camera permission to user: failed to retrieve rows affected count: ", err.Error())
		return false, err
	}
	return rowsAffected == 1, nil
}

// Deletes all occurrences of a given camera, used if a camera is deleted
func RemoveCameraFromPermissions(cameraId string) error {
	query, err := db.Prepare(`
	DELETE FROM
	hasCameraPermission
	WHERE Camera=?
	`)
	if err != nil {
		log.Error("Failed to remove camera from permissions: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(cameraId); err != nil {
		log.Error("Failed to remove camera from permissions: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Removes all camera permissions of a given user, used when deleting a user
func RemoveAllCameraPermissionsOfUser(username string) error {
	query, err := db.Prepare(`
	DELETE FROM
	hasCameraPermission
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to remove all camera permissions of user: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(username); err != nil {
		log.Error("Failed to remove all camera permissions of user: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Used in userHasCameraPermission
func UserHasCameraPermissionQuery(username string, cameraId string) (bool, error) {
	query, err := db.Prepare(`
	SELECT
		Camera
	FROM hasCameraPermission
	WHERE Username=? AND Camera=?
	`)
	if err != nil {
		log.Error("Failed to check user camera permission: preparing query failed: ", err.Error())
		return false, err
	}
	defer query.Close()
	if err := query.QueryRow(username, cameraId).Scan(&cameraId); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Error("Failed to check user camera permission: executing query failed: ")
	}
	return true, nil
}

// Returns a boolean indicating whether a user has a camera permission
func UserHasCameraPermission(username string, cameraId string) (bool, error) {
	hasPermission, err := UserHasCameraPermissionQuery(username, cameraId)
	if err != nil {
		return false, err
	}
	if hasPermission {
		return true, nil
	}
	// If there is no matching permission, check for the '* | modifyRooms' permissions
	return UserHasPermission(username, PermissionModifyRooms)
}
