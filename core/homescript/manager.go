package homescript

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/smarthome-go/homescript/v2/homescript"
	hmsErrors "github.com/smarthome-go/homescript/v2/homescript/errors"
	"github.com/smarthome-go/smarthome/core/database"
)

type HomescriptInitiator string

const (
	InitiatorAutomation         HomescriptInitiator = "automation"
	InitiatorAutomationOnNotify HomescriptInitiator = "automation_on_notify"
	InitiatorScheduler          HomescriptInitiator = "scheduler"
	InitiatorExec               HomescriptInitiator = "exec_target"
	InitiatorInternal           HomescriptInitiator = "internal"
	InitiatorAPI                HomescriptInitiator = "api"
	InitiatorWidget             HomescriptInitiator = "widget"
)

type HomescriptSigterm int

const (
	HmsSigtermSuccess         HomescriptSigterm = 0
	HmsSigtermCanceled        HomescriptSigterm = 10
	HmsSigtermRuntimeExceeded HomescriptSigterm = 20
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

type ApiJob struct {
	Id           uint64              `json:"id"`
	Initiator    HomescriptInitiator `json:"initiator"`
	HomescriptId string              `json:"homescriptId"`
}

type HmsExecRes struct {
	ReturnValue   homescript.Value
	RootScope     map[string]*homescript.Value
	ExitCode      int
	WasTerminated bool
	Errors        []HmsError
}

type HmsError struct {
	Kind         string         `json:"kind"`
	Message      string         `json:"message"`
	Span         hmsErrors.Span `json:"span"`
	FileContents string         `json:"code"`
}

func convertErrors(input []hmsErrors.Error) []HmsError {
	output := make([]HmsError, 0)
	for _, err := range input {
		output = append(output, HmsError{
			Kind:    err.Kind.String(),
			Message: err.Message,
			Span:    err.Span,
		})
	}
	return output
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
func (m *Manager) Analyze(
	scriptLabel string,
	scriptCode string,
	callStack []string,
	initiator HomescriptInitiator,
	username string,
	moduleStack []string,
	moduleName string,
	scopeInjections map[string]homescript.Value,
) []HmsError {
	executor := &AnalyzerExecutor{
		Username: username,
	}

	// Append the executor to the jobs
	id := m.PushJob(
		&Executor{SigTerm: make(chan int)},
		initiator,
		make(chan uint64),
	)

	scopeAdditionsFinal := make(map[string]homescript.Value)

	for key, value := range scopeInjections {
		_, exists := scopeAdditionsFinal[key]
		if exists {
			panic(fmt.Sprintf("Duplicate scope insertion key `%s`", key))
		}
		// insert this value
		scopeAdditionsFinal[key] = value
	}

	if _, exists := scopeAdditionsFinal["context"]; !exists {
		// Include `context` in order to prevent false errors during analysis
		scopeAdditionsFinal["context"] = homescript.ValueBuiltinVariable{
			Callback: func(executor homescript.Executor, span hmsErrors.Span) (homescript.Value, *hmsErrors.Error) {
				return homescript.ValueObject{IsDynamic: true, IsProtected: true, ObjFields: make(map[string]*homescript.Value)}, nil
			},
		}
	}

	// Run the script
	diagnostics, _, _ := homescript.Analyze(
		executor,
		scriptCode,
		scopeAdditionsFinal,
		moduleStack,
		moduleName,
		moduleName,
	)

	// Remove the Job from the jobs list when this function ends
	m.removeJob(id)

	diagnosticsErr := make([]HmsError, 0)
	for _, diagnostic := range diagnostics {
		diagnosticsErr = append(diagnosticsErr, HmsError{
			Kind:    diagnostic.Kind.String(),
			Message: diagnostic.Message,
			Span:    diagnostic.Span,
		})
	}

	return diagnosticsErr
}

func (m *Manager) AnalyzeById(
	scriptId string,
	username string,
	callStack []string,
	initiator HomescriptInitiator,
	scopeInjections map[string]homescript.Value,
) ([]HmsError, error) {
	homescriptItem, hasBeenFound, err := database.GetUserHomescriptById(scriptId, username)
	if err != nil {
		return nil, err
	}
	if !hasBeenFound {
		return nil, errors.New("invalid Homescript id: no data associated with id")
	}
	return m.Analyze(
		scriptId,
		homescriptItem.Data.Code,
		append(callStack, scriptId),
		initiator,
		username,
		make([]string, 0),
		scriptId,
		scopeInjections,
	), nil
}

// Executes arbitrary Homescript-code as a given user, returns the output and a possible error slice
// The `scriptLabel` argument is used internally to allow for better error-display
// The `excludedCalls` argument specifies which Homescripts may not be called by this Homescript in order to prevent recursion
func (m *Manager) Run(
	username string,
	scriptLabel string,
	scriptCode string,
	arguments map[string]string,
	callStack []string,
	initiator HomescriptInitiator,
	sigTerm chan int,
	outputWriter io.Writer,
	idChan *chan uint64,
	scopeInjections map[string]homescript.Value,
) HmsExecRes {
	// Is passed to the executor so that it can forward messages from its own `SigTerm` onto the `sigTermInternalPtr`
	// Is also passed to `homescript.Run` so that the newly spawned interpreter uses the same channel
	interpreterSigTerm := make(chan int)

	executor := &Executor{
		Username:   username,
		ScriptName: scriptLabel,
		DryRun:     false,
		CallStack:  callStack,
		// This channel will receive the initial sigTerm which can quit the currently running callback function
		// Additionally, the executor forwards the sigTerm to the interpreter which finally prevents any further node-evaluation
		// => Required for host functions to quit expensive / slow operations (sleep), then invokes an interpreter sigTerm
		SigTerm: sigTerm,
		// The sigterm pointer is also passed into the executor
		// => This pointer must ONLY be used internally, in this case is invoked from inside the `Executor`
		sigTermInternalPtr: &interpreterSigTerm,
		StartTime:          time.Now(),
		OutputWriter:       outputWriter,
		Initiator:          initiator,
	}

	// Append the executor to the jobs
	id := m.PushJob(
		executor,
		initiator,
		make(chan uint64),
	)

	// Only send back the id if the channel exists
	if idChan != nil {
		*idChan <- id
	}

	valueArgs := make(map[string]homescript.Value)
	for key, value := range arguments {
		valueArgs[key] = homescript.ValueString{Value: value}
	}

	scopeAdditionsFinal := make(map[string]homescript.Value)
	for key, value := range scopeInjections {
		_, exists := scopeAdditionsFinal[key]
		if exists {
			panic(fmt.Sprintf("Duplicate scope insertion key `%s`", key))
		}
		// insert this value
		scopeAdditionsFinal[key] = value
	}

	if _, exists := scopeAdditionsFinal["context"]; !exists {
		// Include a default null `context` if ther is no other
		scopeAdditionsFinal["context"] = homescript.ValueBuiltinVariable{
			Callback: func(executor homescript.Executor, span hmsErrors.Span) (homescript.Value, *hmsErrors.Error) {
				return homescript.ValueNull{}, nil
			},
		}
	}

	// Run the script
	returnValue, exitCode, rootScope, hmsErrors := homescript.Run(
		executor,
		&interpreterSigTerm,
		scriptCode,
		scopeAdditionsFinal,
		valueArgs,
		false,
		10000,
		make([]string, 0),
		scriptLabel,
		scriptLabel,
	)

	wasTerminated := executor.WasTerminated

	// Remove the Job from the jobs list when this function ends
	m.removeJob(id)

	if len(hmsErrors) > 0 {
		log.Debug(fmt.Sprintf("Homescript '%s' ran by user '%s' has terminated: %s", scriptLabel, username, hmsErrors[0].Message))
	} else if wasTerminated {
		log.Debug(fmt.Sprintf("Homescript '%s' ran by user '%s' was terminated", scriptLabel, username))
	} else {
		log.Debug(fmt.Sprintf("Homescript '%s' ran by user '%s' was executed successfully", scriptLabel, username))
	}

	if returnValue == nil {
		returnValue = homescript.ValueNull{}
	}

	// Process outcome
	return HmsExecRes{
		ReturnValue:   returnValue,
		RootScope:     rootScope,
		ExitCode:      exitCode,
		WasTerminated: wasTerminated,
		Errors:        convertErrors(hmsErrors),
	}
}

// Executes a given Homescript from the database and returns its output, exit-code and possible error
func (m *Manager) RunById(
	scriptId string,
	username string,
	callStack []string,
	arguments map[string]string,
	initiator HomescriptInitiator,
	sigTerm chan int,
	outputWriter io.Writer,
	idChan *chan uint64,
	scopeInjections map[string]homescript.Value,
) (HmsExecRes, error) {
	homescriptItem, hasBeenFound, err := database.GetUserHomescriptById(scriptId, username)
	if err != nil {
		return HmsExecRes{}, err
	}
	if !hasBeenFound {
		return HmsExecRes{}, errors.New("invalid Homescript id: no data associated with id")
	}
	return m.Run(
		username,
		scriptId,
		homescriptItem.Data.Code,
		arguments,
		// The script's id is added to the callStack (exec blacklist)
		append(callStack, scriptId),
		initiator,
		sigTerm,
		outputWriter,
		idChan,
		scopeInjections,
	), nil
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
func (m *Manager) Kill(jobId uint64, sigtermType HomescriptSigterm) bool {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	for _, job := range m.Jobs {
		if job.Id == jobId {
			job.Executor.InExpensiveBuiltin.Mutex.Lock()
			if job.Executor.InExpensiveBuiltin.Value {
				job.Executor.InExpensiveBuiltin.Mutex.Unlock()
				// If the executor is currently handling an expensive builtin function, terminate it
				log.Trace("Dispatching sigTerm to executor channel")
				job.Executor.SigTerm <- int(sigtermType)
				log.Trace("Successfully dispatched sigTerm to executor channel")
			} else {
				job.Executor.InExpensiveBuiltin.Mutex.Unlock()
				// Otherwise, terminate the interpreter directly
				log.Trace("Dispatching sigTerm to HMS interpreter channel")
				*job.Executor.sigTermInternalPtr <- int(sigtermType)
				log.Trace("Successfully dispatched sigTerm to HMS interpreter channel")
			}
			return true
		}
	}
	return false
}

// Terminates all jobs which are executing a given Homescript-ID / Homescript-label
// The returned boolean indicates whether a job was killed or not
func (m *Manager) KillAllId(hmsId string, sigtermType HomescriptSigterm) (count uint64, success bool) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	for _, job := range m.Jobs {
		// Only standalone scripts may be terminated (callstack validation)
		if job.Executor.ScriptName == hmsId && len(job.Executor.CallStack) < 2 {
			job.Executor.InExpensiveBuiltin.Mutex.Lock()
			if job.Executor.InExpensiveBuiltin.Value {
				// If the executor is currently handling an expensive builtin function, terminate it
				log.Trace("Dispatching sigTerm to executor channel")
				job.Executor.SigTerm <- int(sigtermType)
				log.Trace("Successfully dispatched sigTerm to executor channel")
			} else {
				// Otherwise, terminate the interpreter directly
				log.Trace("Dispatching sigTerm to HMS interpreter channel")
				*job.Executor.sigTermInternalPtr <- int(sigtermType)
				log.Trace("Successfully dispatched sigTerm to HMS interpreter channel")
			}
			job.Executor.InExpensiveBuiltin.Mutex.Unlock()
			success = true
			count++
		}
	}
	return count, success
}

// Can be used to access the manager's jobs from the outside in a safe manner
func (m *Manager) GetJobList() []Job {
	m.Lock.RLock()
	defer m.Lock.RUnlock()
	return m.Jobs
}

// Returns just the jobs which are executed by the specified user
// Filter out any indirect runtimes which are managed by this manager
func (m *Manager) GetUserDirectJobs(username string) []ApiJob {
	allJobs := m.GetJobList()
	jobs := make([]ApiJob, 0)

	for _, job := range allJobs {
		// Skip any jobs which are not executed by the specified user
		if job.Executor.Username != username {
			continue
		}
		// Skip any indirect jobs
		if len(job.Executor.CallStack) > 1 {
			continue
		}
		jobs = append(jobs, ApiJob{
			Id:           job.Id,
			Initiator:    job.Initiator,
			HomescriptId: job.Executor.ScriptName,
		})
	}
	return jobs
}
