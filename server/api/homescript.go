package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type HomescriptResponse struct {
	Success  bool                         `json:"success"`
	Exitcode int                          `json:"exitCode"`
	Message  string                       `json:"message"`
	Output   string                       `json:"output"`
	Errors   []homescript.HomescriptError `json:"error"`
}

type CreateHomescriptRequest struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	QuickActionsEnabled bool   `json:"quickActionsEnabled"`
	SchedulerEnabled    bool   `json:"schedulerEnabled"`
	Code                string `json:"code"`
	MDIcon              string `json:"mdIcon"`
}

type HomescriptArg struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type HomescriptLiveRunRequest struct {
	Code string          `json:"code"`
	Args []HomescriptArg `json:"args"`
}

type HomescriptIdRunRequest struct {
	Id   string          `json:"id"`
	Args []HomescriptArg `json:"args"`
}

type HomescriptIdRequest struct {
	Id string `json:"id"`
}

// Runs any given Homescript as a string
func RunHomescriptString(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request HomescriptLiveRunRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Fill the arguments using the request
	args := make(map[string]string, 0)
	for _, arg := range request.Args {
		args[arg.Key] = arg.Value
	}
	// Run the Homescript
	output, exitCode, hmsErrors := homescript.Run(username, "live", request.Code, false, args)
	if len(hmsErrors) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(
			HomescriptResponse{
				Success:  false,
				Exitcode: exitCode,
				Message:  "Homescript terminated abnormally",
				Output:   output,
				Errors:   hmsErrors,
			}); err != nil {
			log.Error(err.Error())
			Res(w, Response{Success: false, Message: "could not encode response", Error: "could not encode response"})
		}
		return
	}
	if exitCode != 0 {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(
			HomescriptResponse{
				Success:  false,
				Exitcode: exitCode,
				Message:  "Homescript exited with a non-0 status code",
				Output:   output,
				Errors:   hmsErrors,
			}); err != nil {
			log.Error(err.Error())
			Res(w, Response{Success: false, Message: "could not encode response", Error: "could not encode response"})
		}
		return
	}
	if err := json.NewEncoder(w).Encode(
		HomescriptResponse{
			Success:  true,
			Message:  "Homescript ran successfully",
			Output:   output,
			Exitcode: exitCode,
			Errors:   hmsErrors,
		}); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "could not encode response", Error: "could not encode response"})
	}
}

// Lints a given Homescript string and checks it for errors
func LintHomescriptString(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request HomescriptLiveRunRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Fill the arguments using the request
	args := make(map[string]string, 0)
	for _, arg := range request.Args {
		args[arg.Key] = arg.Value
	}
	// Lint the Homescript
	output, exitCode, hmsErrors := homescript.Run(username, "lint", request.Code, true, args)
	if len(hmsErrors) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(
			HomescriptResponse{
				Success:  false,
				Exitcode: exitCode,
				Message:  "Linting discovered errors",
				Output:   output,
				Errors:   hmsErrors,
			}); err != nil {
			log.Error(err.Error())
			Res(w, Response{Success: false, Message: "could not encode response", Error: "could not encode response"})
		}
		return
	}
	if exitCode != 0 {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(
			HomescriptResponse{
				Success:  false,
				Exitcode: exitCode,
				Message:  "Linting exited with non-0 status code but ran successfully",
				Output:   output,
				Errors:   hmsErrors,
			}); err != nil {
			log.Error(err.Error())
			Res(w, Response{Success: false, Message: "could not encode response", Error: "could not encode response"})
		}
		return
	}
	if err := json.NewEncoder(w).Encode(
		HomescriptResponse{
			Success:  true,
			Message:  "Linting discovered no errors",
			Output:   output,
			Exitcode: exitCode,
			Errors:   hmsErrors,
		}); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "could not encode response", Error: "could not encode response"})
	}
}

// Returns a list of Homescripts which are owned by the current user
func ListPersonalHomescripts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	homescriptList, err := database.ListHomescriptOfUser(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to list personal Homescript", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(homescriptList); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to list personal Homescript", Error: "could not encode response"})
	}
}

// Creates a new Homescript
func CreateNewHomescript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request CreateHomescriptRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	alreadyExists, err := database.DoesHomescriptExist(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add Homescript", Error: "database failure"})
		return
	}
	if alreadyExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add Homescript", Error: fmt.Sprintf("the id: '%s' is already present in the database, use another one", request.Id)})
		return
	}
	homescriptToAdd := database.Homescript{
		Owner: username,
		Data: database.HomescriptData{
			Id:                  request.Id,
			Name:                request.Name,
			Description:         request.Description,
			QuickActionsEnabled: request.QuickActionsEnabled,
			SchedulerEnabled:    request.SchedulerEnabled,
			Code:                request.Code,
			MDIcon:              request.MDIcon,
		},
	}
	if err := database.CreateNewHomescript(homescriptToAdd); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create new Homescript", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully created new Homescript"})
}

// Deletes a Homescript by its Id, checks if it exists and if the user has permission to delete it
func DeleteHomescriptById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request HomescriptIdRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, exists, err := database.GetUserHomescriptById(request.Id, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete Homescript: could not validate existence", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete Homescript", Error: "not found / permission denied: no data is associated to this id"})
		return
	}
	hasDependentAutomations, err := homescript.HasDependentAutomations(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete Homescript: could not validate deletion safety", Error: "database failure"})
		return
	}
	if hasDependentAutomations {
		w.WriteHeader(http.StatusConflict)
		Res(w, Response{Success: false, Message: "can not delete Homescript: safety violation", Error: "script is used in one or more automations"})
		return
	}
	if err := database.DeleteHomescriptById(request.Id); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete Homescript", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted Homescript"})
}

// Modifies the metadata of a given Homescript
func ModifyHomescript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request CreateHomescriptRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, exists, err := database.GetUserHomescriptById(request.Id, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify Homescript: could not validate existence", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify Homescript", Error: "not found / permission denied: no data is associated to this id"})
		return
	}
	homescriptMetadata := database.HomescriptData{
		Name:                request.Name,
		Description:         request.Description,
		QuickActionsEnabled: request.QuickActionsEnabled,
		SchedulerEnabled:    request.SchedulerEnabled,
		Code:                request.Code,
		MDIcon:              request.MDIcon,
	}
	if err := database.ModifyHomescriptById(request.Id, homescriptMetadata); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify Homescript", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully modified Homescript"})
}

// Returns the metadata of an arbitrary Homescript-id to which the user has access to
func GetUserHomescriptById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to get Homescript by id", Error: "no id provided"})
		return
	}
	homescript, exists, err := database.GetUserHomescriptById(id, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to get Homescript by id", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to get Homescript by id", Error: "invalid id: no such Homescript exists"})
		return
	}
	if err := json.NewEncoder(w).Encode(homescript); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to list personal Homescript", Error: "could not encode response"})
	}
}
