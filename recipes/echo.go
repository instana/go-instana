// (c) Copyright IBM Corp. 2022

package recipes

import (
	"github.com/instana/go-instana/registry"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
)

func init() {
	registry.Default.Register("github.com/labstack/echo/v4", NewEcho())
}

func NewEcho() *Echo {
	return &Echo{InstanaPkg: "instaecho"}
}

// Echo instruments github.com/labstack/echo/v4 package with Instana
type Echo struct {
	InstanaPkg string
}

// ImportPath returns instrumentation import path
func (recipe *Echo) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instaecho"
}

// Instrument applies recipe to the ast Node
func (recipe *Echo) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (result ast.Node, changed bool) {
	result = astutil.Apply(f,
		func(c *astutil.Cursor) bool {
			return true
		},
		func(c *astutil.Cursor) bool {
			switch node := c.Node().(type) {
			case *ast.CallExpr:
				changed = recipe.instrumentMethodCall(node, targetPkg, sensorVar) || changed
			}

			return true
		},
	)

	if changed {
		if val, ok := f.(*ast.File); ok {
			log.Printf("AddNamedImport: %s %s", recipe.InstanaPkg, recipe.ImportPath())
			astutil.AddNamedImport(fset, val, recipe.InstanaPkg, recipe.ImportPath())
		}
	}

	return result, changed
}

func (recipe *Echo) instrumentMethodCall(call *ast.CallExpr, targetPkg, sensorVar string) bool {
	pkgName, fnName, ok := extractFunctionName(call)
	if !ok {
		return false
	}

	if pkgName != targetPkg {
		return false
	}

	if fnName != "New" {
		return false
	}

	*call = ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(recipe.InstanaPkg),
			Sel: ast.NewIdent("New"),
		},
		Args: []ast.Expr{
			ast.NewIdent(sensorVar),
		},
	}

	return true
}
