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
// The `dryRun` argument specifies whether the script should be linted or executed
// The `args` argument represents the arguments passed to the Homescript runtime and
// can be used from the script via the `CheckArg` and `GetArg` functions
// The `excludedCalls` argument specifies which Homescripts may not be called by this Homescript in order to prevent recursion
func Old_Run(username string, scriptId string, scriptCode string, callStack []string, dryRun bool, args map[string]string) (string, int, []HomescriptError) {
	executor := &Executor{
		Username:   username,
		ScriptName: scriptId,
		DryRun:     dryRun,
		Args:       args,
		// Because the code cannot reference itself, the blacklist is left empty
		CallStack: callStack,
	}
	sigTerm := make(chan int)
	exitCode, hmsErrors := homescript.Run(
		executor,
		scriptId,
		scriptCode,
		&sigTerm,
	)
	if len(hmsErrors) > 0 {
		log.Debug(fmt.Sprintf("Homescript '%s' ran by user '%s' has terminated: %s", scriptId, username, hmsErrors[0].Message))
		return executor.Output, 1, convertErrors(hmsErrors...)
	}
	log.Debug(fmt.Sprintf("Homescript '%s' ran by user '%s' was executed successfully", scriptId, username))
	return executor.Output, exitCode, make([]HomescriptError, 0)
}

// Executes a given Homescript from the database and returns its output, exit-code and possible error
func Old_RunById(username string, scriptId string, callStack []string, dryRun bool, args map[string]string) (string, int, error) {
	homescriptItem, hasBeenFound, err := database.GetUserHomescriptById(scriptId, username)
	if err != nil {
		return "database error", 5, err
	}
	if !hasBeenFound {
		return "not found error", 4, errors.New("Invalid Homescript id: no data associated with id")
	}
	output, exitCode, hmsErrors := Old_Run(
		username,
		scriptId,
		homescriptItem.Data.Code,
		// The script's id is added to the blacklist
		append(callStack, scriptId),
		dryRun,
		args,
	)
	if len(hmsErrors) > 0 {
		return "execution error", exitCode, fmt.Errorf("Homescript terminated with exit code %d: %s", exitCode, hmsErrors[0].Message)
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
