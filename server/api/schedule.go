package api

import (
	"encoding/json"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

type NewScheduleRequest struct {
}

type ModifyScheduleRequest struct {
}

type DeleteScheduleRequest struct {
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
