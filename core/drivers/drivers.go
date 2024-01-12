package drivers

import "github.com/sirupsen/logrus"

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}
