package homescript

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/smarthome-go/homescript/homescript"
	"github.com/smarthome-go/smarthome/core/database"
)

type HomescriptInitiator string

var (
	InitiatorAutomation HomescriptInitiator = "automation"
	InitiatorScheduler  HomescriptInitiator = "scheduler"
	InitiatorExec       HomescriptInitiator = "exec_target"
	InitiatorInternal   HomescriptInitiator = "internal"
	InitiatorAPI        HomescriptInitiator = "api"
)

// Global manager
var HmsManager Manager

// Initializes the Homescript manager
func InitManager() {
	HmsManager = Manager{
		Lock: sync.RWMutex{},
		Jobs: make([]Job, 0),
	}
}

type Manager struct {
	Lock sync.RWMutex
	Jobs []Job
}

type Job struct {
	Id        uint64              `json:"id"`
	Initiator HomescriptInitiator `json:"initiator"`
	Executor  *Executor           `json:"executor"`
}

func (m *Manager) PushJob(
	executor *Executor,
	initiator HomescriptInitiator,
	idReceiver chan uint64,
) uint64 {
	m.Lock.Lock()
	id := uint64(len(m.Jobs))
	m.Jobs = append(m.Jobs, Job{
		Id:        id,
		Executor:  executor,
		Initiator: initiator,
	})
	m.Lock.Unlock()
	return id
}

func (m *Manager) debugPrint() {
	output := "=== JOBS: "
	m.Lock.RLock()
	for _, job := range m.Jobs {
		output += fmt.Sprintf("[%d] ", job.Id)
	}
	m.Lock.RUnlock()
	output += " ==="
	log.Debug(output)
}

// Executes arbitrary Homescript-code as a given user, returns the output and a possible error slice
// The `scriptLabel` argument is used internally to allow for better error-display
// The `dryRun` argument specifies wheter the script should be linted or executed
// The `args` argument represents the arguments passed to the Homescript runtime and
// can be used from the script via the `CheckArg` and `GetArg` functions
// The `excludedCalls` argument specifies which Homescripts may not be called by this Homescript in order to prevent recursion
func (m *Manager) Run(
	username string,
	scriptLabel string,
	scriptCode string,
	dryRun bool,
	arguments map[string]string,
	callStack []string,
	initiator HomescriptInitiator,
) (string, int, []HomescriptError) {
	// Is passed to the executor so that it can forward messages from its own `SigTerm` onto the `sigTermInternalPtr`
	// Is also passed to `homescript.Run` so that the newly spawned interpreter uses the same channel
	interpreterSigTerm := make(chan int)

	executor := &Executor{
		Username:   username,
		ScriptName: scriptLabel,
		DryRun:     dryRun,
		Args:       arguments,
		CallStack:  callStack,
		// This channel will receive the initial sigTerm which can quit the currently running callback function
		// Additionally, the executor forwards the sigTerm to the interpreter which finally prevents any further node-evaluation
		// => Required for host functions to quit expensive / slow operations (sleep), then invokes an interpreter sigTerm
		SigTerm: make(chan int),
		// The sigterm pointer is also passed into the executor
		// => This pointer must ONLY be used internally, in this case is invoked from inside the `Executor`
		sigTermInternalPtr: &interpreterSigTerm,
		StartTime:          time.Now(),
	}

	// Append the executor to the jobs
	id := m.PushJob(
		executor,
		initiator,
		make(chan uint64),
	)

	m.debugPrint()

	// Run the script
	exitCode, hmsErrors := homescript.Run(
		executor,
		scriptLabel,
		scriptCode,
		&interpreterSigTerm,
	)

	// Remove the Job from the jobs list
	m.removeJob(id)

	m.debugPrint()

	// Process outcome
	if len(hmsErrors) > 0 {
		log.Debug(fmt.Sprintf("Homescript '%s' ran by user '%s' has terminated: %s", scriptLabel, username, hmsErrors[0].Message))
		return executor.Output, 1, convertErrors(hmsErrors...)
	}
	log.Debug(fmt.Sprintf("Homescript '%s' ran by user '%s' was executed successfully", scriptLabel, username))
	return executor.Output, exitCode, make([]HomescriptError, 0)
}

// Executes a given Homescript from the database and returns its output, exit-code and possible error
func (m *Manager) RunById(
	scriptId string,
	username string,
	callStack []string,
	dryRun bool,
	arguments map[string]string,
	initiator HomescriptInitiator,
) (string, int, error) {
	homescriptItem, hasBeenFound, err := database.GetUserHomescriptById(scriptId, username)
	if err != nil {
		return "database error", 500, err
	}
	if !hasBeenFound {
		return "not found error", 404, errors.New("Invalid Homescript id: no data associated with id")
	}
	output, exitCode, hmsErrors := m.Run(
		username,
		scriptId,
		homescriptItem.Data.Code,
		dryRun,
		arguments,
		// The script's id is added to the blacklist
		append(callStack, scriptId),
		initiator,
	)
	if len(hmsErrors) > 0 {
		return "execution error", exitCode, fmt.Errorf("Homescript terminated with exit code %d: %s", exitCode, hmsErrors[0].Message)
	}
	return output, exitCode, nil

}

// Removes an arbitrary job from the job list
// However, this function should only be used internally
// The function is automatically called when a Homescript execution ends
func (m *Manager) removeJob(jobId uint64) bool {
	jobsTemp := make([]Job, 0)
	success := false
	m.Lock.Lock()
	defer m.Lock.Unlock()
	for _, job := range m.Jobs {
		if job.Id == jobId {
			success = true
			continue
		}
		jobsTemp = append(jobsTemp, job)
	}
	m.Jobs = jobsTemp
	return success
}

// Returns a job given its ID
func (m *Manager) GetJobById(jobId uint64) (Job, bool) {
	m.Lock.RLock()
	defer m.Lock.RUnlock()
	for _, job := range m.Jobs {
		if job.Id == jobId {
			return job, true
		}
	}
	return Job{}, false
}

// Terminates a job given its internal job ID
// This method operates on all types of run-type
// The returned boolean indicates whether a job was killed or not
func (m *Manager) Kill(jobId uint64) bool {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	for _, job := range m.Jobs {
		if job.Id == jobId {
			// Exit code 10 means `killed via sigterm`
			job.Executor.SigTerm <- 10
			return true
		}
	}
	return false
}

// Terminates all jobs which are executing a given Homescript-ID / Homescript-label
// The returned boolean indicates whether a job was killed or not
func (m *Manager) KillAllId(hmsId string) (count uint64, success bool) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	for _, job := range m.Jobs {
		if job.Executor.ScriptName == hmsId {
			// Exit code 10 means `killed via sigterm`
			job.Executor.SigTerm <- 10
			success = true
			count++
		}
	}
	return count, success
}
