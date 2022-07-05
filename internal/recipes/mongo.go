// (c) Copyright IBM Corp. 2022

package recipes

import (
	"github.com/instana/go-instana/internal/registry"
	"go/ast"
	"go/token"
)

func init() {
	registry.Default.Register("go.mongodb.org/mongo-driver/mongo", NewMongo())
}

func NewMongo() *Mongo {
	return &Mongo{InstanaPkg: "instamongo"}
}

// Mongo instruments go.mongodb.org/mongo-driver/mongo package with Instana
type Mongo struct {
	InstanaPkg    string
	defaultRecipe defaultRecipe
}

// ImportPath returns instrumentation import path
func (recipe *Mongo) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instamongo"
}

// Instrument applies recipe to the ast Node
func (recipe *Mongo) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (changed bool) {
	return recipe.defaultRecipe.instrument(fset, f, targetPkg, sensorVar, recipe.InstanaPkg, recipe.ImportPath(), map[string]insertOption{
		"Connect":   {sensorPosition: 1},
		"NewClient": {},
	})
}
