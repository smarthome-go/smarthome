package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func handleProfileUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// If the file is too large, reject the request
	if handler.Size > int64(maxUploadSize) {
		log.Error("file is over max filesize")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "file too large", Error: "could not use file: filesize too large"})
		return
	}
	// Create new profile file
	newFile, err := os.Create(handler.Filename)
	if err != nil {
		log.Error("Could not upload file: could not create file: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "file upload failed", Error: "could not create file"})
		return
	}
	defer newFile.Close()

	// Copy the uploaded file to the newly created file on the filesystem
	if _, err := io.Copy(newFile, file); err != nil {
		log.Error("Could not copy file: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully Uploaded File\n")
}
