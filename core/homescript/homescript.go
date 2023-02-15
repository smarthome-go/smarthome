package homescript

import (
	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/automation"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
	automation.InitLogger(logger)
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
