package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gorilla/mux"

	"github.com/smarthome-go/smarthome/core"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type HomescriptResponse struct {
	Success      bool                  `json:"success"`
	Output       string                `json:"output"`
	FileContents map[string]string     `json:"fileContents"`
	Errors       []homescript.HmsError `json:"errors"`
}

type CreateHomescriptRequest struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	QuickActionsEnabled bool   `json:"quickActionsEnabled"`
	IsWidget            bool   `json:"isWidget"`
	SchedulerEnabled    bool   `json:"schedulerEnabled"`
	Code                string `json:"code"`
	MDIcon              string `json:"mdIcon"`
	Workspace           string `json:"workspace"`
}

type ModifyCodeRequest struct {
	Id   string `json:"id"`
	Code string `json:"code"`
}

type HomescriptArg struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type HomescriptLiveRunRequest struct {
	Code string          `json:"code"`
	Args []HomescriptArg `json:"args"`
}

type LintHomescriptStringRequest struct {
	Code       string          `json:"code"`
	Args       []HomescriptArg `json:"args"`
	ModuleName string          `json:"moduleName"`
	IsDriver   bool            `json:"isDriver"`
}

type HomescriptIdRunRequest struct {
	Id       string          `json:"id"`
	Args     []HomescriptArg `json:"args"`
	IsWidget bool            `json:"isWidget"`
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

	hmsData, found, err := homescript.GetPersonalScriptById(request.Id, username)
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

	initiator := homescript.InitiatorAPI
	if request.IsWidget {
		initiator = homescript.InitiatorWidget
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Run the Homescript
	var outputBuffer bytes.Buffer
	res, err := homescript.HmsManager.Run(
		homescript.HMS_PROGRAM_KIND_NORMAL,
		nil,
		username,
		&request.Id,
		hmsData.Data.Code,
		initiator,
		ctx,
		cancel,
		nil,
		args,
		&outputBuffer,
		nil,
	)

	// TODO: check error

	if err := json.NewEncoder(w).Encode(
		HomescriptResponse{
			Success:      res.Success,
			Output:       outputBuffer.String(),
			FileContents: res.FileContents,
			Errors:       res.Errors,
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
	hmsData, found, err := homescript.GetPersonalScriptById(request.Id, username)
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

	programKind := homescript.HMS_PROGRAM_KIND_NORMAL
	if hmsData.Data.Type == database.HOMESCRIPT_TYPE_DRIVER {
		programKind = homescript.HMS_PROGRAM_KIND_DEVICE_DRIVER
	}

	// Lint the Homescript
	_, res, err := homescript.HmsManager.Analyze(
		username,
		request.Id,
		hmsData.Data.Code,
		programKind,
		nil,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "Could not analyze Homescript", Error: "internal failure"})
		return
	}

	if err := json.NewEncoder(w).Encode(
		HomescriptResponse{
			Success:      res.Success,
			FileContents: res.FileContents,
			Errors:       res.Errors,
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

	ctx, cancel := context.WithCancel(context.Background())

	// Run the Homescript
	var outputBuffer bytes.Buffer
	res, err := homescript.HmsManager.Run(
		homescript.HMS_PROGRAM_KIND_NORMAL,
		nil,
		username,
		nil,
		request.Code,
		homescript.InitiatorAPI,
		ctx,
		cancel,
		nil,
		args,
		&outputBuffer,
		nil,
	)
	output := outputBuffer.String()
	// if len(output) > 100_000 {
	// 	output = "Output too large"
	// } TODO: maybe re-introduce such a limit

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "could not run Homescript", Error: "database failure"})
		return
	}

	if err := json.NewEncoder(w).Encode(
		HomescriptResponse{
			Success:      res.Success,
			Output:       output,
			Errors:       res.Errors,
			FileContents: res.FileContents,
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
	var request LintHomescriptStringRequest
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

	programKind := homescript.HMS_PROGRAM_KIND_NORMAL
	var driverMetadata *homescript.AnalyzerDriverMetadata = nil

	if request.IsDriver {
		programKind = homescript.HMS_PROGRAM_KIND_DEVICE_DRIVER

		_, validationErr, databaseErr := homescript.DriverFromHmsId(request.ModuleName)
		if databaseErr != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			Res(w, Response{Success: false, Message: "could not lint Homescript string", Error: "database error"})
			return
		}

		if validationErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			Res(w, Response{Success: false, Message: "could not lint Homescript string", Error: fmt.Sprintf("validation error: %s", validationErr.Error())})
			return
		}

		driverMetadata = &homescript.AnalyzerDriverMetadata{
			// VendorId: driverData.VendorId,
			// ModelId:  driverData.ModelId,
		}
	}

	// Lint the Homescript
	_, res, err := homescript.HmsManager.Analyze(
		username,
		request.ModuleName,
		request.Code,
		programKind,
		driverMetadata,
	)

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to lint Homescript string", Error: "internal server error"})
		return
	}

	if err := json.NewEncoder(w).Encode(
		HomescriptResponse{
			Success:      res.Success,
			Errors:       res.Errors,
			FileContents: res.FileContents,
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
	homescriptList, err := homescript.ListPersonal(username)
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

	// Check if this id is reserved
	for _, id := range homescript.RESERVED_IDS {
		if id == request.Id {
			w.WriteHeader(http.StatusUnprocessableEntity)
			Res(w, Response{Success: false, Message: "failed to add Homescript", Error: fmt.Sprintf("the id: '%s' is reserved, use another one", request.Id)})
			return
		}
	}

	alreadyExists, err := database.DoesHomescriptExist(request.Id, username)
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
	if strings.Contains(request.Id, " ") || utf8.RuneCountInString(request.Id) > database.HOMESCRIPT_ID_LEN {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add Homescript", Error: fmt.Sprintf("the id: '%s' must not exceed %d characters and must not include any whitespaces", request.Id, database.HOMESCRIPT_ID_LEN)})
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
	if request.IsWidget && (request.SchedulerEnabled || request.QuickActionsEnabled) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add Homescript", Error: "cannot use script in scheduler or as quick-action if it is a widget"})
		return
	}
	homescriptToAdd := database.Homescript{
		Owner: username,
		Data: database.HomescriptData{
			Id:                  request.Id,
			Name:                request.Name,
			Description:         request.Description,
			QuickActionsEnabled: request.QuickActionsEnabled,
			IsWidget:            request.IsWidget,
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
	_, exists, err := homescript.GetPersonalScriptById(request.Id, username)
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
	if err := database.DeleteHomescriptById(request.Id, username); err != nil {
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
	_, exists, err := homescript.GetPersonalScriptById(request.Id, username)
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
	newHmsData := database.HomescriptData{
		Name:                request.Name,
		Description:         request.Description,
		QuickActionsEnabled: request.QuickActionsEnabled,
		SchedulerEnabled:    request.SchedulerEnabled,
		IsWidget:            request.IsWidget,
		Code:                request.Code,
		MDIcon:              request.MDIcon,
		Workspace:           request.Workspace,
	}
	if err := database.ModifyHomescriptById(
		request.Id,
		username,
		newHmsData,
	); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify Homescript", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully modified Homescript"})
}

// Modifies the code of a given Homescript
func ModifyHomescriptCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request ModifyCodeRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	found, err := core.ModifyHomescriptCode(
		request.Id,
		username,
		request.Code,
	)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify Homescript code", Error: "database failure"})
		return
	}

	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify Homescript code", Error: "no such script found"})
		return
	}

	Res(w, Response{Success: true, Message: "successfully modified Homescript code"})
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
	homescript, exists, err := homescript.GetPersonalScriptById(id, username)
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
	if !found || job.Username != username {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to kill Homescript job", Error: "invalid id provided"})
		return
	}
	if homescript.HmsManager.Kill(uint64(idInt)) {
		Res(w, Response{Success: true, Message: "successfully killed Homescript job"})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to kill Homescript job", Error: "internal error"})
	}
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
	_, valid, err := homescript.GetPersonalScriptById(id, username)
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
