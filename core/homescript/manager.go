package homescript

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/smarthome-go/homescript/v3/homescript"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/compiler"
	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/interpreter/value"
	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/smarthome/core/database"
)

// this can be decremented if a script uses too many ressources
const CALL_STACK_LIMIT_SIZE = 2048

type HomescriptInitiator uint8

const (
	InitiatorAutomation         HomescriptInitiator = iota // triggered by a normal automation
	InitiatorAutomationOnNotify                            // triggered by an automation which runs on every notification
	InitiatorSchedule                                      // triggered by a schedule
	InitiatorExec                                          // triggered by a call to `exec`
	InitiatorInternal                                      // triggered internally
	InitiatorAPI                                           // triggered through the API
	InitiatorWidget                                        // triggered through a widget
)

//
// Homescript manager
//

type Manager struct {
	Lock         sync.RWMutex
	Jobs         []Job
	CompileCache ManagerCompileCache
}

type ManagerCompileCache struct {
	Cache map[string]compiler.Program
	Lock  sync.RWMutex
}

func newManagerCompileCache() ManagerCompileCache {
	return ManagerCompileCache{
		Cache: make(map[string]compiler.Program),
		Lock:  sync.RWMutex{},
	}
}

type Job struct {
	Username        string
	JobId           uint64
	HmsId           *string
	Initiator       HomescriptInitiator
	CancelCtx       context.CancelFunc
	Vm              *runtime.VM
	EntryModuleName string
	SupportsKill    bool
}

// For external usage (can be marshaled)
type ApiJob struct {
	Jobid uint64  `json:"jobId"`
	HmsId *string `json:"hmsId"`
}

var HmsManager Manager

func InitManager() {
	HmsManager = Manager{
		Lock:         sync.RWMutex{},
		Jobs:         make([]Job, 0),
		CompileCache: newManagerCompileCache(),
	}
}

func (self *Manager) ClearCompileCache() {

}

//
// Results and errors
//

type HmsRes struct {
	Success      bool
	Errors       []HmsError
	FileContents map[string]string
}

type HmsError struct {
	SyntaxError      *HmsSyntaxError      `json:"syntaxError"`
	DiagnosticError  *HmsDiagnosticError  `json:"diagnosticError"`
	RuntimeInterrupt *HmsRuntimeInterrupt `json:"runtimeError"`
	Span             errors.Span          `json:"span"`
}

func (self HmsError) String() string {
	spanDisplay := fmt.Sprintf("%s:%d:%d", self.Span.Filename, self.Span.Start.Line, self.Span.Start.Column)
	if self.SyntaxError != nil {
		return fmt.Sprintf("Syntax error at %s: %s", spanDisplay, self.SyntaxError.Message)
	} else if self.DiagnosticError != nil {
		return "Semantic error"
	} else if self.RuntimeInterrupt != nil {
		return fmt.Sprintf("%s at %s: %s", self.RuntimeInterrupt.Kind, spanDisplay, self.RuntimeInterrupt.Message)
	}
	panic("Illegal HmsError")
}

type HmsSyntaxError struct {
	Message string `json:"message"`
}

type HmsDiagnosticError struct {
	Level   diagnostic.DiagnosticLevel `json:"kind"`
	Message string                     `json:"message"`
	Notes   []string                   `json:"notes"`
}

type HmsRuntimeInterrupt struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

func (m *Manager) PushJob(
	username string,
	initiator HomescriptInitiator,
	cancelCtxFunc context.CancelFunc,
	hmsId *string,
	vm *runtime.VM,
	entryModuleName string,
	supportsKill bool,
) uint64 {
	m.Lock.Lock()
	id := uint64(len(m.Jobs))
	m.Jobs = append(m.Jobs, Job{
		Username:        username,
		JobId:           id,
		HmsId:           hmsId,
		Initiator:       initiator,
		CancelCtx:       cancelCtxFunc,
		Vm:              vm,
		EntryModuleName: entryModuleName,
		SupportsKill:    supportsKill,
	})
	m.Lock.Unlock()
	return id
}

func resolveFileContentsOfErrors(
	username string,
	mainModuleFilename string,
	mainModuleCode string,
	errors []HmsError,
) (map[string]string, error) {
	fileContents := make(map[string]string)

	for _, err := range errors {
		if err.Span.Filename == mainModuleFilename {
			continue
		}

		script, found, dbErr := database.GetUserHomescriptById(err.Span.Filename, username)
		if dbErr != nil {
			return nil, dbErr
		}
		if !found {
			spew.Dump(err.DiagnosticError)
			panic(fmt.Sprintf("Homescript with ID %s owned by user %s was not found", err.Span.Filename, username)) // TODO: no panic
		}

		fileContents[err.Span.Filename] = script.Data.Code
	}

	return fileContents, nil
}

func (m *Manager) Analyze(
	username string,
	filename string,
	code string,
) (map[string]ast.AnalyzedProgram, HmsRes, error) {
	analyzedModules, diagnostics, syntaxErrors := homescript.Analyze(
		homescript.InputProgram{
			Filename:    filename,
			ProgramText: code,
		},
		analyzerScopeAdditions(),
		newAnalyzerHost(username),
	)

	errors := make([]HmsError, 0)
	success := true

	if len(syntaxErrors) > 0 {
		success = false
		for _, syntax := range syntaxErrors {
			errors = append(errors, HmsError{
				SyntaxError: &HmsSyntaxError{
					Message: syntax.Message,
				},
				Span: syntax.Span,
			})
		}
	}

	for _, d := range diagnostics {
		if d.Level == diagnostic.DiagnosticLevelError {
			success = false
		}
		notesTemp := d.Notes
		if d.Notes == nil {
			notesTemp = make([]string, 0)
		}
		errors = append(errors, HmsError{
			DiagnosticError: &HmsDiagnosticError{
				Level:   d.Level,
				Message: d.Message,
				Notes:   notesTemp,
			},
			Span: d.Span,
		})
	}

	fileContents, err := resolveFileContentsOfErrors(
		username,
		filename,
		code,
		errors,
	)
	if err != nil {
		return nil, HmsRes{}, err
	}

	return analyzedModules, HmsRes{
		Errors:       errors,
		FileContents: fileContents,
		Success:      success,
	}, nil
}

func (m *Manager) AnalyzeById(
	id string,
	username string,
) (map[string]ast.AnalyzedProgram, HmsRes, error) {
	hms, found, err := database.GetUserHomescriptById(id, username)
	if err != nil {
		return nil, HmsRes{}, err
	}
	if !found {
		panic(fmt.Sprintf("Homescript with ID %s owned by user %s was not found", id, username)) // TODO: no panic
	}

	return m.Analyze(username, id, hms.Data.Code)
}

func (m *Manager) Run(
	username string,
	filename *string,
	code string,
	initiator HomescriptInitiator,
	cancelCtx context.Context,
	cancelCtxFunc context.CancelFunc,
	idChan *chan uint64,
	args map[string]string,
	outputWriter io.Writer,
	automationContext *AutomationContext,
	allowCache bool,
) (HmsRes, error) {
	// TODO: handle arguments

	// TODO: the @ symbol cannot be used in IDs?
	// FIX: implement this uniqueness properly
	entryModuleName := fmt.Sprintf("live@%d", time.Now().Nanosecond())
	if filename != nil {
		entryModuleName = *filename
	}

	m.CompileCache.Lock.RLock()
	cached, found := m.CompileCache.Cache[code]
	m.CompileCache.Lock.RUnlock()

	var program compiler.Program

	if found && allowCache {
		fmt.Println("Using cache")
		program = cached
	} else {
		modules, res, err := m.Analyze(username, entryModuleName, code)
		if err != nil {
			return HmsRes{}, err
		}

		if !res.Success {
			return res, nil
		}

		comp := compiler.NewCompiler()
		prog := comp.Compile(modules)

		// TODO: remove this debug output
		// i := 0
		// for name, function := range prog.Functions {
		// 	fmt.Printf("%03d ===> func: %s\n", i, name)
		//
		// 	for idx, inst := range function {
		// 		fmt.Printf("%03d | %s\n", idx, inst)
		// 	}
		//
		// 	i++
		// }
		//
		m.CompileCache.Lock.Lock()
		m.CompileCache.Cache[code] = prog
		m.CompileCache.Lock.Unlock()
		program = prog
	}

	log.Debug(fmt.Sprintf("Homescript '%s' of user '%s' is executing...", entryModuleName, username))

	// interpreter := interpreter.NewInterpreter(
	// 	CALL_STACK_LIMIT_SIZE,
	// 	newInterpreterExecutor(
	// 		username,
	// 		outputWriter,
	// 		args,
	// 		automationContext,
	// 		cancelCtxFunc,
	// 	),
	// 	modules,
	// 	interpreterScopeAdditions(),
	// 	&cancelCtx,
	// )

	vm := runtime.NewVM(
		program,
		newInterpreterExecutor(
			username,
			outputWriter,
			args,
			automationContext,
			cancelCtxFunc,
		),
		true,
		&cancelCtx,
		interpreterScopeAdditions(),
	)

	// supportsKill := modules[entryModuleName].SupportsEvent("kill")

	id := m.PushJob(username, initiator, cancelCtxFunc, filename, &vm, entryModuleName, true)
	defer m.removeJob(id)

	// send the id to the id channel (only if it exists)
	if idChan != nil {
		*idChan <- id
	}

	vm.Spawn(fmt.Sprintf("@%s_@init0", entryModuleName), false)

	if i := vm.Wait(); i != nil {
		span := errors.Span{}

		i := *i

		errors := make([]HmsError, 0)

		addErr := false
		isErr := true

		switch i.Kind() {
		case value.ExitInterruptKind: // ignore this
			exitI := i.(value.ExitInterrupt)
			if exitI.Code != 0 {
				errors = append(errors, HmsError{
					RuntimeInterrupt: &HmsRuntimeInterrupt{
						Kind:    "Exit",
						Message: fmt.Sprintf("Terminated using exit-code: %d", exitI.Code),
					},
					Span: exitI.Span,
				})
			} else {
				isErr = false
			}
			addErr = true
		case value.TerminateInterruptKind:
			termI := i.(value.TerminationInterrupt)
			span = termI.Span
		case value.RuntimeErrorInterruptKind:
			runtimeI := i.(value.RuntimeErr)
			span = runtimeI.Span
		default:
			panic(fmt.Sprintf("Another fatal interrupt was added without updating this code: %s", i.Kind()))
		}

		fileContents := make(map[string]string)

		if !addErr {
			errors = append(errors, HmsError{
				RuntimeInterrupt: &HmsRuntimeInterrupt{
					Kind:    i.Kind().String(),
					Message: i.Message(),
				},
				Span: span,
			})
		}

		if isErr {
			fileContentsTemp, err := resolveFileContentsOfErrors(username, entryModuleName, code, errors)
			if err != nil {
				return HmsRes{}, err
			}

			fileContents = fileContentsTemp

			log.Debug(fmt.Sprintf("Homescript '%s' of user '%s' failed: %s", entryModuleName, username, errors[0]))
		}

		return HmsRes{
			Success:      !isErr,
			Errors:       errors,
			FileContents: fileContents,
		}, nil
	}

	log.Debug(fmt.Sprintf("Homescript '%s' of user '%s' executed successfully", entryModuleName, username))

	return HmsRes{Success: true, Errors: make([]HmsError, 0), FileContents: make(map[string]string)}, nil
}

// Executes a given Homescript from the database and returns its output, exit-code and possible error
func (m *Manager) RunById(
	hmsId string,
	username string,
	initiator HomescriptInitiator,
	cancelCtx context.Context,
	cancelCtxFunc context.CancelFunc,
	idChan *chan uint64,
	args map[string]string,
	outputWriter io.Writer,
	automationContext *AutomationContext,
	allowCache bool,
) (HmsRes, error) {
	script, found, err := database.GetUserHomescriptById(hmsId, username)
	if err != nil {
		return HmsRes{}, err
	}
	if !found {
		panic(fmt.Sprintf("Homescript with ID %s owned by user %s was not found", hmsId, username)) // TODO: no panic
	}

	return m.Run(
		username,
		&hmsId,
		script.Data.Code,
		initiator,
		cancelCtx,
		cancelCtxFunc,
		idChan,
		args,
		outputWriter,
		automationContext,
		allowCache,
	)
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
		if job.JobId == jobId {
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
		if job.JobId == jobId {
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
		if job.JobId == jobId {
			m.killJob(job)
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
		if job.HmsId == nil || *job.HmsId != hmsId {
			continue
		}

		// Only standalone scripts may be terminated (callstack validation) | TODO: implement this
		m.killJob(job)

		success = true
		count++
	}
	return count, success
}

func (m *Manager) killJob(job Job) {
	killFn := fmt.Sprintf("@%s_@event_kill0", job.EntryModuleName)
	job.Vm.Spawn(killFn, true)

	canceled := false
	cancelMtx := sync.Mutex{}

	go func() {
		time.Sleep(10 * time.Second) // TODO: allow customization
		job.CancelCtx()
		cancelMtx.Lock()
		defer cancelMtx.Unlock()
		canceled = true
	}()

	// give timeout of 10 secs
	job.Vm.Wait()

	log.Trace("Dispatching sigTerm to HMS interpreter channel")
	cancelMtx.Lock()
	if !canceled {
		job.CancelCtx()
	}
	cancelMtx.Unlock()
	log.Trace("Successfully dispatched sigTerm to HMS interpreter channel")
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
		if job.Username != username {
			continue
		}

		// Skip any indirect jobs | TODO: do this
		// if len(job.Executor.CallStack) > 1 {
		// 	continue
		// }

		jobs = append(jobs, ApiJob{
			Jobid: job.JobId,
			HmsId: job.HmsId,
		})
	}
	return jobs
}
