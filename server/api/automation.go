package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/core/homescript/automation"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type NewAutomationRequest struct {
	Name         string                     `json:"name"`
	Description  string                     `json:"description"`
	HomescriptId string                     `json:"homescriptId"`
	Enabled      bool                       `json:"enabled"`
	Trigger      database.AutomationTrigger `json:"trigger"`

	// For the `cron` trigger
	Hour   *uint `json:"hour"`   // 24 >= h >= 0 | Can only be used with minute, specifies the exact hour in which the automation will run, 0 is midnight, 15 is 3PM, 3 is 3AM -> 24h format
	Minute *uint `json:"minute"` // 60 >= m >= 0 | Can only be used with hour, specifies the exact minute on which the automation will run

	// For the `cron` and `sunrise` / `sunset` triggers
	Days *[]uint8 `json:"days"` //  6 >= d >= 0 | Can contain 7 elements at maximum, value `0` represents Sunday, value `6` represents Saturday

	// For the `interval` trigger
	TriggerIntervalSeconds *uint `json:"triggerInterval"`
}

type ModifyAutomationRequest struct {
	Id                     uint                       `json:"id"`
	Name                   string                     `json:"name"`
	Description            string                     `json:"description"`
	Hour                   uint                       `json:"hour"`
	Minute                 uint                       `json:"minute"`
	Days                   []uint8                    `json:"days"`
	HomescriptId           string                     `json:"homescriptId"`
	Enabled                bool                       `json:"enabled"`
	DisableOnce            bool                       `json:"disableOnce"`
	Trigger                database.AutomationTrigger `json:"trigger"`
	TriggerCronExpression  *string                    `json:"triggerCronExpression"`
	TriggerIntervalSeconds *uint                      `json:"triggerInterval"`
}

type DeleteAutomationRequest struct {
	Id uint `json:"id"`
}

type AutomationActivationRequest struct {
	Enabled bool `json:"enabled"`
}

// Returns a list of all automations set up by the current user
func GetUserAutomations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	automations, err := homescript.GetUserAutomations(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to list personal automations", Error: "internal server error"})
		return
	}
	if err := json.NewEncoder(w).Encode(automations); err != nil {
		Res(w, Response{Success: false, Message: "failed to list personal automations", Error: "could not encode content"})
	}
}

// Creates a new automation
func CreateNewAutomation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request NewAutomationRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Check if the provided HomescriptId is valid
	hmsData, homescriptValid, err := database.GetUserHomescriptById(request.HomescriptId, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create new automation", Error: "database failure"})
		return
	}
	if !homescriptValid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create new automation", Error: "homescript id is invalid or not found"})
		return
	}

	// Check if automations are enabled for the user
	if !hmsData.Data.SchedulerEnabled {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create automation", Error: fmt.Sprintf("Homescript `%s` has disabled scheduler selection", request.HomescriptId)})
		return
	}

	// Check if the trigger is valid
	if !database.IsValidAutomationTrigger(string(request.Trigger)) { // TODO: is this even required?
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to create new automation", Error: "invalid timing mode"})
		return
	}

	// TODO: add errors when fields are populated when there is no need

	// Perform checks for the different triggers
	switch request.Trigger {
	case database.TriggerCron:
		// Check if the provided hour, minute and days are valid
		if *request.Hour > 24 || *request.Minute > 60 { // Checks the minute and hour, values below 0 are checked implicitly through `uint`
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "failed to create new automation", Error: "invalid hour and / or minute"})
			return
		}

		if len(*request.Days) > 7 || len(*request.Days) == 0 { // Check if there are more than 7 days or 0
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "failed to create new automation", Error: "length of `days` cannot be greater than 7 or none (0)"})
			return
		}

		// Check for duplicates and if each provided day is valid
		containsDays := make([]uint8, 0) // Contains the days, is used to check if there are duplicates in the days
		for _, day := range *request.Days {
			if day > 6 {
				w.WriteHeader(http.StatusBadRequest)
				Res(w, Response{Success: false, Message: "failed to create new automation", Error: "invalid day in `days`: day must be >= 0 and <= 6"})
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
				Res(w, Response{Success: false, Message: "failed to create new automation", Error: "duplicate entries in `days`"})
				return
			}
			containsDays = append(containsDays, day) // If the day is not already present, add it
		}

		// Make sure that there is no trigger
		if request.TriggerIntervalSeconds != nil {
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "failed to create new automation", Error: "trigger interval must be null when using `cron`"})
			return
		}
	case database.TriggerInterval:
		if request.TriggerIntervalSeconds == nil {
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "failed to create new automation", Error: "trigger interval must not be null when using the interval trigger"})
			return
		}
		if *request.TriggerIntervalSeconds > (60 * 60 * 24 * 356) {
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "failed to create new automation", Error: "trigger interval must must be 0 > i > 60*60*24*356"})
			return
		}

		// Make sure that there is no cron-information
		if request.Days != nil || request.Hour != nil || request.Minute != nil {
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "failed to create new automation", Error: "`days`, `hour`, and `minute` can only be used with `cron`"})
			return
		}
	default:
		if request.TriggerIntervalSeconds != nil || request.Days != nil || request.Hour != nil || request.Minute != nil {
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "failed to create new automation", Error: "`days`, `hour`, `minute`, and `interval` can not be used in with this trigger"})
			return
		}
	}

	id, err := homescript.CreateNewAutomation(
		request.Name,
		request.Description,
		request.HomescriptId,
		username,
		request.Enabled,
		request.Hour,
		request.Minute,
		request.Days,
		request.Trigger,
		request.TriggerIntervalSeconds,
	)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to create new automation", Error: "backend failure"})
		return
	}
	Res(w, Response{Success: true, Message: fmt.Sprintf("successfully added new automation %d", id)})
}

// Stops, then removes the given automation from the system
func RemoveAutomation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request DeleteAutomationRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, doesExists, err := homescript.GetUserAutomationById(username, request.Id) // Checks if the automation exists and if the user is allowed to delete it
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to delete automation", Error: "backend failure"})
		return
	}
	if !doesExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete automation", Error: "invalid id / not found"})
		return
	}
	if err := homescript.RemoveAutomation(request.Id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to delete automation", Error: "backend failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted automation"})
}

// Modifies a existing automation, also restarts the schedule
func ModifyAutomation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request ModifyAutomationRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Check if the requested automation is valid
	_, automationValid, err := homescript.GetUserAutomationById(username, request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify automation", Error: "database failure"})
		return
	}
	if !automationValid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify automation", Error: "automation id is invalid or not found"})
		return
	}
	// Check if the provided HomescriptId is valid
	hmsData, homescriptValid, err := database.GetUserHomescriptById(request.HomescriptId, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify automation", Error: "database failure"})
		return
	}
	if !homescriptValid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify automation", Error: "homescript id is invalid or not found"})
		return
	}
	if !hmsData.Data.SchedulerEnabled {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify automation", Error: fmt.Sprintf("Homescript `%s` has disabled scheduler selection", request.HomescriptId)})
		return
	}

	var TriggerCronExpression *string = nil

	if request.Trigger == database.TriggerCron {
		if request.Hour > 24 || request.Minute > 60 { // Checks the minute and hour, values below 0 are checked implicitly through `uint`
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "failed to modify automation", Error: "invalid hour and / or minute"})
			return
		}

		// Check if the provided hour, minute and days are valid
		if len(request.Days) > 7 || len(request.Days) == 0 { // Check if there are more than 7 days or 0
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "failed to modify automation", Error: "length of `days` cannot be greater than 7 or none (0)"})
			return
		}
		// Check for duplicates and if each provided day is valid
		containsDays := make([]uint8, 0) // Contains the days, is used to check if there are duplicates in the days
		for _, day := range request.Days {
			if day > 6 {
				w.WriteHeader(http.StatusBadRequest)
				Res(w, Response{Success: false, Message: "failed to modify automation", Error: "invalid day in `days`: day must be >= 0 and <= 6"})
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
				Res(w, Response{Success: false, Message: "failed to modify automation", Error: "duplicate entries in `days`"})
				return
			}
			containsDays = append(containsDays, day) // If the day is not already present, add it
		}

		cronExpr, err := automation.GenerateCronExpression(
			uint8(request.Hour),
			uint8(request.Minute),
			request.Days,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			Res(w, Response{Success: false, Message: "failed to modify automation", Error: "could not create cron expression"})
			return
		}

		TriggerCronExpression = &cronExpr
	}

	// TODO: validate other stuff

	newAutomation := database.AutomationData{
		Name:                   request.Name,
		Description:            request.Description,
		HomescriptId:           request.HomescriptId,
		Enabled:                request.Enabled,
		DisableOnce:            request.DisableOnce,
		Trigger:                request.Trigger,
		TriggerCronExpression:  TriggerCronExpression,
		TriggerIntervalSeconds: request.TriggerIntervalSeconds,
	}
	if err := homescript.ModifyAutomationById(request.Id, newAutomation); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to modify automation", Error: "internal server error"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully modified automation"})
}

// Activate or deactivate the entire automation system
func ChangeActivationAutomation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request AutomationActivationRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	serverConfig, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to change activation state of automations", Error: "database failure"})
		return
	}
	if serverConfig.AutomationEnabled == request.Enabled {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to change activation state of automations", Error: fmt.Sprintf("current activation mode of automation is already set to %t", serverConfig.AutomationEnabled)})
		return
	}

	if err := database.SetAutomationSystemActivation(request.Enabled); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to change automation system status", Error: "database failure"})
		return
	}

	if request.Enabled {
		// HACK: manually alter the in-memory state of the automation system
		// otherwise, the automation wont be re-activated properly
		serverConfig.AutomationEnabled = request.Enabled

		if err := homescript.ActivateAutomationSystem(serverConfig); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			Res(w, Response{Success: false, Message: "failed to activate automations", Error: "internal server error"})
			return
		}
	} else {
		if err := homescript.DeactivateAutomationSystem(serverConfig); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			Res(w, Response{Success: false, Message: "failed to deactivate automations", Error: "internal server error"})
			return
		}
	}
	Res(w, Response{Success: true, Message: fmt.Sprintf("successfully set activation mode of automations to %t", request.Enabled)})
}
