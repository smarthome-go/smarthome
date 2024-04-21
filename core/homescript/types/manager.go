package types

import (
	"context"
	"io"

	"github.com/smarthome-go/homescript/v3/homescript"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	driverTypes "github.com/smarthome-go/smarthome/core/device/driver/types"
)

type Job struct {
	Context   ExecutionContext
	JobID     uint64
	HmsID     string
	CancelCtx context.CancelFunc
	VM        *runtime.VM
}

type ProgramInvocation struct {
	Identifier         homescript.InputProgram
	FunctionInvocation *runtime.FunctionInvocation
	SingletonsToLoad   map[string]value.Value
}

type Cancelation struct {
	Context    context.Context
	CancelFunc context.CancelFunc
}

type Manager interface {
	GetPersonalScriptById(homescriptID string, username string) (database.Homescript, bool, error)

	Analyze(
		input homescript.InputProgram,
		context ExecutionContext,
	) (map[string]ast.AnalyzedProgram, HmsDiagnosticsContainer, error)

	AnalyzeUserScript(
		context ExecutionContextUser,
	) (map[string]ast.AnalyzedProgram, HmsDiagnosticsContainer, error)

	// TODO: create functions which load the source code (and required metadata) based on an execution context.

	AnalyzeDriver(
		context ExecutionContextDriver,
	) (map[string]ast.AnalyzedProgram, HmsDiagnosticsContainer, error)

	RunGeneric(
		invocation ProgramInvocation,
		context ExecutionContext,
		cancelation Cancelation,
		// This is required for the asynchronous runtime.
		idChan *chan uint64,
		outputWriter io.Writer,
	) (HmsRes, error)

	RunUserScript(
		programID, username string,
		function *runtime.FunctionInvocation,
		cancelation Cancelation,
		outputWriter io.Writer,
		idChan *chan uint64,
	) (HmsRes, error)

	RunDriverScript(
		driverIDs driverTypes.DriverInvocationIDs,
		invocation runtime.FunctionInvocation,
		cancelation Cancelation,
		outputWriter io.Writer,
	) (HmsRes, error)

	GetJobList() []Job
	GetJobById(jobID uint64) (Job, bool)
	KillAllId(hmsID string) (count uint64, success bool)

	InvalidateCompileCacheEntry(programID string)
}
