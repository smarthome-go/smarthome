package hardware

import "time"

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

// Whether a job daemon loop is already running
var daemonRunning bool

// Contains the queue for all pending jobs
var jobQueue []PowerJob = make([]PowerJob, 0)

// temporarely stores the result of each executed job
var jobResults []JobResult = make([]JobResult, 0)
var jobsWithErrorInHandlerCount uint16

// Time to be waited after each job (in milliseconds)
const cooldown = 500

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
	item := PowerJob{SwitchName: switchName, Power: turnOn, Id: id}
	jobQueue = append(jobQueue, item)
	if !daemonRunning {
		jobsWithErrorInHandlerCount = 0
		jobResults = make([]JobResult, 0)
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
		err := setPowerOnAllNodes(currentJob.SwitchName, currentJob.Power)
		jobResults = append(jobResults, JobResult{Id: currentJob.Id, Error: err})
		if err != nil {
			jobsWithErrorInHandlerCount += 1
		}
		// Removes the first value in the queue, in this case the freshly completed job
		var tempQueue []PowerJob = make([]PowerJob, 0)
		for i, item := range jobQueue {
			if i != 0 {
				tempQueue = append(tempQueue, item)
			}
		}
		jobQueue = tempQueue
		// Only sleep if other jobs are in the current queue
		if len(jobQueue) > 0 {
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

// Returns the number of currently pending jobs in the queue
func GetPendingJobCount() int {
	return len(jobQueue)
}

// Returns the number of registered failed jobs of the last running daemon (can also be the current daemon)
func GetJobsWithErrorInHandler() uint16 {
	return jobsWithErrorInHandlerCount
}

// Returns the current state of the job queue
func GetPendingJobs() []PowerJob {
	return jobQueue
}

// Returns the current state of the results queue
func GetResults() []JobResult {
	return jobResults
}
