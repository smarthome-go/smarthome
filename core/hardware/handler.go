package hardware

import "time"

// If a job daemon loop is already running
var daemonRunning bool

// Contains the queue for all pending jobs
var jobQueue []PowerJob

// temporarely stores the result of each executed job
var jobResults []JobResult

// Main interface for interacting with the queuing system
// Usage: SetPower("s1", true)
// Waits until all jobs are completed, can return an error
func SetPower(switchName string, turnOn bool) error {
	uniqueId := time.Now().UnixNano()
	addJobToQueue(switchName, turnOn, uniqueId)
	result := consumeResult(uniqueId)
	return result.Error
}

// Used for adding a job to a queue, keeps track of daemons and spawns them if needed
// Waits until the daemon quits, waiting for all (and the new) job(s) to be completed.
func addJobToQueue(switchName string, turnOn bool, id int64) {
	item := PowerJob{SwitchName: switchName, TurnOn: turnOn, Id: id}
	jobQueue = append(jobQueue, item)
	if !daemonRunning {
		daemonRunning = true
		ch := make(chan bool)
		go jobDaemon(ch)
		<-ch
	} else {
		for daemonRunning {
			time.Sleep(time.Second)
		}
	}
}

// Executes each job one after another
// Jobs can be added while the daemon is running
// If all jobs are completed, the daemon terminates itself in order to save ressources
// If a new job should be executed whilst no daemon is active, a new daemon is required.
func jobDaemon(ch chan bool) {
	for {
		if len(jobQueue) == 0 {
			daemonRunning = false
			ch <- true
			break
		}
		currentJob := jobQueue[0]
		err := setPowerOnAllNodes(currentJob.SwitchName, currentJob.TurnOn)
		jobResults = append(jobResults, JobResult{Id: currentJob.Id, Error: err})
		// Removes the first value in the queue, in this case the freshly completed job
		var tempQueue []PowerJob
		for i, item := range jobQueue {
			if i != 0 {
				tempQueue = append(tempQueue, item)
			}
		}
		jobQueue = tempQueue
		time.Sleep(500 * time.Millisecond)
	}
}

// This `garbage collector` consumes the result after it has been passed to the client
// After removing the wanted result from the slice, it is returned for further processing
func consumeResult(id int64) JobResult {
	var resultsTemp []JobResult
	var returnValue JobResult
	for _, result := range jobResults {
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