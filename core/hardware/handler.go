package hardware

import (
	"sync"
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

type jobQueueType struct {
	JobQueue []PowerJob
	m        sync.RWMutex
}

type resultQueueType struct {
	JobResults []JobResult
	m          sync.RWMutex
}

var jobQueue = jobQueueType{
	JobQueue: make([]PowerJob, 0),
}
var jobResults = resultQueueType{
	JobResults: make([]JobResult, 0),
}

// Whether a job daemon loop is already running
// var daemonRunning bool
var daemonRunning atomic.Value

// Contains the queue for all pending jobs
// var jobQueue []PowerJob = make([]PowerJob, 0)

// temporarely stores the result of each executed job
// var jobResults []JobResult = make([]JobResult, 0)
var jobsWithErrorInHandlerCount atomic.Value

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

	jobQueue.m.Lock()
	jobQueue.JobQueue = append(jobQueue.JobQueue, item)
	jobQueue.m.Unlock()

	if !daemonRunning.Load().(bool) {
		jobsWithErrorInHandlerCount.Store(0)

		jobResults.m.Lock()
		jobResults.JobResults = make([]JobResult, 0)
		jobResults.m.Unlock()

		daemonRunning.Store(true)
		ch := make(chan bool)
		go jobDaemon(ch)
		// TODO: Evaluate whether to replace with waitgroup
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
		jobQueue.m.RLock() // Lock for getting length
		length := len(jobQueue.JobQueue)
		jobQueue.m.RUnlock() // Unlock if condition
		if length == 0 {
			daemonRunning.Store(false)
			ch <- true
			break
		}

		jobQueue.m.RLock()
		currentJob := jobQueue.JobQueue[0]
		jobQueue.m.RUnlock()

		// Call the function which interacts with the hardware
		err := setPowerOnAllNodes(currentJob.Switch, currentJob.Power)

		jobResults.m.Lock()
		jobResults.JobResults = append(jobResults.JobResults, JobResult{Id: currentJob.Id, Error: err})
		jobResults.m.Unlock()

		// Increase `job-with-error count` if the job failed
		if err != nil {
			jobsWithErrorInHandlerCount.Store(jobsWithErrorInHandlerCount.Load().(int) + 1)
		}
		// Removes the first value in the queue, in this case the freshly completed job
		var tempQueue []PowerJob = make([]PowerJob, 0)

		jobQueue.m.RLock() // Lock while loop is iterating
		for i, item := range jobQueue.JobQueue {
			if i != 0 {
				tempQueue = append(tempQueue, item)
			}
		}
		jobQueue.m.RUnlock() // Unlock loop

		jobQueue.m.Lock()
		jobQueue.JobQueue = tempQueue
		jobQueue.m.Unlock()

		// Only sleep if other jobs are in the current queue
		jobQueue.m.RLock() // Lock for the if condition
		if len(jobQueue.JobQueue) > 0 {
			time.Sleep(cooldown * time.Millisecond)
		}
		jobQueue.m.RUnlock() // Unlock if condition
	}
}

// This `garbage collector` consumes the result after it has been passed to the client
// If a client cancels a request, the according response is not consumed. This response is cleared when a new handler is launched
// After removing the desired result from the slice, it is returned for further processing
func consumeResult(id int64) JobResult {
	var resultsTemp []JobResult = make([]JobResult, 0)
	var returnValue JobResult

	jobResults.m.Lock() // Lock for iteration
	for _, result := range jobResults.JobResults {
		if result.Id != id {
			resultsTemp = append(resultsTemp, result)
		} else {
			returnValue = result
		}
	}
	jobResults.JobResults = resultsTemp
	jobResults.m.Unlock()

	return returnValue
}

// Checks if the job with the current id has finished
func hasFinished(id int64) bool {
	jobResults.m.RLock()
	defer jobResults.m.RUnlock()
	for _, result := range jobResults.JobResults {
		if result.Id == id {
			return true
		}
	}
	return false
}

func Init() {
	// Needed for initializing atomics
	// Initialize thread-safe variables, for more info, look at the top for mutexes
	jobsWithErrorInHandlerCount.Store(0)
	daemonRunning.Store(false)
}

// Returns the number of currently pending jobs in the queue
func GetPendingJobCount() int {
	jobQueue.m.RLock()
	defer jobQueue.m.RUnlock()
	return len(jobQueue.JobQueue)
}

// Returns the number of registered failed jobs of the last running daemon (can also be the current daemon)
func GetJobsWithErrorInHandler() uint16 {
	return uint16(jobsWithErrorInHandlerCount.Load().(int))
}

// Returns the current state of the job queue
func GetPendingJobs() []PowerJob {
	jobQueue.m.RLock()
	defer jobQueue.m.RUnlock()
	return jobQueue.JobQueue
}

// Returns the current state of the results queue
func GetResults() []JobResult {
	jobResults.m.RLock()
	defer jobResults.m.RUnlock()
	return jobResults.JobResults
}
