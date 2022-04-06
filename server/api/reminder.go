package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

type AddReminderRequest struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Priority    database.NotificationPriority `json:"priority"`
	DueDate     uint                          `json:"dueDate"` // Will be sent as unix millis
	Owner       string                        `json:"owner"`
}

// Adds a new reminder to the database
func AddReminder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	var request AddReminderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	dueDate := time.Unix(int64(request.DueDate)/1000, 0)
	if dueDate.Before(time.Now().AddDate(0, -1, 0)) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add reminder", Error: "due date is more than 1 month in the past"})
		return
	}

	if request.Priority > 4 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add reminder", Error: "permission must be between 0 and 4"})
		return
	}

	id, err := database.CreateNewReminder(
		request.Name,
		request.Description,
		dueDate,
		username,
		request.Priority,
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to add reminder", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: fmt.Sprintf("successfully added reminder '%d'", id)})
}

// Returns a list of reminders that the current user has added
func GetReminders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	reminders, err := database.GetUserReminders(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to list reminders", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(reminders); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "could not encode response"})
	}
}
