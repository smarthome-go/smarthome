package api

import (
	"encoding/json"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/scheduler/automation"
	"github.com/MikMuellerDev/smarthome/core/scheduler/scheduler"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

type NewScheduleRequest struct {
	Name           string `json:"name"`
	Hour           uint   `json:"hour"`
	Minute         uint   `json:"minute"`
	HomescriptCode string `json:"homescriptCode"` // Will be executed if the scheduler runs the job
}

type ModifyScheduleRequest struct {
	Id             uint   `json:"id"`
	Name           string `json:"name"`
	Hour           uint   `json:"hour"`
	Minute         uint   `json:"minute"`
	HomescriptCode string `json:"homescriptCode"` // Will be executed if the scheduler runs the job
}

type DeleteScheduleRequest struct {
	Id uint `json:"id"`
}

// Returns a list of all schedules set up by the current user
func GetUserSchedules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	schedules, err := database.GetUserSchedules(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to list personal schedules", Error: "internal server error"})
		return
	}
	json.NewEncoder(w).Encode(schedules)
}

// Creates a new schedule
func CreateNewSchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request NewScheduleRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	if request.Hour > 24 || request.Minute > 60 { // Checks the minute and hour, values below 0 are checked implicitly through `uint`
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to create new automation", Error: "invalid hour and / or minute"})
		return
	}
	if err := scheduler.CreateNewSchedule(database.Schedule{
		Name:           request.Name,
		Owner:          username,
		Hour:           request.Hour,
		Minute:         request.Minute,
		HomescriptCode: request.HomescriptCode,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add schedule", Error: "internal server error"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully created new schedule"})
}

// Stops, then removes the given schedule from the system
func RemoveSchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request DeleteScheduleRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, doesExists, err := database.Get(username, request.Id) // Checks if the automation exists and if the user is allowed to delete it
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to delete automation", Error: "backend failure"})
		return
	}
	if !doesExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to delete automation", Error: "invalid id / not found"})
		return
	}
	if err := automation.RemoveAutomation(request.Id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to delete automation", Error: "backend failure"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully deleted automation"})
}
