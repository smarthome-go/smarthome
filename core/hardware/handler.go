package hardware

import (
	"sync/atomic"
	"time"
)

/*
Feature-spec of the handler:
- can handle async request, for example concurrent users or one user toggling power in the frontend fast
	- ideal for async job requests, like frontend
- Acts synchronous if one power job is awaited after the other
	- ideal for normal scripting

Time to complete:
(n) synchronous requests  -> n * repeats * 20 ms
(n) asynchronous requests -> n * (repeats * 20 ms + cooldown) - cooldown
*/

var jobQueue atomic.Value
var jobResults atomic.Value

// Whether a job daemon loop is already running
// var daemonRunning bool
var daemonRunning atomic.Value

// Contains the queue for all pending jobs
// var jobQueue []PowerJob = make([]PowerJob, 0)

// temporarely stores the result of each executed job
// var jobResults []JobResult = make([]JobResult, 0)
var jobsWithErrorInHandlerCount uint16

// Time to be waited after each job (in milliseconds)
const cooldown = 500

// Main interface for interacting with the queuing system
// Usage: SetPower("s1", true)
// Waits until all jobs are completed, can return an error
func SetPower(switchName string, powerOn bool) error {
	uniqueId := time.Now().UnixNano()
	addJobToQueue(switchName, powerOn, uniqueId)
	result := consumeResult(uniqueId)
	return result.Error
}

// Used for adding a job to a queue, keeps track of daemons and spawns them if needed
// Waits until the daemon quits, waiting for all (and the new) job(s) to be completed.
func addJobToQueue(switchId string, turnOn bool, id int64) {
	item := PowerJob{Switch: switchId, Power: turnOn, Id: id}
	jobQueue.Store(append(jobQueue.Load().([]PowerJob), item))
	if !daemonRunning.Load().(bool) {
		jobsWithErrorInHandlerCount = 0
		jobResults.Store(make([]JobResult, 0))
		daemonRunning.Store(true)
		ch := make(chan bool)
		go jobDaemon(ch)
		for {
			select {
			case <-ch:
				return
			default:
				time.Sleep(time.Millisecond * 50)
				if hasFinished(id) {
					return
				}
			}
		}
	} else {
		for daemonRunning.Load().(bool) && !hasFinished(id) {
			time.Sleep(time.Millisecond * 50)
		}
	}
}

// Executes each job one after another
// Jobs can be added while the daemon is running
// If all jobs are completed, the daemon terminates itself in order to save resources
// If a new job should be executed whilst no daemon is active, a new daemon is created.
func jobDaemon(ch chan bool) {
	for {
		if len(jobQueue.Load().([]PowerJob)) == 0 {
			daemonRunning.Store(false)
			ch <- true
			break
		}
		currentJob := jobQueue.Load().([]PowerJob)[0]
		err := setPowerOnAllNodes(currentJob.Switch, currentJob.Power)
		jobResults.Store(append(jobResults.Load().([]JobResult), JobResult{Id: currentJob.Id, Error: err}))
		if err != nil {
			jobsWithErrorInHandlerCount += 1
		}
		// Removes the first value in the queue, in this case the freshly completed job
		var tempQueue []PowerJob = make([]PowerJob, 0)
		for i, item := range jobQueue.Load().([]PowerJob) {
			if i != 0 {
				tempQueue = append(tempQueue, item)
			}
		}
		jobQueue.Store(tempQueue)
		// Only sleep if other jobs are in the current queue
		if len(jobQueue.Load().([]PowerJob)) > 0 {
			time.Sleep(cooldown * time.Millisecond)
		}
	}
}

// This `garbage collector` consumes the result after it has been passed to the client
// If a client cancels a request, the according response is not consumed. This response is cleared when a new handler is launched
// After removing the desired result from the slice, it is returned for further processing
func consumeResult(id int64) JobResult {
	var resultsTemp []JobResult = make([]JobResult, 0)
	var returnValue JobResult
	for _, result := range jobResults.Load().([]JobResult) {
		if result.Id != id {
			resultsTemp = append(resultsTemp, result)
		} else {
			returnValue = result
		}
	}
	jobResults.Store(resultsTemp)
	return returnValue
}

// Checks if the job with the current id has finished
func hasFinished(id int64) bool {
	for _, result := range jobResults.Load().([]JobResult) {
		if result.Id == id {
			return true
		}
	}
	return false
}

func Init() {
	// Initialize thread-safe slice
	jobQueue.Store(make([]PowerJob, 0))
	jobResults.Store(make([]JobResult, 0))
	daemonRunning.Store(false)
}

// Returns the number of currently pending jobs in the queue
func GetPendingJobCount() int {
	queue := jobQueue.Load().([]PowerJob)
	return len(queue)
}

// Returns the number of registered failed jobs of the last running daemon (can also be the current daemon)
func GetJobsWithErrorInHandler() uint16 {
	return jobsWithErrorInHandlerCount
}

// Returns the current state of the job queue
func GetPendingJobs() []PowerJob {
	queue := jobQueue.Load().([]PowerJob)
	return queue
}

// Returns the current state of the results queue
func GetResults() []JobResult {
	return jobResults.Load().([]JobResult)
}
