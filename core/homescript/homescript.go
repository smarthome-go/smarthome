package homescript

import (
	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
)

var RESERVED_IDS = []string{
	"sys",
}

var logger *logrus.Logger

func InitLogger(loggerTemp *logrus.Logger) {
	logger = loggerTemp
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
