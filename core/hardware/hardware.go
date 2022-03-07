package hardware

import (
	"github.com/sirupsen/logrus"
)

type Node struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Token string `json:"token"`
}

type HardwareConfig struct {
	Nodes []Node `json:"nodes"`
}

type PowerJob struct {
	Id         int64
	SwitchName string
	TurnOn     bool
}

type JobResult struct {
	Id    int64
	Error error
}

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}
