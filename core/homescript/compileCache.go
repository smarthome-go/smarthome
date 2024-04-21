package homescript

import (
	"sync"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/compiler"
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
) (compiler.CompileOutput, error) {
	// Try to use a cached version.
	m.CompileCache.Lock.RLock()
	cached, valid := m.CompileCache.Cache[entryPointModule]
	m.CompileCache.Lock.RUnlock()

	if valid {
		logger.Tracef("Using compilation cache for program `%s`...\n", entryPointModule)
		return cached, nil
	}

	// Otherwise, trigger a rebuild.
	logger.Tracef("Compiling program (invalid cache) `%s`...\n", entryPointModule)

	comp := compiler.NewCompiler(modules, entryPointModule)
	compOut, err := comp.Compile()
	if err != nil {
		return compiler.CompileOutput{}, err
	}

	m.CompileCache.Lock.Lock()
	m.CompileCache.Cache[entryPointModule] = compOut
	m.CompileCache.Lock.Unlock()

	return compOut, nil
}

func (m *Manager) InvalidateCompileCacheEntry(programID string) {
	m.CompileCache.Lock.Lock()
	delete(m.CompileCache.Cache, programID)
	m.CompileCache.Lock.Unlock()
}
