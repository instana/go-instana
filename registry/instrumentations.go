// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package registry

import (
	"go/ast"
	"go/token"
	"sync"
)

var Default = NewRegistry()

// NewRegistry returns Registry instance
func NewRegistry() *Registry {
	return &Registry{
		instrumentation: make(map[string]Recipe),
	}
}

// Registry is responsible for keeping mapping between packages and their instrumentation.
type Registry struct {
	mu              sync.Mutex
	instrumentation map[string]Recipe
}

// Register creates a mapping between targetPkg and instrumentation. It should be invoked with `init()` function
func (r *Registry) Register(targetPkg string, instrumentation Recipe) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.instrumentation[targetPkg] = instrumentation
}

// InstrumentationImportPath returns instrumentation import path for targetPkg, if any registered or empty string otherwise.
func (r *Registry) InstrumentationImportPath(targetPkg string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if instrumentation, ok := r.instrumentation[targetPkg]; ok {
		return instrumentation.ImportPath()
	}

	return ""
}

// InstrumentationRecipe returns recipe for the targetPkg if any registered.
func (r *Registry) InstrumentationRecipe(targetPkg string) Recipe {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.instrumentation[targetPkg]
}

type Instrumentation interface {
	// ImportPath returns instrumentation import path
	ImportPath() string
}

type Recipe interface {
	Instrument(fset *token.FileSet, f ast.Node, pkgName, sensorVar string) bool
	Instrumentation
}
