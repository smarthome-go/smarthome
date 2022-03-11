package camera

import (
	"github.com/sirupsen/logrus"
)

// Can be adjusted to define a maximum image size
// Between 0 and 255 Megabytes
const maxImageSize uint8 = 10

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// TODO: add access to the cameras which are in the database
