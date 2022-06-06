package homescript

import (
	"errors"
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
)

// Business logic for Homescript arguments goes here

// Executes a given Homescript from the database and returns its output, exit-code and possible error
func RunHomescriptByIdWithRequiredArgs(username string, homescriptId string, dryRun bool, args map[string]string) (string, int, error) {
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
