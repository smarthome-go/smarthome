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
	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
)

const KillEventFunction = "kill"
const KillEventMaxRuntime = 5 * time.Second

// this can be decremented if a script uses too many resources
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

func HmsErrorsFromDiagnostics(input []diagnostic.Diagnostic) []HmsError {
	output := make([]HmsError, 0)

	for _, diagnosticMsg := range input {
		if diagnosticMsg.Level != diagnostic.DiagnosticLevelError {
			continue
		}

		output = append(output, HmsError{
			SyntaxError: nil,
			DiagnosticError: &HmsDiagnosticError{
				Level:   diagnostic.DiagnosticLevelError,
				Message: diagnosticMsg.Message,
				Notes:   diagnosticMsg.Notes,
			},
			RuntimeInterrupt: nil,
			Span:             diagnosticMsg.Span,
		})
	}

	return output
}

func (self HmsError) String() string {
	spanDisplay := fmt.Sprintf("%s:%d:%d", self.Span.Filename, self.Span.Start.Line, self.Span.Start.Column)
	if self.SyntaxError != nil {
		return fmt.Sprintf("Syntax error at %s: `%s`", spanDisplay, self.SyntaxError.Message)
	} else if self.DiagnosticError != nil {
		return fmt.Sprintf("Semantic error at %s: `%s`", spanDisplay, self.DiagnosticError.Message)
	} else if self.RuntimeInterrupt != nil {
		return fmt.Sprintf("%s at %s: `%s`", self.RuntimeInterrupt.Kind, spanDisplay, self.RuntimeInterrupt.Message)
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

		script, found, dbErr := GetPersonalScriptById(err.Span.Filename, username)
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
	programKind HMS_PROGRAM_KIND,
	driverData *AnalyzerDriverMetadata,
) (map[string]ast.AnalyzedProgram, HmsRes, error) {
	analyzedModules, diagnostics, syntaxErrors := homescript.Analyze(
		homescript.InputProgram{
			Filename:    filename,
			ProgramText: code,
		},
		analyzerScopeAdditions(),
		newAnalyzerHost(username, programKind, driverData),
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
	programKind HMS_PROGRAM_KIND,
	driverData *AnalyzerDriverMetadata,
) (map[string]ast.AnalyzedProgram, HmsRes, error) {
	hms, found, err := GetPersonalScriptById(id, username)
	if err != nil {
		return nil, HmsRes{}, err
	}
	if !found {
		panic(fmt.Sprintf("Homescript with ID %s owned by user %s was not found", id, username)) // TODO: no panic
	}

	return m.Analyze(username, id, hms.Data.Code, programKind, driverData)
}

// NOTE: this is primarily required for the driver.
type HmsRunResultContext struct {
	Singletons map[string]value.Value
	// This is `nil` if no additional function was invoced or the called function did not return a value.
	ReturnValue value.Value
	// This is non zero-valued if an additional function is called.
	CalledFunctionSpan errors.Span
}

func (m *Manager) Run(
	programKind HMS_PROGRAM_KIND,
	driverData *AnalyzerDriverMetadata,
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
	// If this is left non-empty, an additional function is called after `init`.
	functionInvocation *runtime.FunctionInvocation,
	singletonsToLoad map[string]value.Value,
) (HmsRes, HmsRunResultContext, error) {
	// TODO: handle arguments

	// TODO: the @ symbol cannot be used in IDs?
	// FIX: implement this uniqueness properly
	entryModuleName := fmt.Sprintf("live@%d", time.Now().Nanosecond())
	if filename != nil {
		entryModuleName = *filename
	}

	log.Trace(fmt.Sprintf("Homescript '%s' of user '%s' is being analyzed...", entryModuleName, username))

	modules, res, err := m.Analyze(username, entryModuleName, code, programKind, driverData)
	if err != nil {
		return HmsRes{}, HmsRunResultContext{}, err
	}

	if !res.Success {
		return res, HmsRunResultContext{}, nil
	}

	log.Trace(fmt.Sprintf("Homescript '%s' of user '%s' is being compiled...", entryModuleName, username))

	comp := compiler.NewCompiler()
	prog := comp.Compile(modules, entryModuleName)

	// TODO: remove this debug output
	i := 0
	for name, function := range prog.Functions {
		fmt.Printf("%03d ===> func: %s\n", i, name)

		for idx, inst := range function {
			fmt.Printf("%03d | %s\n", idx, inst)
		}

		i++
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
		prog,
		NewInterpreterExecutor(
			username,
			outputWriter,
			args,
			automationContext,
			cancelCtxFunc,
			singletonsToLoad,
		),
		&cancelCtx,
		&cancelCtxFunc,
		interpreterScopeAdditions(),
		runtime.CoreLimits{
			CallStackMaxSize: 10000, // TODO: Make limits dynamic?
			StackMaxSize:     10000,
			MaxMemorySize:    10000,
		},
	)

	// supportsKill := modules[entryModuleName].SupportsEvent("kill")

	id := m.PushJob(username, initiator, cancelCtxFunc, filename, &vm, entryModuleName, true)
	defer m.removeJob(id)

	// send the id to the id channel (only if it exists)
	if idChan != nil {
		*idChan <- id
	}

	fmt.Printf("Calling entry function `%s`\n", compiler.EntryPointFunctionIdent)

	// TODO: maybe add debugger support anytime

	// First, spawn the `@init` function.
	spawnResult := vm.SpawnSync(runtime.FunctionInvocation{
		Function: "@init",
		Args:     []value.Value{},
		FunctionSignature: runtime.FunctionInvocationSignature{
			Params:     map[string]ast.Type{},
			ReturnType: ast.NewNullType(errors.Span{}),
		},
	}, nil)

	// If the `@init` function completed successfully, run the optional function routing.
	if spawnResult.Exception == nil && functionInvocation != nil {
		spawnResult = vm.SpawnSync(*functionInvocation, nil)
	}

	if spawnResult.Exception != nil {
		i := spawnResult.Exception.Interrupt

		span := errors.Span{}

		errors := make([]HmsError, 0)

		addErr := false
		isErr := true

		switch i.Kind() {
		case value.Vm_ExitInterruptKind: // ignore this
			exitI := i.(value.Vm_ExitInterrupt)
			if exitI.Code != 0 {
				errors = append(errors, HmsError{
					RuntimeInterrupt: &HmsRuntimeInterrupt{
						Kind: "Exit",
						Message: fmt.Sprintf(
							"Core %d terminated with exit-code: %d",
							spawnResult.Exception.CoreNum,
							exitI.Code,
						),
					},
					Span: exitI.Span,
				})
			} else {
				isErr = false
			}
			addErr = true
		case value.Vm_TerminateInterruptKind:
			termI := i.(value.VmTerminationInterrupt)
			span = termI.Span
		case value.Vm_FatalExceptionInterruptKind:
			runtimeI := i.(value.VmFatalException)
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
				return HmsRes{}, HmsRunResultContext{}, err
			}

			fileContents = fileContentsTemp

			log.Debug(fmt.Sprintf("Homescript '%s' of user '%s' failed: %s", entryModuleName, username, errors[0]))
		}

		return HmsRes{
			Success:      !isErr,
			Errors:       errors,
			FileContents: fileContents,
		}, HmsRunResultContext{}, nil
	}

	log.Debug(fmt.Sprintf("Homescript '%s' of user '%s' executed successfully", entryModuleName, username))

	// Stores the original (non-mangled) singletons of the entry module.
	singletons := make(map[string]value.Value)
	for name, mangled := range prog.Mappings.Singletons {
		singletons[name] = vm.GetGlobals()[mangled]
	}

	calledFunctionSpan := errors.Span{}
	if functionInvocation != nil {
		calledFunctionSpan = vm.SourceMap(runtime.CallFrame{
			Function:           vm.Program.Mappings.Functions[functionInvocation.Function],
			InstructionPointer: 0,
		})
	}

	return HmsRes{
			Success:      true,
			Errors:       make([]HmsError, 0),
			FileContents: make(map[string]string),
		},
		HmsRunResultContext{
			Singletons:         singletons,
			ReturnValue:        spawnResult.ReturnValue,
			CalledFunctionSpan: calledFunctionSpan,
		}, nil
}

// Executes a given Homescript from the database and returns its output, exit-code and possible error
func (m *Manager) RunById(
	programKind HMS_PROGRAM_KIND,
	driverData *AnalyzerDriverMetadata,
	hmsId string,
	username string,
	initiator HomescriptInitiator,
	cancelCtx context.Context,
	cancelCtxFunc context.CancelFunc,
	idChan *chan uint64,
	args map[string]string,
	outputWriter io.Writer,
	automationContext *AutomationContext,
	singletonsToLoad map[string]value.Value,
) (HmsRes, HmsRunResultContext, error) {
	script, found, err := GetPersonalScriptById(hmsId, username)
	if err != nil {
		return HmsRes{}, HmsRunResultContext{}, err
	}
	if !found {
		panic(fmt.Sprintf("Homescript with ID %s owned by user %s was not found", hmsId, username)) // TODO: no panic
	}

	return m.Run(
		programKind,
		driverData,
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
		// Do not use any user-defined entry function.
		nil,
		singletonsToLoad,
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
	idx := 0

	for {
		m.Lock.Lock()
		jobLen := len(m.Jobs)
		if idx >= jobLen {
			m.Lock.Unlock()
			return false
		}

		job := m.Jobs[idx]
		m.Lock.Unlock()

		if job.JobId == jobId {
			m.killJob(job)
			return true
		}

		idx++
	}
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
	log.Trace("Dispatching sigTerm to HMS interpreter channel...")

	_, killFnExists := job.Vm.Program.Mappings.Functions[KillEventFunction]
	canceled := false
	cancelMtx := sync.Mutex{}
	if killFnExists {
		// Give timeout of 10 secs
		go func() {
			time.Sleep(KillEventMaxRuntime)

			defer cancelMtx.Unlock()
			cancelMtx.Lock()
			if !canceled {
				log.Debugf("Job %d did not quit on time, terminating kill event...", job.JobId)
				job.CancelCtx()
			}
		}()

		job.Vm.SpawnSync(runtime.FunctionInvocation{
			Function: KillEventFunction,
			Args:     []value.Value{},
			FunctionSignature: runtime.FunctionInvocationSignature{
				Params:     map[string]ast.Type{},
				ReturnType: ast.NewNullType(errors.Span{}),
			},
		}, nil)

		cancelMtx.Lock()
		canceled = true
		cancelMtx.Unlock()
	}

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
