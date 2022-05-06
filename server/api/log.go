package api

import (
	"encoding/json"
	"net/http"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
)

// Triggers deletion of internal server logs which are older than 30 days, admin authentication required
// Request: empty | Response: Response
func FlushOldLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := database.FlushOldLogs(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to flush logs", Error: "database failure"})
		return
	}
	go event.Info("Flushed Old Logs", "Logs which are older than 30 days were deleted.")
	Res(w, Response{Success: true, Message: "successfully flushed logs older than 30 days"})
}

// Triggers deletion of ALL internal server logs, admin authentication required
// Request: empty | Response: Response
func FlushAllLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := database.FlushAllLogs(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to flush logs", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully flushed logs"})
}

// Returns a list of logging items in the logging table, admin authentication required
func ListLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	logs, err := database.GetLogs()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "could not get logs", Error: "failed to encode response"})
	}
}
