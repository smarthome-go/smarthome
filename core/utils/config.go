package utils

import "github.com/sirupsen/logrus"

var (
	Version string
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}
