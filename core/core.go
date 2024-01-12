package core

import (
	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/smarthome/core/config"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/drivers"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/core/user"
)

var log *logrus.Logger

// Initialize core loggers
func InitLoggers(logger *logrus.Logger) {
	log = logger

	config.InitLogger(log)
	drivers.InitLogger(log)
	homescript.InitLogger(log)
	database.InitLogger(log)
	hardware.InitLogger(log)
	event.InitLogger(log)
	user.InitLogger(log)
	log.Trace("Core loggers initialized")
}
