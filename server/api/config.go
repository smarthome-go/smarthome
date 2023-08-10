package api

import (
	"encoding/json"
	"net/http"

	"github.com/smarthome-go/smarthome/core"
	"github.com/smarthome-go/smarthome/core/config"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/automation"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type updateLocationRequest struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

type lockDownModeRequest struct {
	Enabled bool `json:"enabled"`
}

type exportConfigurationRequest struct {
	IncludeProfilePictures bool `json:"includeProfilePictures"`
	IncludeCacheData       bool `json:"includeCacheData"`
}

type suntimes struct {
	SunriseHour   uint8 `json:"sunriseHour"`
	SunriseMinute uint8 `json:"sunriseMinute"`
	SunsetHour    uint8 `json:"sunsetHour"`
	SunsetMinute  uint8 `json:"sunsetMinute"`
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
	// Validate if the gelocation is valid
	if request.Latitude < -90 || request.Latitude > 90 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid latitude range: must be (> -90 and < 90)"})
		return
	}
	if request.Longitude < -180 || request.Longitude > 180 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid longitude range: must be (> -180 and < 180)"})
		return
	}
	if err := database.UpdateLocation(request.Latitude, request.Longitude); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to update location", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully updated location"})
}

func GetSunTimes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "could not get suntimes", Error: "database error"})
		return
	}

	rise, set := automation.CalculateSunRiseSet(serverConfig.Latitude, serverConfig.Longitude)
	if err := json.NewEncoder(w).Encode(suntimes{
		SunriseHour:   uint8(rise.Hour),
		SunriseMinute: uint8(rise.Minute),
		SunsetHour:    uint8(set.Hour),
		SunsetMinute:  uint8(set.Minute),
	}); err != nil {
		Res(w, Response{Success: false, Message: "failed to export server configuration", Error: "could not encode content"})
	}
}

// Is used to request an export of the server's configuration
func ExportConfiguration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request exportConfigurationRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	export, err := config.Export(request.IncludeProfilePictures, request.IncludeCacheData)
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
		Res(w, Response{Success: false, Message: "bad request", Error: err.Error()})
		return
	}
	if err := core.RunSetupStruct(request); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to run setup", Error: err.Error()})
		return
	}
	Res(w, Response{Success: true, Message: "successfully ran setup"})
	middleware.InitWithRandomKey()
}

// Is used to reset the Smarthome server to its factory settings
func FactoryReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := core.FactoryReset(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to reset to factory settings", Error: err.Error()})
		return
	}
	Res(w, Response{Success: true, Message: "factory settings were applied successfully"})
	middleware.InitWithRandomKey()
}
