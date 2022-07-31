package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/scheduler/scheduler"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type ModifyGenericScheduleRequest struct {
	Id   uint                  `json:"id"`
	Data database.ScheduleData `json:"data"`
}

type DeleteScheduleRequest struct {
	Id uint `json:"id"`
}

type CurrentUserSchedulerEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

type UserSchedulerEnabledRequest struct {
	Username string `json:"username"`
	Enabled  bool   `json:"enabled"`
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
	var request database.ScheduleData
	if err := decoder.Decode(&request); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	if request.Hour > 24 || request.Minute > 60 { // Checks the minute and hour, values below 0 are checked implicitly through `uint`
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to create new schedule", Error: "invalid hour and / or minute"})
		return
	}
	// Validate target-mode specific data
	switch request.TargetMode {
	case database.ScheduleTargetModeHMS:
		hmsData, HMSfound, err := database.GetUserHomescriptById(
			request.HomescriptTargetId,
			username,
		)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			Res(w, Response{Success: false, Message: "failed to validate `homescriptTargetId`", Error: "database failure"})
			return
		}
		if !HMSfound {
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "failed to create new schedule", Error: "invalid `homescriptTargetId`"})
			return
		}
		if !hmsData.Data.SchedulerEnabled {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Res(w, Response{Success: false, Message: "failed to modify schedule", Error: fmt.Sprintf("Homescript `%s` has disabled scheduler selection", request.HomescriptTargetId)})
			return
		}
	case database.ScheduleTargetModeSwitches:
		// Validate that the switchActions only contain valid switches

		// Only one switch action per switch is allowed
		// For routines or toggling, Homescript must be used
		existentSwitches := make([]string, 0)

		for _, switchItem := range request.SwitchJobs {
			// Validate that the switch is valid and accessible
			found, err := database.UserHasSwitchPermission(username, switchItem.SwitchId)
			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				Res(w, Response{Success: false, Message: "failed to validate `switchJobs`", Error: "database failure"})
				return
			}
			if !found {
				w.WriteHeader(http.StatusBadRequest)
				Res(w, Response{Success: false, Message: "failed to create new schedule", Error: fmt.Sprintf("invalid switch id:`%s`", switchItem.SwitchId)})
				return
			}

			// Check if the switch already exists
			for _, existentSwitch := range existentSwitches {
				if existentSwitch == switchItem.SwitchId {
					w.WriteHeader(http.StatusBadRequest)
					Res(w, Response{Success: false, Message: "failed to create new schedule", Error: fmt.Sprintf("second occurrence of switch `%s`: only one entry per switch-id allowed", switchItem.SwitchId)})
					return
				}
			}
			// Append to the existent switches
			existentSwitches = append(existentSwitches, switchItem.SwitchId)
		}
	case database.ScheduleTargetModeCode:
		// Nothing is validated (could validate Homescript via lint but is omitted)
		break
	default:
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to create new schedule", Error: fmt.Sprintf("invalid `targetMode`: `%s`", request.TargetMode)})
		return
	}
	if err := scheduler.CreateNewSchedule(request, username); err != nil {
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
	// Validate if the schedule exists
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
	// Validate hour and minute
	if request.Data.Hour > 24 || request.Data.Minute > 60 { // Checks the minute and hour, values below 0 are checked implicitly through `uint`
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify schedule", Error: "invalid hour and / or minute"})
		return
	}
	// Validate target-mode specific data
	switch request.Data.TargetMode {
	case database.ScheduleTargetModeHMS:
		hmsData, HMSfound, err := database.GetUserHomescriptById(
			request.Data.HomescriptTargetId,
			username,
		)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			Res(w, Response{Success: false, Message: "failed to validate `homescriptTargetId`", Error: "database failure"})
			return
		}
		if !HMSfound {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Res(w, Response{Success: false, Message: "failed to modify schedule", Error: "invalid `homescriptTargetId`"})
			return
		}
		if !hmsData.Data.SchedulerEnabled {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Res(w, Response{Success: false, Message: "failed to modify schedule", Error: fmt.Sprintf("Homescript `%s` has disabled scheduler selection", request.Data.HomescriptTargetId)})
			return
		}
	case database.ScheduleTargetModeSwitches:
		// Validate that the switchActions only contain valid switches

		// Only one switch action per switch is allowed
		// For routines or toggling, Homescript must be used
		existentSwitches := make([]string, 0)

		for _, switchItem := range request.Data.SwitchJobs {
			// Validate that the switch is valid and accessible
			found, err := database.UserHasSwitchPermission(username, switchItem.SwitchId)
			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				Res(w, Response{Success: false, Message: "failed to validate `switchJobs`", Error: "database failure"})
				return
			}
			if !found {
				w.WriteHeader(http.StatusBadRequest)
				Res(w, Response{Success: false, Message: "failed to modify schedule", Error: fmt.Sprintf("invalid switch id:`%s`", switchItem.SwitchId)})
				return
			}

			// Check if the switch already exists
			for _, existentSwitch := range existentSwitches {
				if existentSwitch == switchItem.SwitchId {
					w.WriteHeader(http.StatusBadRequest)
					Res(w, Response{Success: false, Message: "failed to modify schedule", Error: fmt.Sprintf("second occurrence of switch `%s`: only one entry per switch-id allowed", switchItem.SwitchId)})
					return
				}
			}
			// Append to the existent switches
			existentSwitches = append(existentSwitches, switchItem.SwitchId)
		}
	case database.ScheduleTargetModeCode:
		// Nothing is validated (could validate Homescript via lint but is omitted)
		break
	default:
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to modify schedule", Error: fmt.Sprintf("invalid `targetMode`: `%s`", request.Data.TargetMode)})
		return
	}
	if err := scheduler.ModifyScheduleById(request.Id, request.Data); err != nil {
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
func SetCurrentUserSchedulerEnabled(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request CurrentUserSchedulerEnabledRequest
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

// Set if the user scheduler is enabled or disabled
func SetUserSchedulerEnabled(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserSchedulerEnabledRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	user, found, err := database.GetUserByUsername(request.Username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to set scheduler status", Error: "database failure"})
		return
	}
	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to set scheduler status", Error: "invalid username"})
		return
	}
	if user.SchedulerEnabled == request.Enabled {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to set status of scheduler", Error: "scheduler is already in the requested mode"})
		return
	}
	if err := database.SetUserSchedulerEnabled(request.Username, request.Enabled); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to set scheduler status", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully updated scheduler status"})
}
