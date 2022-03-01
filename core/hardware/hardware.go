package hardware

import (
	"fmt"
	"time"

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
	SwitchName string
	TurnOn     bool
}

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Specifies if a job handling loop is already running
var handlerRunning bool

// Contains the queue for all pending jobs
var jobQueue []PowerJob

func ExecuteJob(switchName string, turnOn bool) bool {
	item := PowerJob{SwitchName: switchName, TurnOn: turnOn}
	jobQueue = append(jobQueue, item)
	if !handlerRunning {
		handlerRunning = true
		ch := make(chan bool)
		go jobHandler(ch)
		fmt.Println("Handler started.")
		success := <-ch
		fmt.Println("Handler Stopped, Success:", success)
		return true
	} else {
		for handlerRunning {
			time.Sleep(time.Second)
		}
		fmt.Println("Handler Stop DETECTED, Success:")
		return true
	}
}

func jobHandler(ch chan bool) {
	for {
		if len(jobQueue) == 0 {
			handlerRunning = false
			fmt.Println("Handler Stopped")
			ch <- true
			break
		}
		for i := 0; i < len(jobQueue); i++ {
			fmt.Printf("%s ", jobQueue[i].SwitchName)
		}
		currentJob := jobQueue[0]
		setPowerOnAllNodes(currentJob.SwitchName, currentJob.TurnOn)
		jobQueue = RemoveJobByIndex(jobQueue, 0)
		time.Sleep(500 * time.Millisecond)
	}
}

func RemoveJobByIndex(s []PowerJob, index int) []PowerJob {
	var temp []PowerJob
	for i, item := range jobQueue {
		if i != index {
			temp = append(temp, item)
		}
	}
	return temp
}

func GetPendingJobs() int {
	return len(jobQueue)
}
