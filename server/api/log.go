package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
)

// Triggers deletion of an arbitrary log event (as long as it exists), admin authentication required
func DeleteLogById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to delete log record", Error: "no id provided"})
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt < 0 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to delete log record", Error: "id is not numeric or < 0"})
		return
	}
	success, err := database.DeleteLogById(uint(idInt))
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete log record", Error: "database failure"})
		return
	}
	if !success {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete log record", Error: "invalid id provided"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted log record"})
}

// Triggers deletion of internal server logs which are older than 30 days, admin authentication required
// Request: empty | Response: Response
func FlushOldLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := event.FlushOldLogs(); err != nil {
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
	if err := event.FlushAllLogs(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to flush logs", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully flushed logs"})
}

// Returns a list of logging items in the logging table, admin authentication required
func ListLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	logs, err := event.GetAllLogsUnixMillis()
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
