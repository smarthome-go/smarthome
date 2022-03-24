package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/homescript"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

type HomescriptResponse struct {
	Success  bool                         `json:"success"`
	Exitcode int                          `json:"exitCode"`
	Message  string                       `json:"message"`
	Output   string                       `json:"output"`
	Errors   []homescript.HomescriptError `json:"error"`
}

type AddHomescriptRequest struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	QuickActionsEnabled bool   `json:"quickActionsEnabled"`
	SchedulerEnabled    bool   `json:"schedulerEnabled"`
	Code                string `json:"code"`
}

type HomescriptRequest struct {
	Code string `json:"code"`
}

// Runs any given Homescript as a string
func RunHomescriptString(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request HomescriptRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	output, exitCode, hmsErrors := homescript.Run(username, "live", request.Code)
	if len(hmsErrors) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(HomescriptResponse{
			Success:  false,
			Exitcode: exitCode,
			Message:  "Homescript terminated abnormally",
			Output:   output,
			Errors:   hmsErrors,
		})
		return
	}
	if exitCode != 0 {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(HomescriptResponse{
			Success:  false,
			Exitcode: exitCode,
			Message:  "Homescript exited with a non-0 status code",
			Output:   output,
			Errors:   hmsErrors,
		})
		return
	}
	json.NewEncoder(w).Encode(HomescriptResponse{
		Success:  true,
		Message:  "Homescript ran successfully",
		Output:   output,
		Exitcode: exitCode,
		Errors:   hmsErrors,
	})
}

// Returns a list of homescripts which are owned by the current user
func ListPersonalHomescripts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	homescriptList, err := database.ListHomescriptOfUser(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to list personal homescript", Error: "database failure"})
		return
	}
	json.NewEncoder(w).Encode(homescriptList)
}

// Creates a new Homescript
func CreateNewHomescript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request AddHomescriptRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	alreadyExists, err := database.DoesHomescriptExist(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add homescript", Error: "database failure"})
		return
	}
	if alreadyExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to add homescript", Error: fmt.Sprintf("the id: '%s' is already present in the database, use another one", request.Id)})
		return
	}
	homescriptToAdd := database.Homescript{
		Id:                  request.Id,
		Owner:               username,
		Name:                request.Name,
		Description:         request.Description,
		QuickActionsEnabled: request.QuickActionsEnabled,
		SchedulerEnabled:    request.SchedulerEnabled,
		Code:                request.Code,
	}
	if err := database.CreateNewHomescript(homescriptToAdd); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to create new homescript", Error: "database failure"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully created new homescript"})
}
