package hardware

import (
	"github.com/sirupsen/logrus"
)

type PowerJob struct {
	Id         int64  `json:"id"`
	SwitchName string `json:"switchName"`
	Power      bool   `json:"power"`
}

type JobResult struct {
	Id    int64 `json:"id"`
	Error error `json:"error"`
}

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}
