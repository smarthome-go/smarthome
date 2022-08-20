package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gorilla/mux"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type HomescriptResponse struct {
	Success  bool                         `json:"success"`
	Id       string                       `json:"id"`
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
	Workspace           string `json:"workspace"`
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

// Runs any given Homescript given its id
func RunHomescriptId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request HomescriptIdRunRequest
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
	hmsData, found, err := database.GetUserHomescriptById(request.Id, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "could not retrieve Homescript from database", Error: "database failure"})
		return
	}
	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "Homescript not found", Error: "no data associated with id"})
		return
	}
	// Run the Homescript
	output, exitCode, _, hmsErrors := homescript.HmsManager.Run(
		username,
		request.Id,
		hmsData.Data.Code,
		false,
		args,
		make([]string, 0),
		homescript.InitiatorAPI,
		make(chan int),
	)
	if len(hmsErrors) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(
			HomescriptResponse{
				Success:  false,
				Id:       request.Id,
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
				Id:       request.Id,
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
			Id:       request.Id,
			Message:  "Homescript ran successfully",
			Output:   output,
			Exitcode: exitCode,
			Errors:   hmsErrors,
		}); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "could not encode response", Error: "could not encode response"})
	}
}

// Lints any given Homescript given its id
func LintHomescriptId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request HomescriptIdRunRequest
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
	hmsData, found, err := database.GetUserHomescriptById(request.Id, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "Could not retrieve Homescript from database", Error: "database failure"})
		return
	}
	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "Homescript not found", Error: "no data associated with id"})
		return
	}
	// Lint the Homescript
	output, exitCode, _, hmsErrors := homescript.HmsManager.Run(
		username,
		request.Id,
		hmsData.Data.Code,
		true,
		args,
		make([]string, 0),
		homescript.InitiatorAPI,
		make(chan int),
	)
	if len(hmsErrors) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(
			HomescriptResponse{
				Success:  false,
				Id:       request.Id,
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
				Id:       request.Id,
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
			Id:       request.Id,
			Message:  "Linting discovered no errors",
			Output:   output,
			Exitcode: exitCode,
			Errors:   hmsErrors,
		}); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "could not encode response", Error: "could not encode response"})
	}
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
	output, exitCode, _, hmsErrors := homescript.HmsManager.Run(
		username,
		"live",
		request.Code,
		false,
		args,
		make([]string, 0),
		homescript.InitiatorAPI,
		make(chan int),
	)
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
	output, exitCode, _, hmsErrors := homescript.HmsManager.Run(
		username,
		"lint",
		request.Code,
		true,
		args,
		make([]string, 0),
		homescript.InitiatorAPI,
		make(chan int),
	)
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
		Res(w, Response{Success: false, Message: "failed to list personal Homescripts", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(homescriptList); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to list personal Homescripts", Error: "could not encode response"})
	}
}

// Returns a list of Homescripts which are owned by the current user
// Additionally, each Homescript also contains its arguments
func ListPersonalHomescriptsWithArgs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	homescriptList, err := homescript.ListPersonalHomescriptWithArgs(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to list personal Homescripts with arguments", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(homescriptList); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to list personal Homescripts with arguments", Error: "could not encode response"})
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
	if strings.Contains(request.Id, " ") || utf8.RuneCountInString(request.Id) > 30 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add Homescript", Error: fmt.Sprintf("the id: '%s' must not exceed 30 characters and must not include any whitespaces", request.Id)})
		return
	}
	if utf8.RuneCountInString(request.Workspace) > 50 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add Homescript", Error: fmt.Sprintf("the workspace: '%s' must not exceed 50 characters", request.Workspace)})
		return
	}
	if utf8.RuneCountInString(request.MDIcon) > 100 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add Homescript", Error: fmt.Sprintf("the mdIcon: '%s' must not exceed 100 characters", request.MDIcon)})
		return
	}
	if utf8.RuneCountInString(request.Name) > 30 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add Homescript", Error: fmt.Sprintf("the name: '%s' must not exceed 30 characters", request.Name)})
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
			Workspace:           request.Workspace,
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

// Kills a Homescript job given its id
func KillJobById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to kill Homescript job", Error: "no id provided"})
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to kill Homescript job", Error: "id must be numeric"})
		return
	}
	job, found := homescript.HmsManager.GetJobById(uint64(idInt))
	if !found || job.Executor.Username != username {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to kill Homescript job", Error: "invalid id provided"})
		return
	}
	_ = homescript.HmsManager.Kill(uint64(idInt))
	Res(w, Response{Success: true, Message: "successfully killed Homescript job"})
}

// Kills all jobs executing an arbitrary Homescript id
func KillAllHMSIdJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to kill Homescript job", Error: "no id provided"})
		return
	}
	// Validate that the user is allowed to kill the requested script id
	_, valid, err := database.GetUserHomescriptById(id, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "could not retrieve Homescript validation from database", Error: "database failure"})
		return
	}
	if !valid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not kill all Homescript jobs", Error: "invalid Homescript id specified"})
		return
	}
	count, _ := homescript.HmsManager.KillAllId(id)
	Res(w, Response{Success: true, Message: fmt.Sprintf("successfully killed %d Homescript job(s)", count)})
}

// Returns a list of currently running HMS jobs
func GetHMSJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	jobs := homescript.HmsManager.GetUserDirectJobs(username)
	if err := json.NewEncoder(w).Encode(jobs); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to list Homescript jobs", Error: "could not encode response"})
	}
}
