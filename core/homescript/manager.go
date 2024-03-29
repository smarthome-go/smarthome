package homescript

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/homescript/v3/homescript"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/compiler"
	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

const maxLinesErrMessage = 20
const KillEventFunction = "kill"
const KillEventMaxRuntime = 5 * time.Second
const jobIDNumDigits = 16

// Onlu for debugging.

const printDebugASM = true

var VM_LIMITS = runtime.CoreLimits{
	CallStackMaxSize: 128,
	StackMaxSize:     512,
	MaxMemorySize:    4096,
}

//
// Homescript manager
//

type Manager struct {
	Lock         sync.RWMutex
	Jobs         map[uint64]types.Job
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

// For external usage (can be marshaled)
type ApiJob struct {
	Jobid uint64  `json:"jobId"`
	HmsId *string `json:"hmsId"`
}

var HmsManager Manager

func InitManager() types.Manager {
	HmsManager = Manager{
		Lock:         sync.RWMutex{},
		Jobs:         make(map[uint64]types.Job),
		CompileCache: newManagerCompileCache(),
	}

	return &HmsManager
}

func (self *Manager) ClearCompileCache() {

}

func (m *Manager) generatePotentialJobId() uint64 {
	maxLimit := int64(int(math.Pow10(jobIDNumDigits)) - 1)
	lowLimit := uint64(math.Pow10(jobIDNumDigits - 1))

	randomNum, err := rand.Int(rand.Reader, big.NewInt(maxLimit))
	if err != nil {
		panic(err.Error())
	}

	randomInt := uint64(randomNum.Int64())

	if uint64(randomInt) <= lowLimit {
		randomInt += lowLimit
	}

	return randomInt
}

func (m *Manager) AllocJobId() uint64 {
	for {
		potential := m.generatePotentialJobId()
		m.Lock.Lock()
		_, invalid := m.Jobs[potential]

		if !invalid {
			m.Jobs[potential] = types.Job{}
			m.Lock.Unlock()
			return potential
		}
	}
}

func (m *Manager) PushJob(
	username string,
	initiator types.HomescriptInitiator,
	cancelCtxFunc context.CancelFunc,
	hmsId *string,
	vm *runtime.VM,
	entryModuleName string,
	supportsKill bool,
) uint64 {
	id := m.AllocJobId()
	m.setJob(id, username, initiator, cancelCtxFunc, hmsId, vm, entryModuleName, supportsKill)
	return id
}

func (m *Manager) setJob(
	id uint64,
	username string,
	initiator types.HomescriptInitiator,
	cancelCtxFunc context.CancelFunc,
	hmsId *string,
	vm *runtime.VM,
	entryModuleName string,
	supportsKill bool,
) {
	m.Lock.Lock()
	m.Jobs[id] = types.Job{
		Username:        username,
		JobID:           id,
		HmsID:           hmsId,
		Initiator:       initiator,
		CancelCtx:       cancelCtxFunc,
		VM:              vm,
		EntryModuleName: entryModuleName,
		SupportsKill:    supportsKill,
	}
	m.Lock.Unlock()
}

func (m *Manager) resolveFileContentsOfErrors(
	username string,
	mainModuleFilename string,
	mainModuleCode string,
	errors []types.HmsError,
) (map[string]string, error) {
	fileContents := make(map[string]string)

	for _, err := range errors {
		if err.Span.Filename == mainModuleFilename {
			continue
		}

		script, found, dbErr := m.GetPersonalScriptById(err.Span.Filename, username)
		if dbErr != nil {
			return nil, dbErr
		}
		if !found {
			spew.Dump(err.DiagnosticError)
			panic(fmt.Sprintf("Homescript with ID %s owned by user %s was not found", err.Span.Filename, username)) // TODO: no panic
		}

		fileContents[err.Span.Filename] = script.Data.Code
	}

	fileContents[mainModuleFilename] = mainModuleCode

	return fileContents, nil
}

func (m *Manager) Analyze(
	username string,
	filename string,
	code string,
	programKind types.HMS_PROGRAM_KIND,
	driverData *types.AnalyzerDriverMetadata,
) (map[string]ast.AnalyzedProgram, types.HmsRes, error) {
	analyzedModules, diagnostics, syntaxErrors := homescript.Analyze(
		homescript.InputProgram{
			Filename:    filename,
			ProgramText: code,
		},
		analyzerScopeAdditions(),
		newAnalyzerHost(username, programKind, driverData),
	)

	errors := make([]types.HmsError, 0)
	success := true

	if len(syntaxErrors) > 0 {
		success = false
		for _, syntax := range syntaxErrors {
			errors = append(errors, types.HmsError{
				SyntaxError: &types.HmsSyntaxError{
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
		errors = append(errors, types.HmsError{
			DiagnosticError: &types.HmsDiagnosticError{
				Level:   d.Level,
				Message: d.Message,
				Notes:   notesTemp,
			},
			Span: d.Span,
		})
	}

	fileContents, err := m.resolveFileContentsOfErrors(
		username,
		filename,
		code,
		errors,
	)
	if err != nil {
		return nil, types.HmsRes{}, err
	}

	return analyzedModules, types.HmsRes{
		Errors:       errors,
		FileContents: fileContents,
		Success:      success,
	}, nil
}

func (m *Manager) AnalyzeById(
	id string,
	username string,
	programKind types.HMS_PROGRAM_KIND,
	driverData *types.AnalyzerDriverMetadata,
) (map[string]ast.AnalyzedProgram, types.HmsRes, error) {
	hms, found, err := m.GetPersonalScriptById(id, username)
	if err != nil {
		return nil, types.HmsRes{}, err
	}
	if !found {
		panic(fmt.Sprintf("Homescript with ID %s owned by user %s was not found", id, username)) // TODO: no panic
	}

	return m.Analyze(username, id, hms.Data.Code, programKind, driverData)
}

func (m *Manager) Run(
	programKind types.HMS_PROGRAM_KIND,
	driverData *types.AnalyzerDriverMetadata,
	username string,
	filename *string,
	code string,
	initiator types.HomescriptInitiator,
	cancelCtx context.Context,
	cancelCtxFunc context.CancelFunc,
	idChan *chan uint64,
	args map[string]string,
	outputWriter io.Writer,
	automationContext *types.AutomationContext,
	// If this is left non-empty, an additional function is called after `init`.
	functionInvocation *runtime.FunctionInvocation,
	singletonsToLoad map[string]value.Value,
) (types.HmsRes, types.HmsRunResultContext, error) {
	// TODO: handle arguments

	// TODO: the @ symbol cannot be used in IDs?
	// FIX: implement this uniqueness properly
	programID := fmt.Sprintf("live@%d", time.Now().Nanosecond())
	if filename != nil {
		programID = *filename
	}

	logger.Trace(fmt.Sprintf("Homescript '%s' of user '%s' is being analyzed...", programID, username))

	modules, res, err := m.Analyze(username, programID, code, programKind, driverData)
	if err != nil {
		return types.HmsRes{}, types.HmsRunResultContext{}, err
	}

	if !res.Success {
		return res, types.HmsRunResultContext{}, nil
	}

	logger.Trace(fmt.Sprintf("Homescript '%s' of user '%s' is being compiled...", programID, username))

	comp := compiler.NewCompiler()
	prog := comp.Compile(modules, programID)

	if printDebugASM {
		fmt.Println(prog.AsmString())
	}

	logger.Debug(fmt.Sprintf("Homescript '%s' of user '%s' is executing...", programID, username))

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

	jobID := m.AllocJobId()

	executor := NewInterpreterExecutor(
		jobID,
		programID,
		username,
		outputWriter,
		args,
		automationContext,
		cancelCtxFunc,
		singletonsToLoad,
	)

	vm := runtime.NewVM(
		prog,
		executor,
		&cancelCtx,
		&cancelCtxFunc,
		interpreterScopeAdditions(),
		VM_LIMITS,
	)

	// supportsKill := modules[entryModuleName].SupportsEvent("kill")

	m.setJob(jobID, username, initiator, cancelCtxFunc, filename, &vm, programID, true)
	defer func() {
		// TODO: does this work?
		executor.Free()
		m.removeJob(jobID)
	}()

	// send the id to the id channel (only if it exists)
	if idChan != nil {
		*idChan <- jobID
	}

	fmt.Printf("Calling entry function `%s`\n", compiler.EntryPointFunctionIdent)

	// TODO: maybe add debugger support anytime

	// First, spawn the `@init` function.
	spawnResult := vm.SpawnSync(runtime.FunctionInvocation{
		Function: "@init",
		Args:     []value.Value{},
		FunctionSignature: runtime.FunctionInvocationSignature{
			Params:     []runtime.FunctionInvocationSignatureParam{},
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

		errors := make([]types.HmsError, 0)

		addErr := false
		isErr := true

		switch i.Kind() {
		case value.Vm_ExitInterruptKind: // ignore this
			exitI := i.(value.Vm_ExitInterrupt)
			if exitI.Code != 0 {
				errors = append(errors, types.HmsError{
					RuntimeInterrupt: &types.HmsRuntimeInterrupt{
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
			errors = append(errors, types.HmsError{
				RuntimeInterrupt: &types.HmsRuntimeInterrupt{
					Kind:    i.KindString(),
					Message: i.Message(),
				},
				Span: span,
			})
		}

		if isErr {
			fileContentsTemp, err := m.resolveFileContentsOfErrors(username, programID, code, errors)
			if err != nil {
				return types.HmsRes{}, types.HmsRunResultContext{}, err
			}

			fileContents = fileContentsTemp

			d := diagnostic.Diagnostic{
				Level:   diagnostic.DiagnosticLevelError,
				Message: errors[0].String(),
				Notes:   []string{},
				Span:    span,
			}

			errMsg := ""
			if logger.GetLevel() == logrus.TraceLevel {
				errMsg = d.Display(fileContentsTemp[errors[0].Span.Filename])
				split := strings.Split(errMsg, "\n")
				if len(split) > maxLinesErrMessage {
					errMsg = fmt.Sprintf("%s\n<%d more lines>", strings.Join(split[0:maxLinesErrMessage], "\n"), len(split)-maxLinesErrMessage)
				}
			} else {
				errMsg = errors[0].String()
			}

			logger.Trace()

			logger.Debug(fmt.Sprintf("Homescript '%s' of user '%s' failed: %s", programID, username, errMsg))
		}

		return types.HmsRes{
			Success:      !isErr,
			Errors:       errors,
			FileContents: fileContents,
		}, types.HmsRunResultContext{}, nil
	}

	logger.Debug(fmt.Sprintf("Homescript '%s' of user '%s' executed successfully", programID, username))

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

	return types.HmsRes{
			Success:      true,
			Errors:       make([]types.HmsError, 0),
			FileContents: make(map[string]string),
		},
		types.HmsRunResultContext{
			Singletons:         singletons,
			ReturnValue:        spawnResult.ReturnValue,
			CalledFunctionSpan: calledFunctionSpan,
		}, nil
}

// Executes a given Homescript from the database and returns its output, exit-code and possible error
func (m *Manager) RunById(
	programKind types.HMS_PROGRAM_KIND,
	driverData *types.AnalyzerDriverMetadata,
	hmsId string,
	username string,
	initiator types.HomescriptInitiator,
	cancelCtx context.Context,
	cancelCtxFunc context.CancelFunc,
	idChan *chan uint64,
	args map[string]string,
	outputWriter io.Writer,
	automationContext *types.AutomationContext,
	singletonsToLoad map[string]value.Value,
) (types.HmsRes, types.HmsRunResultContext, error) {
	script, found, err := m.GetPersonalScriptById(hmsId, username)
	if err != nil {
		return types.HmsRes{}, types.HmsRunResultContext{}, err
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
func (m *Manager) removeJob(jobID uint64) bool {
	m.Lock.Lock()
	_, found := m.Jobs[jobID]
	delete(m.Jobs, jobID)
	m.Lock.Unlock()

	success := found
	return success
}

// Returns a job given its ID
func (m *Manager) GetJobById(jobID uint64) (types.Job, bool) {
	m.Lock.RLock()
	job, found := m.Jobs[jobID]
	defer m.Lock.RUnlock()

	return job, found
}

// Terminates a job given its internal job ID
// This method operates on all types of run-type
// The returned boolean indicates whether a job was killed or not
func (m *Manager) Kill(jobID uint64) bool {
	job, found := m.GetJobById(jobID)
	if !found {
		return false
	}

	m.killJob(job)
	return true
}

// Terminates all jobs which are executing a given Homescript-ID / Homescript-label
// The returned boolean indicates whether a job was killed or not
func (m *Manager) KillAllId(hmsId string) (count uint64, success bool) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	for _, job := range m.Jobs {
		if job.HmsID == nil || *job.HmsID != hmsId {
			continue
		}

		// Only standalone scripts may be terminated (callstack validation) | TODO: implement this
		m.killJob(job)

		success = true
		count++
	}
	return count, success
}

func (m *Manager) killJob(job types.Job) {
	logger.Trace("Dispatching sigTerm to HMS interpreter channel...")

	_, killFnExists := job.VM.Program.Mappings.Functions[KillEventFunction]
	canceled := false
	cancelMtx := sync.Mutex{}

	if killFnExists {
		// Give timeout of 10 secs
		go func() {
			time.Sleep(KillEventMaxRuntime)

			defer cancelMtx.Unlock()
			cancelMtx.Lock()
			if !canceled {
				logger.Debugf("Job %d did not quit on time, terminating kill event...", job.JobID)
				job.CancelCtx()
			}
		}()

		job.VM.SpawnSync(runtime.FunctionInvocation{
			Function: KillEventFunction,
			Args:     []value.Value{},
			FunctionSignature: runtime.FunctionInvocationSignature{
				Params:     []runtime.FunctionInvocationSignatureParam{},
				ReturnType: ast.NewNullType(errors.Span{}),
			},
		}, nil)

		cancelMtx.Lock()
		canceled = true
		cancelMtx.Unlock()
	} else {
		job.CancelCtx()
	}

	logger.Trace("Successfully dispatched sigTerm to HMS interpreter channel")
}

// Can be used to access the manager's jobs from the outside in a safe manner
func (m *Manager) GetJobList() []types.Job {
	m.Lock.RLock()
	defer m.Lock.RUnlock()

	jobList := make([]types.Job, 0)
	for _, job := range m.Jobs {
		jobList = append(jobList, job)
	}

	return jobList
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
			Jobid: job.JobID,
			HmsId: job.HmsID,
		})
	}
	return jobs
}
