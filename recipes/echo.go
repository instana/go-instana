// (c) Copyright IBM Corp. 2022

package recipes

import (
	"github.com/instana/go-instana/registry"
	"go/ast"
	"go/token"
)

func init() {
	registry.Default.Register("github.com/labstack/echo/v4", NewEcho())
}

func NewEcho() *Echo {
	return &Echo{InstanaPkg: "instaecho", defaultRecipe: defaultRecipe{}}
}

// Echo instruments github.com/labstack/echo/v4 package with Instana
type Echo struct {
	InstanaPkg    string
	defaultRecipe defaultRecipe
}

// ImportPath returns instrumentation import path
func (recipe *Echo) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instaecho"
}

// Instrument applies recipe to the ast Node
func (recipe *Echo) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) bool {
	return recipe.defaultRecipe.instrument(fset, f, targetPkg, sensorVar, recipe.InstanaPkg, recipe.ImportPath(), map[string]insertOption{
		"New": {},
	})
}
