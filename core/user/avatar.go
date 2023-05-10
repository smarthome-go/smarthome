package user

import (
	"crypto/md5"
	"fmt"
	"os"

	"github.com/smarthome-go/smarthome/core/database"
)

const defaultFilePath = "./resources/avatar/default.png"

// Will remove the current avatar of a user unless it is set to `default`
func RemoveAvatar(username string) (critical error) {
	// Get current file path
	filepath, err := database.GetAvatarPathByUsername(username)
	if err != nil {
		return err
	}
	if filepath == defaultFilePath {
		log.Trace("Avatar picture is equal to the default image: omitting disk removal")
		return nil
	}
	// Remove file from the filesystem
	if err := os.Remove(filepath); err != nil {
		log.Error(fmt.Sprintf("Failed to remove avatar file at `%s` from disk: \nResetting avatar path to address future errors...", filepath))
	}
	// Set the default path in the database again
	if err := database.SetUserAvatarPath(username, defaultFilePath); err != nil {
		return err
	}
	return nil
}

// Returns the avatar image bytes of an arbitrary user
func GetUserAvatar(username string) ([]byte, error) {
	// Attempt to get the avatar's file path from the database
	filepath, err := database.GetAvatarPathByUsername(username)
	if err != nil {
		return nil, err
	}
	// Atttempt to read the file into a byte slice
	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		// Standard error handling
		log.Error(fmt.Sprintf("Could not get avatar of user `%s`: Could not read file from disk: %s", username, err.Error()))
		if filepath == defaultFilePath {
			return nil, err
		}

		// If the user has set a non-default avatar, reset it
		log.Debug("Attempting to reset to default path and read default avatar file due to previous failure...")
		if err := RemoveAvatar(username); err != nil {
			return nil, err
		}
		log.Info("Scccessfully reset avatar to use the default file path")

		// Serve the default avatar as a replacement
		fileBytes, err = os.ReadFile(defaultFilePath)
		if err != nil {
			log.Error(fmt.Sprintf("Could not get default avatar of user `%s`: Could not read default file from disk: %s", username, err.Error()))
			return nil, err
		}
		return fileBytes, nil
	}
	return fileBytes, nil
}

// Accepts a username, filename and multipart file and creates and processes the file
func UploadAvatar(
	username string,
	filename string,
	file []byte,
) error {
	// Remove the old image first, if it exists
	filepathBefore, err := database.GetAvatarPathByUsername(username)
	if err != nil {
		return err
	}
	// generates a unique hash based on the username and filename combination
	hashPrefix := md5.Sum([]byte(fmt.Sprintf("%s%s", username, filename)))
	filepath := fmt.Sprintf("./data/avatar/%x_%s", hashPrefix, filename)
	// If the filepath is equal, the hash did not change which means that the file is equal and will not be written to disk again
	if filepath == filepathBefore {
		// Stop if file is unchanged
		log.Trace(fmt.Sprintf("Not writing avatar file: hash unchanged (%s)", filepath))
		return nil
	}
	// Check if the user has a custom avatar
	if filepathBefore != defaultFilePath {
		// Remove file from filesystem, ignore errors
		if err := os.Remove(filepathBefore); err != nil {
			log.Warn("Could not remove avatar from user: maybe it was deleted manually: ", err.Error())
			return err
		}
	}
	// Create new profile file
	newFile, err := os.Create(filepath)
	if err != nil {
		if err := os.Mkdir("./data", 0775); err != nil {
			log.Debug("Could not create data directory: likely exists")
		}
		if err := os.Mkdir("./data/avatar", 0775); err != nil {
			log.Error("Could not upload file: could not create new directory: ", err.Error())
			return err
		}
		newFile, err = os.Create(filepath)
		if err != nil {
			log.Error("Could not upload file: could not create file inside new directory: ", err.Error())
			return err
		}
	}
	defer newFile.Close()

	// Copy the uploaded file to the newly created file on the filesystem
	if _, err := newFile.Write(file); err != nil {
		log.Error("Could not copy file to filesystem: ", err.Error())
		return err
	}

	// Update the file location in the database
	if err := database.SetUserAvatarPath(username, filepath); err != nil {
		return err
	}
	log.Trace("Successfully updated avatar")
	return nil
}
