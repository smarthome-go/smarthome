package homescript

import (
	"errors"
	"fmt"

	"github.com/MikMuellerDev/homescript/homescript"
	hmsError "github.com/MikMuellerDev/homescript/homescript/error"
	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/sirupsen/logrus"
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

func convertErrors(errorItems ...hmsError.Error) []HomescriptError {
	var outputErrors []HomescriptError
	for _, errorItem := range errorItems {
		outputErrors = append(outputErrors, convertError(errorItem))
	}
	return outputErrors
}

// Executes a given homescript as a given user, returns the output and a possible error slice
func Run(username string, scriptLabel string, scriptCode string) (string, int, []HomescriptError) {
	executor := &Executor{
		Username:   username,
		ScriptName: scriptLabel,
	}
	exitCode, runtimeErrors := homescript.Run(
		executor,
		scriptLabel,
		scriptCode,
	)
	if len(runtimeErrors) > 0 {
		log.Error(fmt.Sprintf("Homescript '%s' ran by user '%s' has terminated: %s", scriptLabel, username, runtimeErrors[0].Message))
		return executor.Output, 1, convertErrors(runtimeErrors...)
	}
	log.Info(fmt.Sprintf("Homescript '%s' ran by user '%s' was executed successfully", scriptLabel, username))
	return executor.Output, exitCode, make([]HomescriptError, 0)
}

func RunById(username string, homescriptId string) (string, int, error) {
	homescriptItem, hasBeenFound, err := database.GetUserHomescriptById(homescriptId, username)
	if err != nil {
		return "database error", 500, err
	}
	if !hasBeenFound {
		return "not found error", 404, errors.New("Invalid Homescript id: no data associated with id")
	}
	output, exitCode, errorsHms := Run(username, homescriptItem.Id, homescriptItem.Code)
	if len(errorsHms) > 0 {
		return "execution error", exitCode, fmt.Errorf("Homescript terminated with exit code %d: %s", exitCode, errorsHms[0].Message)
	}
	return output, exitCode, nil
}
