package user

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Will return <true / false> based on authentication validity
// `true` means valid authentication parameters provided
// Can return an error if the database fails to return a valid result, meaning service downtime
func ValidateCredentials(username string, password string) (bool, error) {
	_, userExists, err := database.GetUserByUsername(username)
	if err != nil {
		log.Error("Failed to validate password: could not check if user exists: ", err.Error())
		return false, err
	}
	if !userExists {
		log.Trace("Credentials invalid: user does not exist")
		return false, nil
	}
	hash, err := database.GetUserPasswordHash(username)
	if err != nil {
		log.Error("Failed to validate password: database failure")
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true, nil
	}
	if err.Error() != "crypto/bcrypt: hashedPassword is not the hash of the given password" {
		log.Error("failed to check password: ", err.Error())
		return false, err
	}
	log.Trace("password check using bcrypt failed: passwords do not match")
	return false, nil
}

// Changes a users password to a new one
func ChangePassword(username string, newPassword string) error {
	// Generates a new password hash based on a provided computational `cost`
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		log.Error("Failed to create new user: password hashing failed", err.Error())
		return err
	}
	err = database.UpdateUserPasswordHash(
		username,
		string(hashedPassword),
	)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Password of user `%s` was changed successfully", username))
	event.Info("Password Changed", fmt.Sprintf("%s changed their password", username))
	return nil
}

// Removes a user, also removes everything that depends on the user:
// permissions, switchPermissions, cameraPermissions, notifications, reminders, schedulers, automations, homescripts
func DeleteUser(username string) error {
	if err := RemoveAvatar(username); err != nil {
		log.Error("Failed to delete user: removing avatar failed: ", err.Error())
		return err
	}
	if err := database.DeleteUser(username); err != nil {
		log.Error("Failed to delete user: database error: ", err.Error())
		return err
	}
	event.Info("User Deleted", fmt.Sprintf("User %s was deleted", username))
	return nil
}

// Checks if the user is the only entity with user management permission
func IsStandaloneUserAdmin(username string) (bool, error) {
	users, err := database.ListUsers()
	if err != nil {
		return false, err
	}
	for _, user := range users {
		hasPermission, err := database.UserHasPermission(
			user.Username,
			database.PermissionManageUsers,
		)
		if err != nil {
			return false, err
		}
		if hasPermission && user.Username != username {
			return false, nil
		}
	}
	return true, nil
}
