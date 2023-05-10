package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/user"
	"github.com/smarthome-go/smarthome/server/middleware"
)

// Accepts the upload of an image of following allowed formats (png / webp / jpeg / jpg)
// Image should ideally be in `1:1` aspect ratio, authentication required`
func HandleAvatarUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	// Max upload size: 10 MB
	maxUploadSize := 10485760
	if err := r.ParseMultipartForm(int64(maxUploadSize)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to upload avatar", Error: "could not parse form"})
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		// File to large or invalid file
		log.Debug("Could not retrieve file: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to obtain file", Error: "failed to obtain file: could not get file from request"})
		return
	}
	defer file.Close()
	// If the file is too large, reject the request
	if handler.Size > int64(maxUploadSize) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		Res(w, Response{Success: false, Message: "file too large", Error: "could not use file: filesize too large"})
		return
	}
	// Check if the filename matches allowed formats (png / webp / jpeg / jpg)
	allowedFileEndings := []string{"png", "webp", "jpeg", "jpg"}
	fileEnding := strings.ToLower(strings.Split(handler.Filename, ".")[len(strings.Split(handler.Filename, "."))-1])
	var fileEndingValid bool
	for _, value := range allowedFileEndings {
		if fileEnding == value {
			fileEndingValid = true
		}
	}
	if !fileEndingValid {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		Res(w, Response{Success: false, Message: "avatar upload failed", Error: "invalid file type. allowed types are: [png / webp / jpeg / jpg]"})
		return
	}
	// Do the actual setup
	fileBuffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(fileBuffer, file); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "avatar upload failed", Error: fmt.Sprintf("IO error: %s", err.Error())})
		return
	}

	if err := user.UploadAvatar(username, handler.Filename, fileBuffer.Bytes()); err != nil {
		log.Error("Could not update database entry: backend failed: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "avatar upload failed", Error: "internal server error"})
		return
	}
	Res(w, Response{Success: true, Message: "avatar uploaded successfully"})
}

// Deletes the user's currently saved avatar and sets it to default, authentication required
func DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	filepathBefore, err := database.GetAvatarPathByUsername(username)
	if err != nil {
		log.Error("Could remove avatar: ", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "avatar removal failed", Error: "database failure"})
		return
	}
	// Check if the user has a custom avatar
	if filepathBefore == "./web/assets/avatar/default.png" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "avatar removal failed", Error: "the default avatar cannot be removed"})
		return
	}
	if err := user.RemoveAvatar(username); err != nil {
		log.Error("Could remove avatar: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "avatar removal failed", Error: "internal server error"})
		return
	}
	Res(w, Response{Success: true, Message: "avatar removed successfully"})
}

// Returns the user's current avatar as an image, authentication required
func GetAvatar(w http.ResponseWriter, r *http.Request) {
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	fileBytes, err := user.GetUserAvatar(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write(make([]byte, 0)); err != nil {
			log.Error("Failed to return avatar image: writing response bytes failed: ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Set cache validity of image to 2 hours and set the Content-Type
	w.Header().Set("Content-Type", http.DetectContentType(fileBytes))
	w.Header().Set("Cache-Control", "max-age=7200")

	if _, err := w.Write(fileBytes); err != nil {
		log.Error("Failed to return avatar image: writing response bytes failed: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Returns the avatar of any given user. Used for the user management panel
func GetForeignUserAvatar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username, ok := vars["username"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "no username provided", Error: "no username provided"})
		return
	}
	fileBytes, err := user.GetUserAvatar(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write(make([]byte, 0)); err != nil {
			log.Error("Failed to return avatar image: writing response bytes failed: ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Set cache validity of image to 2 hours and set the Content-Type
	w.Header().Set("Content-Type", http.DetectContentType(fileBytes))
	w.Header().Set("Cache-Control", "max-age=7200")

	if _, err := w.Write(fileBytes); err != nil {
		log.Error("Failed to return avatar image: writing response bytes failed: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
