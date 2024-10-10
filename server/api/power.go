package api

import (
	"encoding/json"
	"net/http"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/device/driver"
)

// TODO: replace with device interaction

type PowerRequest struct {
	Switch  string `json:"switch"`
	PowerOn bool   `json:"powerOn"`
}

// Returns the power draw points from the last N hours.
func GetPowerDrawFrom24Hours(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the records from the last N hours.
	const N = 24

	powerUsageData, err := driver.GetPowerUsageRecordsUnixMillis(N)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "could not get power usage data from the last 24 hours", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(powerUsageData); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to get power usage data from the last 24 hours", Error: "could not encode content"})
	}
}

// Returns all power draw data points.
func GetPowerDrawAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	powerUsageData, err := driver.GetPowerUsageRecordsUnixMillis(-1)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "could not get complete power usage data", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(powerUsageData); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to get complete power usage data", Error: "could not encode content"})
	}
}

// Is used to flush the power usage records manually.
// Deletes all records, regardless of their age.
func PurgePowerRecords(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, err := database.FlushPowerUsageRecords(0); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to purge power usage data", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully purged power usage data"})
}
