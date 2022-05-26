package homescript

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/homescript/homescript"
	hmsError "github.com/smarthome-go/homescript/homescript/error"
	"github.com/smarthome-go/smarthome/core/database"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

type Location struct {
	Filename string `json:"filename"`
	Line     uint   `json:"line"`
	Column   uint   `json:"column"`
	Index    uint   `json:"index"`
}

type HomescriptError struct {
	ErrorType string   `json:"errorType"`
	Location  Location `json:"location"`
	Message   string   `json:"message"`
}

// Converts the error provided by the Homescript package to a Smarthome-specific, more user friendly struct
// Is required to allow prettier display of errors via the API
func convertError(errorItem hmsError.Error) HomescriptError {
	return HomescriptError{
		ErrorType: errorItem.TypeName,
		Location: Location{
			Filename: errorItem.Location.Filename,
			Line:     errorItem.Location.Line,
			Column:   errorItem.Location.Column,
			Index:    errorItem.Location.Index,
		},
		Message: errorItem.Message,
	}
}

// Takes multiple Homescript-errors and uses the `convertError` function on them individually
// Is used for preprocessing of possible errors before returning them to the client (via API)
func convertErrors(errorItems ...hmsError.Error) []HomescriptError {
	var outputErrors []HomescriptError
	for _, errorItem := range errorItems {
		outputErrors = append(outputErrors, convertError(errorItem))
	}
	return outputErrors
}

// Executes arbitrary Homescript-code as a given user, returns the output and a possible error slice
// The `scriptLabel` argument is used internally to allow for better error-display
// The `dryRun` argument specifies wheter the script should be linted or executed
// The `args` argument represents the arguments passed to the Homescript runtime and
// can be used from the script via the `CheckArg` and `GetArg` functions
func Run(username string, scriptLabel string, scriptCode string, dryRun bool, args map[string]string) (string, int, []HomescriptError) {
	executor := &Executor{
		Username:   username,
		ScriptName: scriptLabel,
		DryRun:     dryRun,
		Args:       args,
	}
	exitCode, runtimeErrors := homescript.Run(
		executor,
		scriptLabel,
		scriptCode,
	)
	if len(runtimeErrors) > 0 {
		log.Debug(fmt.Sprintf("Homescript '%s' ran by user '%s' has terminated: %s", scriptLabel, username, runtimeErrors[0].Message))
		return executor.Output, 1, convertErrors(runtimeErrors...)
	}
	log.Debug(fmt.Sprintf("Homescript '%s' ran by user '%s' was executed successfully", scriptLabel, username))
	return executor.Output, exitCode, make([]HomescriptError, 0)
}

// Executes a given Homescript from the database and returns it's output, exit-code and possible error
func RunById(username string, homescriptId string, dryRun bool, args map[string]string) (string, int, error) {
	homescriptItem, hasBeenFound, err := database.GetUserHomescriptById(homescriptId, username)
	if err != nil {
		return "database error", 500, err
	}
	if !hasBeenFound {
		return "not found error", 404, errors.New("Invalid Homescript id: no data associated with id")
	}
	output, exitCode, errorsHms := Run(username, homescriptItem.Data.Id, homescriptItem.Data.Code, dryRun, args)
	if len(errorsHms) > 0 {
		return "execution error", exitCode, fmt.Errorf("Homescript terminated with exit code %d: %s", exitCode, errorsHms[0].Message)
	}
	return output, exitCode, nil
}

// Checks whether a given Homescript has automations which rely on it
// Is used to decide whether a Homescript is safe to delete or not
func HasDependentAutomations(homescriptId string) (bool, error) {
	automations, err := database.GetAutomations()
	if err != nil {
		return false, err
	}
	for _, automation := range automations {
		if automation.Data.HomescriptId == homescriptId {
			return true, nil
		}
	}
	return false, nil
}
