package utils

import (
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func NewLogger(logLevel logrus.Level) (*logrus.Logger, error) {
	// Create new logger
	logger := logrus.New()
	logger.SetLevel(logLevel)

	// Add filesystem hook in order to log to files
	pathMap := lfshook.PathMap{
		logrus.InfoLevel:  "./log/application.log",
		logrus.WarnLevel:  "./log/application.log",
		logrus.ErrorLevel: "./log/error.log",
		logrus.FatalLevel: "./log/error.log",
	}
	var hook *lfshook.LfsHook = lfshook.NewHook(
		pathMap,
		&logrus.JSONFormatter{PrettyPrint: false},
	)
	logger.Hooks.Add(hook)
	return logger, nil
}
