package event

import (
	"fmt"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func logEvent(name string, description string, level int) error {
	err := database.AddLogEvent(name, description, level)
	log.Trace(fmt.Printf("[EVENT](%d) %s: %s", level, name, description))
	if err != nil {
		log.Error("Could not log event: failed to communicate with database", err.Error())
		return err
	}
	return nil
}

func Trace(name string, description string) {
	logEvent(name, description, 0)
}

func Debug(name string, description string) {
	logEvent(name, description, 1)
}

func Info(name string, description string) {
	logEvent(name, description, 2)
}

func Warn(name string, description string) {
	logEvent(name, description, 3)
}

func Error(name string, description string) {
	logEvent(name, description, 4)
}

func Fatal(name string, description string) {
	logEvent(name, description, 5)
}
