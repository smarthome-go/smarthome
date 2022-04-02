package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Used during <Init> of the database, only called once
// Creates the table containing <users> if it doesn't already exist
// Can return an error if the database fails
func createUserTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	user(
		Username VARCHAR(20) PRIMARY KEY,
		Firstname VARCHAR(20) DEFAULT " ",
		Surname VARCHAR(20)   DEFAULT " ",
		PrimaryColor CHAR(7)  DEFAULT "#88ff70",
		Password text,
		AvatarPath text
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create user table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Lists users which are currently in the Database
// Returns an empty list with an error when failing
func ListUsers() ([]User, error) {
	query := `SELECT Username, Firstname, Surname, PrimaryColor FROM user`
	res, err := db.Query(query)
	if err != nil {
		log.Error("Could not list users. Failed to execute query: ", err.Error())
		return []User{}, err
	}
	var userList []User
	for res.Next() {
		var user User
		err := res.Scan(&user.Username, &user.Firstname, &user.Surname, &user.PrimaryColor)
		if err != nil {
			log.Error("Failed to scan user values from database results: ", err.Error())
		}
		userList = append(userList, user)
	}
	return userList, nil
}

// Creates a new user based on a the supplied `User` struct
// Won't panic if user already exists, but will change password
func InsertUser(user FullUser) error {
	query, err := db.Prepare(`
	INSERT INTO
	user(Username, Firstname, Surname, PrimaryColor, Password, AvatarPath)
	VALUES(?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE Password=VALUES(Password)`)
	if err != nil {
		log.Error("Could not create user. Failed to prepare query: ", err.Error())
		return err
	}
	_, err = query.Exec(
		user.Username,
		user.Firstname,
		user.Surname,
		user.PrimaryColor,
		user.Password,
		"./resources/avatar/default.png",
	)
	if err != nil {
		log.Error("Could not create user. Failed to execute query: ", err.Error())
		return err
	}
	defer query.Close()
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
	query, err := db.Prepare(`
	DELETE FROM user WHERE Username=?
	`)
	if err != nil {
		log.Error("Could not delete user. Failed to prepare query: ", err.Error())
		return err
	}
	_, err = query.Exec(username)
	if err != nil {
		log.Error("Could not delete user. Failed to execute query: ", err.Error())
		return err
	}
	defer query.Close()
	return nil
}

// Helper function to create a User which is given a set of basic permissions
// Will return an error if the database fails
// Does not check for duplicate users
func AddUser(user FullUser) error {
	userExists, err := DoesUserExist(user.Username)
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
	if _, err = AddUserPermission(user.Username, "authentication"); err != nil {
		return err
	}
	if err = AddNotification(user.Username, "Hello!", "Welcome to Smarthome, the privacy-focused home automation system.", 1); err != nil {
		return err
	}
	return nil
}

// Returns <true> if a provided user exists
// If the database fails, it returns an error
func DoesUserExist(username string) (bool, error) {
	userList, err := ListUsers()
	if err != nil {
		return false, err
	}
	for _, userItem := range userList {
		if userItem.Username == username {
			return true, nil
		}
	}
	return false, nil
}

// Returns a user struct based on a username, does not check if the user exists, additional checks needed beforehand
func GetUserByUsername(username string) (User, error) {
	query, err := db.Prepare(`
	SELECT
	Username, Firstname, Surname, PrimaryColor
	FROM user
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Could not get user by username: failed to prepare query: ", err.Error())
		return User{}, err
	}
	res, err := query.Query(username)
	if err != nil {
		log.Error("Could not get user by username: failed to execute query: ", err.Error())
		return User{}, err
	}
	user := User{}
	for res.Next() {
		err := res.Scan(&user.Username, &user.Firstname, &user.Surname, &user.PrimaryColor)
		if err != nil {
			log.Error("Failed to get user by username: failed to scan query: ", err.Error())
			return User{}, err
		}
	}
	return user, nil
}

// Returns the password of a given user
func GetUserPasswordHash(username string) (string, error) {
	query, err := db.Prepare(`
	SELECT Password FROM user WHERE Username=?
	`)
	if err != nil {
		log.Error("Failed to get user password hash: preparing query failed: ", err.Error())
		return "", err
	}
	var passwordHash string
	if err := query.QueryRow(username).Scan(&passwordHash); err != nil {
		log.Error("Failed to get user password hash: executing query failed: ", err.Error())
	}
	return passwordHash, nil
}

// Returns the path of the avatar image of a given user, does not check if the user exists, additional checks needed beforehand
func GetAvatarPathByUsername(username string) (string, error) {
	query, err := db.Prepare(`
	SELECT AvatarPath
	FROM user
	WHERE Username=?
	`)
	if err != nil {
		log.Error("Could not get avatar path by username: failed to prepare query: ", err.Error())
		return "", err
	}
	res, err := query.Query(username)
	if err != nil {
		log.Error("Could not get avatar path by username: failed to execute query: ", err.Error())
		return "", err
	}
	var avatarPath string
	for res.Next() {
		err := res.Scan(&avatarPath)
		if err != nil {
			log.Error("Failed to get avatar path by username: failed to scan query: ", err.Error())
			return "", err
		}
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
	_, err = query.Exec(avatarPath, username)
	if err != nil {
		log.Error("Failed to set AvatarPath for user: executing query failed: ", err.Error())
		return err
	}
	return nil
}
