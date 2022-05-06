package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/server/middleware"
	"github.com/smarthome-go/smarthome/services/reminder"
)

type AddReminderRequest struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Priority    database.NotificationPriority `json:"priority"`
	DueDate     uint                          `json:"dueDate"` // Will be sent as unix millis
}
type ModifyReminderRequest struct {
	Id          uint                          `json:"id"`
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Priority    database.NotificationPriority `json:"priority"`
	DueDate     uint                          `json:"dueDate"` // Will be sent as unix millis
}

type AddedReminderResponse struct {
	Id      uint   `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RemoveReminderRequest struct {
	Id uint `json:"id"`
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
	// Checks if the due date is more than a month in the past
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
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add reminder", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(AddedReminderResponse{Id: id, Success: true, Message: fmt.Sprintf("successfully added reminder '%d'", id)}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "could not encode response"})
	}
}

// Returns a list of reminders that the current user has added
func GetReminders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	reminders, err := reminder.GetUserReminders(username)
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

// Deletes a reminder, for example if it's task is finished
func DeleteReminder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	var request RemoveReminderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, exists, err := database.GetReminderById(request.Id, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete reminder by id", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete reminder by id", Error: "reminder does not exist"})
		return
	}
	if err := database.DeleteUserReminderById(username, request.Id); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete reminder by id", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted reminder"})
}

// Modifies a reminder given its id and new metadata
func ModifyReminder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	var request ModifyReminderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, exists, err := database.GetReminderById(request.Id, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify reminder", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify reminder", Error: "reminder does not exist"})
		return
	}
	// Checks if the due date is more than a month in the past
	dueDate := time.Unix(int64(request.DueDate)/1000, 0)
	if dueDate.Before(time.Now().AddDate(0, -1, 0)) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify reminder", Error: "due date is more than 1 month in the past"})
		return
	}
	if request.Priority > 4 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify reminder", Error: "priority must be between 0 and 4"})
		return
	}
	if err := database.ModifyReminder(
		request.Id,
		request.Name,
		request.Description,
		dueDate,
		request.Priority,
	); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify reminder", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully updated reminder"})
}
