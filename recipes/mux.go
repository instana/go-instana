// (c) Copyright IBM Corp. 2022

package recipes

import (
	"github.com/instana/go-instana/registry"
	"go/ast"
	"go/token"
)

func init() {
	registry.Default.Register("github.com/gorilla/mux", NewMux())
}

// NewMux returns Mux recipe
func NewMux() *Mux {
	return &Mux{InstanaPkg: "instamux", defaultRecipe: defaultRecipe{}}
}

// Mux instruments github.com/gorilla/mux package with Instana
type Mux struct {
	InstanaPkg    string
	defaultRecipe defaultRecipe
}

// ImportPath returns instrumentation import path
func (recipe *Mux) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instamux"
}

// Instrument applies recipe to the ast Node
func (recipe *Mux) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (result ast.Node, changed bool) {
	return recipe.defaultRecipe.instrument(fset, f, targetPkg, sensorVar, recipe.InstanaPkg, recipe.ImportPath(), map[string]insertOption{
		"NewRouter": {},
	})
}
