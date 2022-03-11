package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/user"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

// Accepts the upload of an image of following allowed formats (png / webp / jpeg / jpg)
// Image should ideally be in `1:1` aspect ratio, authentication required`
func handleAvatarUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	// Max upload size: 10 MB
	maxUploadSize := 10485760
	if err := r.ParseMultipartForm(int64(maxUploadSize)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to upload avatar", Error: "could not parse form"})
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		// File to large or invalid file
		log.Debug("Could not retrieve file: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to obtain file", Error: "failed to obtain file: could not get file from request"})
		return
	}
	defer file.Close()
	// If the file is too large, reject the request
	if handler.Size > int64(maxUploadSize) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "file too large", Error: "could not use file: filesize too large"})
		return
	}
	// Check if the filename matches allowed formats (png / webp / jpeg / jpg)
	allowedFileEndings := []string{"png", "webp", "jpeg", "jpg"}
	fileEnding := strings.Split(handler.Filename, ".")[len(strings.Split(handler.Filename, "."))-1]
	var fileEndingValid bool
	for _, value := range allowedFileEndings {
		if fileEnding == value {
			fileEndingValid = true
		}
	}
	if !fileEndingValid {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "avatar upload failed", Error: "invalid file type. allowed types are: [png / webp / jpeg / jpg]"})
		return
	}
	// Do the actual setup
	if err := user.UploadAvatar(username, handler.Filename, file); err != nil {
		log.Error("Could not update database entry: backend failed: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "avatar upload failed", Error: "internal server error"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "avatar uploaded successfully"})
}

// Deletes the user's currently saved avatar and sets it to default, authentication required
func deleteAvatar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	filepathBefore, err := database.GetAvatarPathByUsername(username)
	if err != nil {
		log.Error("Could remove avatar: ", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "avatar removal failed", Error: "database error"})
		return
	}
	// Check if the user has a custom avatar
	if filepathBefore == "./web/assets/avatar/default.png" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "avatar removal failed", Error: "the default avatar cannot be removed"})
		return
	}
	if err := user.RemoveAvatar(username); err != nil {
		log.Error("Could remove avatar: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "avatar removal failed", Error: "internal server error"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "avatar removed successfully"})
}

// Returns the user's current avatar as an image, authentication required
// TODO: automatically detect content type
// TODO: Test autodetect of content type
func getAvatar(w http.ResponseWriter, r *http.Request) {
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	var filepath string
	filepath, err = database.GetAvatarPathByUsername(username)
	if err != nil {
		log.Error("Could not get avatar image: panic serving default image: ", err.Error())
		filepath = "./web/assets/avatar/default.png"
	}
	fileBytes, err := ioutil.ReadFile(filepath)
	w.Header().Set("Content-Type", http.DetectContentType(fileBytes))
	if err != nil {
		log.Error("Could not display avatar: could not read image", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write(make([]byte, 0)); err != nil {
			log.Error("Failed to return avatar image: writing response bytes failed: ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if _, err := w.Write(fileBytes); err != nil {
		log.Error("Failed to return avatar image: writing response bytes failed: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
