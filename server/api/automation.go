package api

import (
	"encoding/json"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/scheduler/automation"
	scheduler "github.com/MikMuellerDev/smarthome/core/scheduler/automation"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

type NewAutomationRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Hour         uint    `json:"hour"`   // 24 >= h >= 0 | Can only be used with minute, specifies the exact hour in which the automation will run, 0 is midnight, 15 is 3PM, 3 is 3AM -> 24h format
	Minute       uint    `json:"minute"` // 60 >= m >= 0 | Can only be used with hour, specifies the exact minute on which the automation will run
	Days         []uint8 `json:"days"`   //  6 >= d >= 0 | Can contain 7 elements at maximum, value `0` represents Sunday, value `6` represents Saturday
	HomescriptId string  `json:"homescriptId"`
}

type ModifyAutomationRequest struct {
	Id           uint    `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Hour         uint    `json:"hour"`
	Minute       uint    `json:"minute"`
	Days         []uint8 `json:"days"`
	HomescriptId string  `json:"homescriptId"`
}

type DeleteAutomationRequest struct {
	Id uint `json:"id"`
}

// Returns a list of all automations set up by the current user
func GetUserAutomations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	automations, err := scheduler.GetUserAutomations(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to list personal automations", Error: "internal server error"})
		return
	}
	json.NewEncoder(w).Encode(automations)
}

// Creates a new automation
func CreateNewAutomation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request NewAutomationRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Check if the provided HomescriptId is valid
	_, homescriptValid, err := database.GetUserHomescriptById(request.HomescriptId, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to create new automation", Error: "database failure"})
		return
	}
	if !homescriptValid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to create new automation", Error: "homescript id is invalid or not found"})
		return
	}
	// Check if the provided hour, minute and days are valid
	if len(request.Days) > 7 || len(request.Days) == 0 { // Check if there are more than 7 days or 0
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to create new automation", Error: "length of `days` cannot be greater than 7 or none (0)"})
		return
	}
	// Check for duplicates and if each provided day is valid
	containsDays := make([]uint8, 0) // Contains the days, is used to check if there are duplicates in the days
	for _, day := range request.Days {
		if day > 6 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to create new automation", Error: "invalid day in `days`: day must be >= 0 and <= 6"})
			return
		}
		dayIsAlreadyPresend := false
		for _, dayTemp := range containsDays {
			if dayTemp == day {
				dayIsAlreadyPresend = true
			}
		}
		if dayIsAlreadyPresend {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to create new automation", Error: "duplicate entries in `days`"})
			return
		}
		containsDays = append(containsDays, day) // If the day is not already present, add it
	}
	if request.Hour > 24 || request.Minute > 60 { // Checks the minute and hour, values below 0 are checked implicitly through `uint`
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to create new automation", Error: "invalid hour and / or minute"})
		return
	}
	if err := scheduler.CreateNewAutomation(
		request.Name,
		request.Description,
		uint8(request.Hour),
		uint8(request.Minute),
		request.Days,
		request.HomescriptId,
		username,
	); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to create new automation", Error: "backend failure"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully added new automation"})
}

// Stops, then removes the given automation from the system
func RemoveAutomation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request DeleteAutomationRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, doesExists, err := automation.GetUserAutomationById(username, request.Id) // Checks if the automation exists and if the user is allowed to delete it
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

// TODO: modify automations
func ModifyAutomation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request ModifyAutomationRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Check if the requested automation is valid
	_, automationValid, err := automation.GetUserAutomationById(username, request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to modify automation", Error: "database failure"})
		return
	}
	if !automationValid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to modify automation", Error: "automation id is invalid or not found"})
		return
	}
	// Check if the provided HomescriptId is valid
	_, homescriptValid, err := database.GetUserHomescriptById(request.HomescriptId, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to modify automation", Error: "database failure"})
		return
	}
	if !homescriptValid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to modify automation", Error: "homescript id is invalid or not found"})
		return
	}
	// Check if the provided hour, minute and days are valid
	if len(request.Days) > 7 || len(request.Days) == 0 { // Check if there are more than 7 days or 0
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to modify automation", Error: "length of `days` cannot be greater than 7 or none (0)"})
		return
	}
	// Check for duplicates and if each provided day is valid
	containsDays := make([]uint8, 0) // Contains the days, is used to check if there are duplicates in the days
	for _, day := range request.Days {
		if day > 6 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to modify automation", Error: "invalid day in `days`: day must be >= 0 and <= 6"})
			return
		}
		dayIsAlreadyPresend := false
		for _, dayTemp := range containsDays {
			if dayTemp == day {
				dayIsAlreadyPresend = true
			}
		}
		if dayIsAlreadyPresend {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to modify automation", Error: "duplicate entries in `days`"})
			return
		}
		containsDays = append(containsDays, day) // If the day is not already present, add it
	}
	if request.Hour > 24 || request.Minute > 60 { // Checks the minute and hour, values below 0 are checked implicitly through `uint`
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to modify automation", Error: "invalid hour and / or minute"})
		return
	}
	cronExpr, err := automation.GenerateCronExpression(
		uint8(request.Hour),
		uint8(request.Minute),
		request.Days,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to modify automation", Error: "could not create cron expression"})
		return
	}
	newAutomation := database.AutomationWithoutIdAndUsername{
		Name:           request.Name,
		Description:    request.Description,
		CronExpression: cronExpr,
		HomescriptId:   request.HomescriptId,
	}
	if err := automation.ModifyAutomationById(request.Id, newAutomation); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to modify automation", Error: "internal server error"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully modified automation"})
}
