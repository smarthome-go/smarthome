package types

import (
	"context"
	"io"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	driverTypes "github.com/smarthome-go/smarthome/core/device/driver/types"
)

type ProgramInvocation struct {
	ProgramID string
	DriverIDs *driverTypes.DriverInvocationIDs
}

type Job struct {
	Username        string
	JobID           uint64
	HmsID           *string
	Initiator       HomescriptInitiator
	CancelCtx       context.CancelFunc
	VM              *runtime.VM
	EntryModuleName string
	SupportsKill    bool
}

type Manager interface {
	GetPersonalScriptById(homescriptID string, username string) (database.Homescript, bool, error)

	Analyze(
		username string,
		filename string,
		code string,
		programKind HMS_PROGRAM_KIND,
		driverData *driverTypes.DriverInvocationIDs,
	) (map[string]ast.AnalyzedProgram, HmsRes, error)

	AnalyzeById(
		id string,
		username string,
		programKind HMS_PROGRAM_KIND,
		driverData *driverTypes.DriverInvocationIDs,
	) (map[string]ast.AnalyzedProgram, HmsRes, error)

	Run(
		programKind HMS_PROGRAM_KIND,
		driverData *driverTypes.DriverInvocationIDs,
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
	) (HmsRes, HmsRunResultContext, error)

	RunById(
		programKind HMS_PROGRAM_KIND,
		driverData *driverTypes.DriverInvocationIDs,
		hmsID string,
		username string,
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
	) (HmsRes, HmsRunResultContext, error)

	GetJobList() []Job
	GetJobById(jobID uint64) (Job, bool)
	KillAllId(hmsID string) (count uint64, success bool)

	InvalidateCompileCacheEntry(ids ProgramInvocation)
}
