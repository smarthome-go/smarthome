package api

import (
	"encoding/json"
	"net/http"

	"github.com/smarthome-go/smarthome/core/config"
	"github.com/smarthome-go/smarthome/core/database"
)

type updateLocationRequest struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

type lockDownModeRequest struct {
	Enabled bool `json:"enabled"`
}

func GetSystemConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to get server configuration", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "could not get server configuration", Error: "failed to encode response"})
	}
}

// Can be used to enter and leave lockdown mode
func UpdateLockDownMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request lockDownModeRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	if err := database.SetLockDownModeEnabled(request.Enabled); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to update lock-down mode", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully updated lock-down mode"})
}

// Can be used to update the server's latitude and longitude
func UpdateLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request updateLocationRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	if err := database.UpdateLocation(request.Latitude, request.Longitude); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to update location", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully updated location"})
}

// Is used to request an export of the server's configuration
func ExportConfiguration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	export, err := config.Export()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to perform configuration export", Error: "internal server error"})
		return
	}
	if err := json.NewEncoder(w).Encode(export); err != nil {
		Res(w, Response{Success: false, Message: "failed to export server configuration", Error: "could not encode content"})
	}
}

// Is used to import a configuration using the `setup.json` structure
func ImportConfiguration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request config.SetupStruct
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	if err := config.RunSetupStruct(request); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to run setup", Error: err.Error()})
		return
	}
	Res(w, Response{Success: true, Message: "successfully ran setup"})
}

// Is used to flush the Homescript URL cache manually
// Deletes all recors older than 12 hours
func ClearHomescriptURLCache(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := database.FlushHomescriptUrlCache(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to flush Homescript URL cache", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully flushed URL cache"})
}

// Is used to flush the Homescript URL cache manually
// Deletes all records, regardless of their age
func PurgeHomescriptUrlCache(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := database.PurgeHomescriptUrlCache(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to purge Homescript URL cache", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully purged URL cache"})
}
