package homescript

import (
	"sync"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/compiler"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	dispTypes "github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
)

// TODO: all used singletons should also be hashed (in case of drivers at least)

type ManagerCompileCache struct {
	Cache map[string]compiler.CompileOutput
	Lock  sync.RWMutex
}

func newManagerCompileCache() ManagerCompileCache {
	return ManagerCompileCache{
		Cache: make(map[string]compiler.CompileOutput),
		Lock:  sync.RWMutex{},
	}
}

func (m *Manager) Compile(
	modules map[string]ast.AnalyzedProgram,
	entryPointModule string,
	username string,
) compiler.CompileOutput {
	// Try to use a cached version.
	m.CompileCache.Lock.RLock()
	cached, valid := m.CompileCache.Cache[entryPointModule]
	m.CompileCache.Lock.RUnlock()

	if valid {
		logger.Tracef("Using compilation cache for program `%s`...\n", entryPointModule)
		return cached
	}

	logger.Tracef("Compiling program (invalid cache) `%s`...\n", entryPointModule)

	comp := compiler.NewCompiler(modules, entryPointModule)
	compOut := comp.Compile()

	m.CompileCache.Lock.Lock()
	m.CompileCache.Cache[entryPointModule] = compOut
	m.CompileCache.Lock.Unlock()

	//
	// Register any trigger annotations OF THE MAIN MODULE.
	// Only process annotations if this is a persistent file (live scripts not allowed).
	// This also shuld only trigger once, meaning after the hash of the data changed ONCE.
	//

	for key, annotationList := range compOut.Annotations {
		// If the module is the entrypoint, process this annotation.
		if key.Module == entryPointModule {
			for _, annotation := range annotationList.Items {
				if annotation.Kind() == compiler.CompiledAnnotationKindTrigger {
					trigger := annotation.(compiler.TriggerCompiledAnnotation)

					callmodeAdaptive := dispTypes.CallMode(dispTypes.CallModeAdaptive{
						Username: username,
					})

					if _, err := registerTriggerOverride(
						trigger.CallbackFnIdent,
						trigger.TriggerSource,
						errors.Span{}, // TODO: use real span.
						trigger.TriggerArgs,
						&callmodeAdaptive,
						username,
						entryPointModule,
						nil, // During compilation, this program does not have a job ID yet.
					); err != nil {
						logger.Warn(err.Error())
						// TODO: maybe fail the entire compilation?
					}

					// panic("TODO: handle trigger annotation")
				}
			}
		}
	}

	//
	// End trigger annotations.
	//

	return compOut
}
