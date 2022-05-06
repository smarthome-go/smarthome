package user

import (
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
)

// Adds a permission to an arbitrary user.
// If the user already has the given permission, a database operation is omitted
// Does not validate the user's existence
func AddPermission(username string, permission database.PermissionType) (modified bool, err error) {
	alreadyHasPermission, err := database.UserHasPermission(username, permission)
	if err != nil {
		return false, err
	}
	if alreadyHasPermission {
		return false, nil
	}
	if err := database.AddUserPermission(username, permission); err != nil {
		return false, err
	}
	// Log event in order to inform administrators about a possible security flaw
	go event.Info("Added User Permission", fmt.Sprintf("Granted permission %s to user %s.", permission, username))
	return true, nil
}

// Removes a permission from an arbitrary user.
// If the user does not have the given permission, a database operation is omitted
// Does not validate the user's existence
func RemovePermission(username string, permission database.PermissionType) (modified bool, err error) {
	hasPermission, err := database.UserHasPermission(username, permission)
	if err != nil {
		return false, err
	}
	if !hasPermission {
		return false, nil
	}
	if err := database.RemoveUserPermission(username, permission); err != nil {
		return false, err
	}
	// Log event in order to inform administrators about a possible security flaw
	go event.Info("Removed User Permission", fmt.Sprintf("Removed permission %s from user %s.", permission, username))
	return true, nil
}
