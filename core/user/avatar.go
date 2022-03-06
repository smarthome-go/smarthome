package user

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/MikMuellerDev/smarthome/core/database"
)

func RemoveAvatar(username string) error {
	// Get current file path
	filepath, err := database.GetAvatarPathByUsername(username)
	if err != nil {
		return err
	}
	// Remove file from filesystem
	if err := os.Remove(filepath); err != nil {
		return err
	}
	// Set the default path in the database again
	if err := database.SetUserAvatarPath(username, "./web/assets/avatar/default.png"); err != nil {
		return err
	}
	return nil
}

func UploadAvatar(username string, filename string, file multipart.File) error {
	// Remove the old image first, if it exists
	filepathBefore, err := database.GetAvatarPathByUsername(username)
	if err != nil {
		return err
	}
	// Check if the user has a custom avatar
	if filepathBefore != "./web/assets/avatar/default.png" {
		// Remove file from filesystem, ignore errors
		if err := os.Remove(filepathBefore); err != nil {
			log.Warn("Could not remove avatar from user: maybe it was deleted manually: ", err.Error())
		}
	}
	// Create new profile file
	// generates a unique hash based on the username and filename combination
	hashPrefix := md5.Sum([]byte(fmt.Sprintf("%s%s", username, filename)))
	filepath := fmt.Sprintf("./data/avatars/%x_%s", hashPrefix, filename)
	var newFile *os.File
	newFile, err = os.Create(filepath)
	if err != nil {
		if err := os.Mkdir("./data", 0775); err != nil {
			log.Debug("Could not create data directory: likely exists")
		}
		if err := os.Mkdir("./data/avatars", 0775); err != nil {
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
	if _, err := io.Copy(newFile, file); err != nil {
		log.Error("Could not copy file to filesystem: ", err.Error())
		return err
	}
	// Update the file location in the database
	if err := database.SetUserAvatarPath(username, filepath); err != nil {
		return err
	}
	return nil
}
