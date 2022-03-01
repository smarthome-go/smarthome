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

// Specifies if a job handling loop is already running
var handlerRunning bool

// Contains the queue for all pending jobs
var jobQueue []PowerJob

// temporarely stores the result of each executed job
var jobResults []JobResult

// TODO: write an actual documentation about the following code
func SetPower(switchName string, turnOn bool) error {
	uniqueId := time.Now().UnixNano()
	ExecuteJob(switchName, turnOn, uniqueId)
	result := consumeResult(uniqueId)
	return result.Error
}

func ExecuteJob(switchName string, turnOn bool, id int64) {
	item := PowerJob{SwitchName: switchName, TurnOn: turnOn, Id: id}
	jobQueue = append(jobQueue, item)
	if !handlerRunning {
		handlerRunning = true
		ch := make(chan bool)
		go jobHandler(ch)
		<-ch
	} else {
		for handlerRunning {
			time.Sleep(time.Second)
		}
	}
}

func jobHandler(ch chan bool) {
	for {
		if len(jobQueue) == 0 {
			handlerRunning = false
			ch <- true
			break
		}
		currentJob := jobQueue[0]
		err := setPowerOnAllNodes(currentJob.SwitchName, currentJob.TurnOn)
		jobResults = append(jobResults, JobResult{Id: currentJob.Id, Error: err})
		// TODO: REMOVE PRINT
		fmt.Printf("Jobs: ")
		fmt.Println(jobQueue)
		fmt.Printf("Results: ")
		fmt.Println(jobResults)
		//
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

func consumeResult(id int64) JobResult {
	var resultsTemp []JobResult
	var returnValue JobResult
	for _, result := range jobResults {
		// `the garbage collector` consumes the result after it has been passed to the client
		if result.Id != id {
			resultsTemp = append(resultsTemp, result)
		} else {
			returnValue = result
		}
	}
	jobResults = resultsTemp
	return returnValue
}

func GetPendingJobs() int {
	return len(jobQueue)
}
