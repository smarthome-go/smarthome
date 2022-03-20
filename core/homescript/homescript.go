package homescript

import (
	"fmt"

	"github.com/MikMuellerDev/homescript/homescript"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Executes a given homescript as a given user, returns the output and a possible error
func Run(username string, scriptLabel string, scriptCode string) (string, error) {
	output, err := homescript.Run(
		Executor{
			Username:   username,
			ScriptName: scriptLabel,
		},
		scriptCode,
	)
	if err != nil {
		log.Error(fmt.Sprintf("Homescript '%s' ran by user '%s' has terminated: %s", scriptLabel, username, err.Error()))
		return output, err
	}
	log.Info(fmt.Sprintf("Homescript '%s' ran by user '%s' was executed successfully", scriptLabel, username))
	return output, nil
}
