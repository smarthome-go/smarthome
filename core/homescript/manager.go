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
	"github.com/smarthome-go/smarthome/core/database"
	driverTypes "github.com/smarthome-go/smarthome/core/device/driver/types"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

const maxLinesErrMessage = 20
const KillEventFunction = "kill"
const KillEventMaxRuntime = 5 * time.Second
const jobIDNumDigits = 16

// Only for debugging.

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

// For external usage (can be marshaled)
type ApiJob struct {
	Jobid uint64 `json:"jobId"`
	HmsId string `json:"hmsId"`
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

func (m *Manager) ClearCompileCache() {
	panic("TODO")
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
	context types.ExecutionContext,
	cancelCtxFunc context.CancelFunc,
	hmsID string,
	vm *runtime.VM,
	supportsKill bool,
) uint64 {
	id := m.AllocJobId()
	m.setJob(
		id,
		hmsID,
		cancelCtxFunc,
		vm,
		context,
	)
	return id
}

func (m *Manager) setJob(
	jobID uint64,
	hmsId string,
	cancelCtxFunc context.CancelFunc,
	vm *runtime.VM,
	context types.ExecutionContext,
) {
	m.Lock.Lock()
	m.Jobs[jobID] = types.Job{
		Context:   context,
		JobID:     jobID,
		HmsID:     hmsId,
		CancelCtx: cancelCtxFunc,
		VM:        vm,
	}
	m.Lock.Unlock()
}

func (m *Manager) resolveFileContentsOfErrors(
	source homescript.InputProgram,
	errors []types.HmsError,
	context types.ExecutionContext,
) (map[string]string, error) {
	fileContents := make(map[string]string)

	for _, err := range errors {
		if err.Span.Filename == source.Filename {
			continue
		}

		var code string

		if context.Username() != nil {
			script, found, dbErr := m.GetPersonalScriptById(err.Span.Filename, *context.Username())
			if dbErr != nil {
				return nil, dbErr
			}
			if !found {
				spew.Dump(err.DiagnosticError)
				panic(fmt.Sprintf("Homescript with ID %s owned by user %s was not found", err.Span.Filename, *context.Username()))
			}
			code = script.Data.Code
		} else {
			script, found, dbErr := m.GetScriptById(err.Span.Filename) // TODO: this will probably not work
			if dbErr != nil {
				return nil, dbErr
			}
			if !found {
				spew.Dump(err.DiagnosticError)
				panic(fmt.Sprintf("Homescript with ID %s was not found", err.Span.Filename))
			}
			code = script.Data.Code
		}

		fileContents[err.Span.Filename] = code
	}

	fileContents[source.Filename] = source.ProgramText

	return fileContents, nil
}

func (m *Manager) Analyze(
	input homescript.InputProgram,
	context types.ExecutionContext,
) (map[string]ast.AnalyzedProgram, types.HmsDiagnosticsContainer, error) {
	logger.Trace(fmt.Sprintf("Homescript `%s` is being analyzed...", input.Filename))

	analyzedModules, diagnostics, syntaxErrors := homescript.Analyze(
		input,
		analyzerScopeAdditions(),
		newAnalyzerHost(context),
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

	if !success {

	}

	fileContents, err := m.resolveFileContentsOfErrors(
		homescript.InputProgram{
			ProgramText: input.ProgramText,
			Filename:    input.Filename,
		},
		errors,
		context,
	)
	if err != nil {
		return nil, types.HmsDiagnosticsContainer{}, err
	}

	return analyzedModules, types.HmsDiagnosticsContainer{
		ContainsError: !success,
		Diagnostics:   errors,
		FileContents:  fileContents,
	}, nil
}

func (m *Manager) AnalyzeUserScript(
	context types.ExecutionContextUser,
) (map[string]ast.AnalyzedProgram, types.HmsDiagnosticsContainer, error) {
	hms, found, err := m.GetPersonalScriptById(context.Filename, context.UsernameData)
	if err != nil {
		return nil, types.HmsDiagnosticsContainer{}, err
	}
	if !found {
		return nil,
			types.HmsDiagnosticsContainer{},
			fmt.Errorf("Homescript with ID `%s` owned by user %s was not found", context.Filename, context.UsernameData)
	}

	return m.Analyze(
		homescript.InputProgram{
			ProgramText: hms.Data.Code,
			Filename:    hms.Data.Id,
		},
		context,
	)
}

func (m *Manager) AnalyzeDriver(
	context types.ExecutionContextDriver,
) (map[string]ast.AnalyzedProgram, types.HmsDiagnosticsContainer, error) {
	driver, found, err := database.GetDeviceDriver(
		context.DriverVendor,
		context.DriverModel,
	)

	if err != nil {
		return nil, types.HmsDiagnosticsContainer{}, err
	}

	if !found {
		return nil,
			types.HmsDiagnosticsContainer{},
			fmt.Errorf("Driver `%s:%s` was not found", driver.VendorId, driver.ModelId)
	}

	return m.Analyze(
		homescript.InputProgram{
			ProgramText: driver.HomescriptCode,
			Filename: types.CreateDriverHmsId(database.DriverTuple{
				VendorID: driver.VendorId,
				ModelID:  driver.ModelId,
			}),
		},
		context,
	)
}

func (m *Manager) RunGeneric(
	invocation types.ProgramInvocation,
	context types.ExecutionContext,
	cancelation types.Cancelation,
	// This is required for the asyncronous runtime.
	idChan *chan uint64,
	outputWriter io.Writer,
) (types.HmsRes, error) {
	modules, analyzerRes, err := m.Analyze(
		invocation.Identifier,
		context,
	)
	if err != nil {
		return types.HmsRes{}, err
	}

	if analyzerRes.ContainsError {
		return types.HmsRes{
			Errors:             analyzerRes,
			Singletons:         nil,
			ReturnValue:        nil,
			CalledFunctionSpan: errors.Span{},
		}, nil
	}

	logger.Tracef("Homescript `%s` is being compiled...", invocation.Identifier.Filename)

	jobID := m.AllocJobId()

	compOut, err := m.Compile(modules, invocation.Identifier.Filename)
	if err != nil {
		return types.HmsRes{}, err
	}

	if printDebugASM {
		fmt.Println(compOut.AsmString())
	}

	logger.Debugf("Homescript `%s` is executing...", invocation.Identifier.Filename)

	executor := NewInterpreterExecutor(
		jobID,
		invocation.Identifier.Filename,
		outputWriter,
		context,
	)

	vm := runtime.NewVM(
		compOut,
		executor,
		&cancelation.Context,
		&cancelation.CancelFunc,
		interpreterScopeAdditions(),
		VM_LIMITS,
	)

	m.setJob(
		jobID,
		invocation.Identifier.Filename,
		cancelation.CancelFunc,
		&vm,
		context,
	)

	defer func() {
		executor.Free()
		m.removeJob(jobID)
	}()

	// Send the id to the id channel (only if it exists).
	if idChan != nil {
		*idChan <- jobID
	}

	// If there is no explicit invocation, call the `main` function.
	functionInvocation := runtime.FunctionInvocation{
		Function: compiler.MainFunctionIdent,
		Args:     []value.Value{},
		FunctionSignature: runtime.FunctionInvocationSignature{
			Params:     []runtime.FunctionInvocationSignatureParam{},
			ReturnType: ast.NewNullType(errors.Span{}),
		},
	}

	if invocation.FunctionInvocation != nil {
		functionInvocation = *invocation.FunctionInvocation
	}

	spawnResult := vm.SpawnSync(*invocation.FunctionInvocation, nil)

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
			fileContentsTemp, err := m.resolveFileContentsOfErrors(
				invocation.Identifier,
				errors,
				context,
			)
			if err != nil {
				return types.HmsRes{}, err
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

			logger.Debug(fmt.Sprintf("Homescript `%s` failed: %s", invocation.Identifier.Filename, errMsg))
		}

		return types.HmsRes{
			Errors: types.HmsDiagnosticsContainer{
				ContainsError: true,
				Diagnostics:   errors,
				FileContents:  fileContents,
			},
			Singletons:  nil,
			ReturnValue: nil,
			// CalledFunctionSpan: span,
		}, nil
	}

	logger.Debug(fmt.Sprintf("Homescript `%s` executed successfully", invocation.Identifier.Filename))

	// Stores the original (non-mangled) singletons of the entry module.
	singletons := make(map[string]value.Value)
	for name, mangled := range compOut.Mappings.Singletons {
		singletons[name] = vm.GetGlobals()[mangled]
	}

	calledFunctionSpan := errors.Span{}
	if invocation.FunctionInvocation != nil {
		calledFunctionSpan = vm.SourceMap(runtime.CallFrame{
			Function:           vm.Program.Mappings.Functions[functionInvocation.Function],
			InstructionPointer: 0,
		})
	}

	fileContentsTemp, err := m.resolveFileContentsOfErrors(
		invocation.Identifier,
		analyzerRes.Diagnostics,
		context,
	)
	if err != nil {
		return types.HmsRes{}, err
	}

	return types.HmsRes{
		Errors: types.HmsDiagnosticsContainer{
			ContainsError: false,
			Diagnostics:   analyzerRes.Diagnostics,
			FileContents:  fileContentsTemp,
		},
		Singletons:         singletons,
		ReturnValue:        spawnResult.ReturnValue,
		CalledFunctionSpan: calledFunctionSpan,
	}, nil
}

// TODO: maybe add argument support
func (m *Manager) RunUserScript(
	programID, username string,
	function *runtime.FunctionInvocation,
	cancelation types.Cancelation,
	outputWriter io.Writer,
	idChan *chan uint64,
) (types.HmsRes, error) {
	script, found, err := m.GetPersonalScriptById(programID, username)
	if err != nil {
		return types.HmsRes{}, err
	}
	if !found {
		return types.HmsRes{}, fmt.Errorf("Homescript with ID `%s` owned by user `%s` was not found", programID, username)
	}

	return m.RunGeneric(
		types.ProgramInvocation{
			Identifier: homescript.InputProgram{
				ProgramText: script.Data.Id,
				Filename:    script.Data.Code,
			},
			FunctionInvocation: function,
			SingletonsToLoad:   map[string]value.Value{},
		},
		types.NewExecutionContextUser(
			username,
			nil,
		),
		cancelation,
		idChan,
		outputWriter,
	)
}

func (m *Manager) RunDriverScript(
	driverIDs driverTypes.DriverInvocationIDs,
	invocation runtime.FunctionInvocation,
	cancelation types.Cancelation,
	outputWriter io.Writer,
) (types.HmsRes, error) {
	return types.HmsRes{}, nil
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
		if job.HmsID != hmsId {
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
		if job.Context.Username() == nil || *job.Context.Username() != username {
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
