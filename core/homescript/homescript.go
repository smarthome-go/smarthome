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

// Executes a given homescript as a given user, returns the output and a possible error slice
func Run(username string, scriptLabel string, scriptCode string) (string, []error) {
	executor := &Executor{
		Username:   username,
		ScriptName: scriptLabel,
	}
	err := homescript.Run(
		executor,
		scriptCode,
	)
	if err != nil && len(err) > 0 {
		log.Error(fmt.Sprintf("Homescript '%s' ran by user '%s' has terminated: %s", scriptLabel, username, err[0].Error()))
		return executor.Output, err
	}
	log.Info(fmt.Sprintf("Homescript '%s' ran by user '%s' was executed successfully", scriptLabel, username))
	return executor.Output, nil
}
