// (c) Copyright IBM Corp. 2022

package recipes

import (
	"github.com/instana/go-instana/registry"
	"go/ast"
	"go/token"
)

func init() {
	registry.Default.Register("github.com/gin-gonic/gin", NewGin())
}

// NewGin returns Gin recipe
func NewGin() *Gin {
	return &Gin{InstanaPkg: "instagin", defaultRecipe: defaultRecipe{}}
}

// Gin instruments github.com/labstack/gin/v4 package with Instana
type Gin struct {
	InstanaPkg    string
	defaultRecipe defaultRecipe
}

// ImportPath returns instrumentation import path
func (recipe *Gin) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instagin"
}

// Instrument applies recipe to the ast Node
func (recipe *Gin) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (result ast.Node, changed bool) {
	return recipe.defaultRecipe.instrument(fset, f, targetPkg, sensorVar, recipe.InstanaPkg, recipe.ImportPath(), map[string]struct{}{
		"New":     {},
		"Default": {},
	})
}
