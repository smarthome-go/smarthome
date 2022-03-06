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
// This image should ideally be in a 1:1 aspect ratio
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
	r.ParseMultipartForm(int64(maxUploadSize))
	file, handler, err := r.FormFile("file")
	if err != nil {
		// File to large or invalid file
		log.Debug("Could not retrieve file: ", err.Error())
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
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "avatar upload failed", Error: "invalid file type. allowed types are: [png / webp / jpeg / jpg]"})
		return
	}
	// Do the actual setup
	if err := user.UploadAvatar(username, handler.Filename, file); err != nil {
		log.Error("Could not update database entry: backend failed:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "avatar upload failed", Error: "internal server error"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "avatar uploaded successfully", Error: ""})
}

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
		w.WriteHeader(http.StatusBadGateway)
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
	json.NewEncoder(w).Encode(Response{Success: true, Message: "avatar removed successfully", Error: ""})
}

func getAvatar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
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
	if err != nil {
		log.Error("Could display avatar: could not read image", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(make([]byte, 0))
		return
	}
	w.Write(fileBytes)
}
