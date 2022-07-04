// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package recipes

import (
	"github.com/instana/go-instana/internal/registry"
	"go/ast"
	"go/token"
)

func init() {
	recipe := NewDatabaseSQL()
	registry.Default.Register("database/sql", recipe)
}

func NewDatabaseSQL() *DatabaseSQL {
	return &DatabaseSQL{InstanaPkg: "instana"}
}

// DatabaseSQL instruments database/sql package with Instana
type DatabaseSQL struct {
	InstanaPkg    string
	defaultRecipe defaultRecipe
}

// ImportPath returns instrumentation import path
func (recipe *DatabaseSQL) ImportPath() string {
	return "github.com/instana/go-sensor"
}

// Instrument instruments sql.Open()
func (recipe *DatabaseSQL) Instrument(fset *token.FileSet, node ast.Node, targetPkg, sensorVar string) (changed bool) {
	return recipe.defaultRecipe.instrument(fset, node, targetPkg, sensorVar, recipe.InstanaPkg, recipe.ImportPath(), map[string]insertOption{
		"Open": {functionName: "SQLInstrumentAndOpen"},
	})
}
