package event

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Adds a log event to the database and prints it to the console
// Used by the other functions below
func logEvent(name string, description string, level database.LogLevel) error {
	err := database.AddLogEvent(
		name,
		description,
		level,
	)
	log.Trace(fmt.Sprintf("[EVENT](%d) %s: %s", level, name, description))
	if err != nil {
		log.Error("Could not log event: failed to communicate with database", err.Error())
		return err
	}
	return nil
}

func Trace(name string, description string) {
	if err := logEvent(name, description, database.LogLevelTrace); err != nil {
		log.Error("Failed to log trace event")
	}
}

func Debug(name string, description string) {
	if err := logEvent(name, description, database.LogLevelDebug); err != nil {
		log.Error("Failed to log debug event")
	}
}

func Info(name string, description string) {
	if err := logEvent(name, description, database.LogLevelInfo); err != nil {
		log.Error("Failed to log info event")
	}
}

func Warn(name string, description string) {
	if err := logEvent(name, description, database.LogLevelWarn); err != nil {
		log.Error("Failed to log warn event")
	}
}

func Error(name string, description string) {
	if err := logEvent(name, description, database.LogLevelError); err != nil {
		log.Error("Failed to log error event")
	}
}

func Fatal(name string, description string) {
	if err := logEvent(name, description, database.LogLevelFatal); err != nil {
		log.Error("Failed to log fatal event")
	}
}

func FlushOldLogs() error {
	log.Trace("Flushing logs older than 30 days...")
	return database.FlushOldLogs()
}

func FlushAllLogs() error {
	log.Trace("Flushing all logs...")
	return database.FlushAllLogs()
}
