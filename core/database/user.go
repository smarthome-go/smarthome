package database

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Identified by a username, has a password and an avatar path
type FullUser struct {
	Username          string `json:"username"`
	Forename          string `json:"forename"`
	Surname           string `json:"surname"`
	PrimaryColorDark  string `json:"primaryColorDark"`
	PrimaryColorLight string `json:"primaryColorLight"`
	Password          string `json:"password"`
	AvatarPath        string `json:"avatarPath"`
	SchedulerEnabled  bool   `json:"schedulerEnabled"` // Specifies whether the user's schedules and automations may be executed
	DarkTheme         bool   `json:"darkTheme"`
}

type User struct {
	Username          string `json:"username"`
	Forename          string `json:"forename"`
	Surname           string `json:"surname"`
	PrimaryColorDark  string `json:"primaryColorDark"`
	PrimaryColorLight string `json:"primaryColorLight"`
	SchedulerEnabled  bool   `json:"schedulerEnabled"`
	DarkTheme         bool   `json:"darkTheme"`
}

type UserDetails struct {
	User        User     `json:"user"`
	Permissions []string `json:"permissions"`
}

// Used during <Init> of the database, only called once
// Creates the table containing <users> if it doesn't already exist
// Can return an error if the database fails
func createUserTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	user(
		Username          VARCHAR(20) PRIMARY KEY,
		Forename          VARCHAR(20) DEFAULT "Forename",
		Surname           VARCHAR(20) DEFAULT "Surname",
		PrimaryColorDark  CHAR(7)     DEFAULT "#88FF70",
		PrimaryColorLight CHAR(7)     DEFAULT "#2E7D32",
		SchedulerEnabled  BOOLEAN     DEFAULT TRUE,
		DarkTheme         BOOLEAN     DEFAULT TRUE,
		Password text,
		AvatarPath text
	)`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create user table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Lists users which are currently in the Database
func ListUsers() ([]User, error) {
	query := `
	SELECT
		Username,
		Forename,
		Surname,
		PrimaryColorDark,
		PrimaryColorLight,
		SchedulerEnabled,
		DarkTheme
	FROM user`
	res, err := db.Query(query)
	if err != nil {
		log.Error("Could not list users. Failed to execute query: ", err.Error())
		return nil, err
	}
	defer res.Close()
	userList := make([]User, 0)
	for res.Next() {
		var user User
		err := res.Scan(
			&user.Username,
			&user.Forename,
			&user.Surname,
			&user.PrimaryColorDark,
			&user.PrimaryColorLight,
			&user.SchedulerEnabled,
			&user.DarkTheme,
		)
		if err != nil {
			log.Error("Failed to scan user values from database results: ", err.Error())
			return nil, err
		}
		userList = append(userList, user)
	}
	return userList, nil
}

// Creates a new user based on a the supplied `User` struct
// Won't return an error if user already exists, but will change the password
func InsertUser(user FullUser) error {
	query, err := db.Prepare(`
	INSERT INTO
	user(
		Username,
		Forename,
		Surname,
		PrimaryColorDark,
		PrimaryColorLight,
		Password,
		AvatarPath,
		SchedulerEnabled,
		DarkTheme
	)
	VALUES(?, ?, ?, ?, ?, ?, ?, DEFAULT, DEFAULT)
	ON DUPLICATE KEY
	UPDATE
		Password=VALUES(Password)
	`)
	if err != nil {
		log.Error("Could not create user. Failed to prepare query: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(
		user.Username,
		user.Forename,
		user.Surname,
		user.PrimaryColorDark,
		user.PrimaryColorLight,
		user.Password,
		"./resources/avatar/default.png",
	)
	if err != nil {
		log.Error("Could not create user. Failed to execute query: ", err.Error())
		return err
	}
	return nil
}

// Deletes a User based on a given Username, can return an error if the database fails
// The function does not validate the existence of this username itself, so additional checks should be done beforehand
// The avatar is removed in `core/user/user`
func DeleteUser(username string) error {
	if err := RemoveAllPermissionsOfUser(username); err != nil {
		return err
	}
	if err := RemoveAllSwitchPermissionsOfUser(username); err != nil {
		return err
	}
	if err := DeleteAllNotificationsFromUser(username); err != nil {
		return err
	}
	if err := DeleteAllAutomationsFromUser(username); err != nil {
		return err
	}
	if err := DeleteAllHomescriptsOfUser(username); err != nil {
		return err
	}
	if err := DeleteAllSchedulesFromUser(username); err != nil {
		return err
	}
	if err := RemoveAllCameraPermissionsOfUser(username); err != nil {
		return err
	}
	query, err := db.Prepare(`
	DELETE FROM user
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Could not delete user. Failed to prepare query: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(username)
	if err != nil {
		log.Error("Could not delete user. Failed to execute query: ", err.Error())
		return err
	}
	return nil
}

// Helper function to create a User which is given a set of basic permissions
// Will return an error if the database fails
// TODO: Remove business logic from here, move to core/user
func AddUser(user FullUser) error {
	_, userExists, err := GetUserByUsername(user.Username)
	if err != nil {
		return err
	}
	if userExists {
		return errors.New("could not add user: user already exists")
	}
	// Generates a new password hash based on a provided computational `cost`
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to create new user: password hashing failed", err.Error())
		return err
	}
	user.Password = string(hashedPassword)
	if err = InsertUser(user); err != nil {
		return err
	}
	if err = AddNotification(user.Username, "Hello!", "Welcome to Smarthome, a self-built home automation system.", 1); err != nil {
		return err
	}
	log.Debug(fmt.Sprintf("Added user %s %s <%s>", user.Forename, user.Surname, user.Username))
	return nil
}

// Returns a user struct based on a username, does not check if the user exists, additional checks needed beforehand
func GetUserByUsername(username string) (User, bool, error) {
	query, err := db.Prepare(`
	SELECT
		Username,
		Forename,
		Surname,
		PrimaryColorDark,
		PrimaryColorLight,
		SchedulerEnabled,
		DarkTheme
	FROM user
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Could not get user by username: failed to prepare query: ", err.Error())
		return User{}, false, err
	}
	defer query.Close()
	var user User
	if err := query.QueryRow(username).Scan(
		&user.Username,
		&user.Forename,
		&user.Surname,
		&user.PrimaryColorDark,
		&user.PrimaryColorLight,
		&user.SchedulerEnabled,
		&user.DarkTheme,
	); err != nil {
		if err == sql.ErrNoRows {
			return User{}, false, nil
		}
		log.Error("Failed to get user by username: ", err.Error())
		return User{}, false, err
	}
	return user, true, nil
}

// Returns the users information and their permissions
func GetUserDetails(username string) (UserDetails, bool, error) {
	user, found, err := GetUserByUsername(username)
	if err != nil {
		return UserDetails{}, false, err
	}
	if !found {
		return UserDetails{}, false, nil
	}
	permissions, err := GetUserPermissions(username)
	if err != nil {
		return UserDetails{}, false, err
	}
	return UserDetails{
		User:        user,
		Permissions: permissions,
	}, true, nil
}

// Returns the password of a given user
func GetUserPasswordHash(username string) (string, error) {
	query, err := db.Prepare(`
	SELECT
		Password
	FROM user
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to get user password hash: preparing query failed: ", err.Error())
		return "", err
	}
	defer query.Close()
	var passwordHash string
	if err := query.QueryRow(username).Scan(&passwordHash); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		log.Error("Failed to get user password hash: executing query failed: ", err.Error())
	}
	return passwordHash, nil
}

// Returns the path of the avatar image of a given user, does not check if the user exists, additional checks needed beforehand
func GetAvatarPathByUsername(username string) (string, error) {
	query, err := db.Prepare(`
	SELECT
		AvatarPath
	FROM user
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Could not get avatar path by username: failed to prepare query: ", err.Error())
		return "", err
	}
	defer query.Close()
	var avatarPath string
	if err := query.QueryRow(username).Scan(&avatarPath); err != nil {
		log.Error("Could not get avatar path by username: failed to scan query reqults: ", err.Error())
		return "", err
	}
	return avatarPath, nil
}

// Sets the path of the avatar for a given user, does not check if the user exists, additional checks needed beforehand
func SetUserAvatarPath(username string, avatarPath string) error {
	query, err := db.Prepare(`
	UPDATE user
	SET AvatarPath=?
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to set AvatarPath for user: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(avatarPath, username)
	if err != nil {
		log.Error("Failed to set AvatarPath for user: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Set whether the scheduler is enabled for the current user
func SetUserSchedulerEnabled(username string, enabled bool) error {
	query, err := db.Prepare(`
	UPDATE user
	SET SchedulerEnabled=?
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to set SchedulerEnabled for user: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(enabled, username)
	if err != nil {
		log.Error("Failed to set SchedulerEnabled for user: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Set whether the user uses the dark theme or the light theme
func SetUserDarkThemeEnabled(username string, useDarkTheme bool) error {
	query, err := db.Prepare(`
	UPDATE user
	SET DarkTheme=?
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to set dark theme for user: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(useDarkTheme, username)
	if err != nil {
		log.Error("Failed to set dark theme for user: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Sets the users primary colors
func UpdateUserMetadata(username string, forename string, surname string, primaryColorDark string, primaryColorLight string) error {
	query, err := db.Prepare(`
	UPDATE user
	SET
		Forename=?,
		Surname=?,
		PrimaryColorDark=?,
		PrimaryColorLight=?
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to update user metadata: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(
		forename,
		surname,
		primaryColorDark,
		primaryColorLight,
		username,
	)
	if err != nil {
		log.Error("Failed to update user metadata: executing query failed: ", err.Error())
		return err
	}
	return nil
}
