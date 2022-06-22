// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package recipes

import (
	"github.com/instana/go-instana/registry"
	"go/ast"
	"go/token"
	"log"

	"golang.org/x/tools/go/ast/astutil"
)

func init() {
	recipe := NewDatabaseSQL()
	registry.Default.Register("db", recipe)
	registry.Default.Register("sql", recipe)
}

func NewDatabaseSQL() *DatabaseSQL {
	return &DatabaseSQL{InstanaPkg: "instana"}
}

// DatabaseSQL instruments database/sql package with Instana
type DatabaseSQL struct {
	InstanaPkg string
}

// ImportPath returns instrumentation import path
func (recipe *DatabaseSQL) ImportPath() string {
	return "github.com/instana/go-sensor"
}

// Instrument instruments sql.Open()
func (recipe *DatabaseSQL) Instrument(fset *token.FileSet, node ast.Node, targetPkg, sensorVar string) (changed bool) {
	astutil.Apply(node, func(c *astutil.Cursor) bool {
		return true
	}, func(c *astutil.Cursor) bool {
		switch node := c.Node().(type) {
		case *ast.CallExpr:
			changed = recipe.instrumentMethodCall(node, targetPkg) || changed
		}

		return true
	})

	return changed
}

func (recipe *DatabaseSQL) instrumentMethodCall(call *ast.CallExpr, targetPkg string) bool {
	pkg, fnName, ok := extractFunctionName(call)
	if !ok {
		return false
	}

	if pkg != targetPkg {
		return false
	}

	fn := call.Fun.(*ast.SelectorExpr)

	// replace sql.Open() with instana.SQLOpen() preserving the arguments
	if fnName == "Open" {
		log.Println("instrumenting database/sql.Open() at pos", call.Pos())

		fn.X = ast.NewIdent(recipe.InstanaPkg)
		fn.Sel = ast.NewIdent("SQLOpen")

		return true
	}

	return false
}
