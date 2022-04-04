package api

import (
	"encoding/json"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/hardware"
	"github.com/MikMuellerDev/smarthome/core/scheduler/scheduler"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

type NewScheduleRequest struct {
	Name           string `json:"name"`
	Hour           uint   `json:"hour"`
	Minute         uint   `json:"minute"`
	HomescriptCode string `json:"homescriptCode"` // Will be executed if the scheduler runs the job
}

type NewPowerScheduleRequest struct {
	Name      string                  `json:"name"`
	Hour      uint                    `json:"hour"`
	Minute    uint                    `json:"minute"`
	PowerJobs []hardware.PowerRequest `json:"powerJobs"`
}

type ModifyGenericScheduleRequest struct {
	Id             uint   `json:"id"`
	Name           string `json:"name"`
	Hour           uint   `json:"hour"`
	Minute         uint   `json:"minute"`
	HomescriptCode string `json:"homescriptCode"` // Will be executed if the scheduler runs the job
}

type ModifyPowerScheduleRequest struct {
	Id        uint                    `json:"id"`
	Name      string                  `json:"name"`
	Hour      uint                    `json:"hour"`
	Minute    uint                    `json:"minute"`
	PowerJobs []hardware.PowerRequest `json:"powerJobs"` // Will be parsed to HMS code
}

type DeleteScheduleRequest struct {
	Id uint `json:"id"`
}

type UserSchedulerEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

// Returns a list of all schedules set up by the current user
func GetUserSchedules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	schedules, err := database.GetUserSchedules(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to list personal schedules", Error: "internal server error"})
		return
	}
	if err := json.NewEncoder(w).Encode(schedules); err != nil {
		log.Error(err)
		Res(w, Response{Success: false, Message: "failed to list personal schedules", Error: "failed to encode response"})
	}
}

// Creates a new generic schedule which runs homescript
func CreateNewSchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request NewScheduleRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	if request.Hour > 24 || request.Minute > 60 { // Checks the minute and hour, values below 0 are checked implicitly through `uint`
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to create new schedule", Error: "invalid hour and / or minute"})
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
		Res(w, Response{Success: false, Message: "failed to add schedule", Error: "internal server error"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully created new schedule"})
}

// Modify a generic schedule which already exists
func ModifySchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request ModifyGenericScheduleRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	if request.Hour > 24 || request.Minute > 60 { // Checks the minute and hour, values below 0 are checked implicitly through `uint`
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to modify schedule", Error: "invalid hour and / or minute"})
		return
	}
	_, doesExists, err := scheduler.GetUserScheduleById(username, request.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to modify schedule", Error: "internal server error"})
		return
	}
	if !doesExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify schedule", Error: "invalid id / not found"})
		return
	}
	if err := scheduler.ModifyScheduleById(request.Id,
		database.Schedule{
			Name:           request.Name,
			Hour:           request.Hour,
			Minute:         request.Minute,
			HomescriptCode: request.HomescriptCode,
		},
	); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to modify schedule", Error: "internal server error"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully modified and restarted schedule"})
}

// Stops, then removes the given schedule from the system
func RemoveSchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request DeleteScheduleRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, doesExists, err := scheduler.GetUserScheduleById(username, request.Id) // Checks if the schedule exists and if the user is allowed to delete it
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to delete schedule", Error: "backend failure"})
		return
	}
	if !doesExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete schedule", Error: "invalid id / not found"})
		return
	}
	if err := scheduler.RemoveScheduleById(request.Id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to delete schedule", Error: "backend failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted schedule"})
}

// Set if the user scheduler is enabled or disabled
func SetUserSchedulerEnabled(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserSchedulerEnabledRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	user, found, err := database.GetUserByUsername(username)
	if err != nil || !found {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to set scheduler status", Error: "database failure"})
		return
	}
	if user.SchedulerEnabled == request.Enabled {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to set status of scheduler", Error: "scheduler is already in the requested mode"})
		return
	}
	if err := database.SetUserSchedulerEnabled(username, request.Enabled); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to set scheduler status", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully updated scheduler status"})
}
