// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package recipes

import (
	"go/ast"
	"log"

	"golang.org/x/tools/go/ast/astutil"
)

// DatabaseSQL instruments database/sql package with Instana
type DatabaseSQL struct {
	TargetPkg  string
	InstanaPkg string
	SensorVar  string
}

// Instrument instruments sql.Open()
func (recipe DatabaseSQL) Instrument(node ast.Node) (result ast.Node, changed bool) {
	return astutil.Apply(node, func(c *astutil.Cursor) bool {
		return true
	}, func(c *astutil.Cursor) bool {
		switch node := c.Node().(type) {
		case *ast.CallExpr:
			changed = recipe.instrumentMethodCall(node) || changed
		}

		return true
	}), changed
}

func (recipe DatabaseSQL) instrumentMethodCall(call *ast.CallExpr) bool {
	pkg, fnName, ok := extractFunctionName(call)
	if !ok {
		return false
	}

	if pkg != recipe.TargetPkg {
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
