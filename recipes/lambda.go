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
	registry.Default.Register("github.com/aws/aws-lambda-go/lambda", NewLambda())
}

func NewLambda() *Lambda {
	return &Lambda{InstanaPkg: "instalambda"}
}

// Lambda instruments github.com/aws/aws-lambda-go package with Instana
type Lambda struct {
	InstanaPkg string
}

// ImportPath returns instrumentation import path
func (recipe *Lambda) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instalambda"
}

func (recipe *Lambda) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (result ast.Node, changed bool) {
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

func (recipe *Lambda) instrumentMethodCall(call *ast.CallExpr, targetPkg, sensorVar string) bool {
	pkgName, fnName, ok := extractFunctionName(call)
	if !ok {
		return false
	}

	if pkgName != targetPkg {
		return false
	}

	switch fnName {
	case "Start":
		if recipe.argumentsAlreadyInstrumented(call.Args, sensorVar) {
			return false
		}

		call.Args[0] = recipe.instrumentationCallExpr("NewHandler", call.Args[0], sensorVar)

		return true

	case "StartHandler":
		if recipe.argumentsAlreadyInstrumented(call.Args, sensorVar) {
			return false
		}

		call.Args[0] = recipe.instrumentationCallExpr("WrapHandler", call.Args[0], sensorVar)

		return true
	case "StartHandlerWithContext":
		if recipe.argumentsAlreadyInstrumented(call.Args, sensorVar) {
			return false
		}

		call.Args[1] = recipe.instrumentationCallExpr("WrapHandler", call.Args[1], sensorVar)

		return true
	case "StartWithOptions":
		if recipe.argumentsAlreadyInstrumented(call.Args, sensorVar) {
			return false
		}

		call.Args[0] = recipe.instrumentationCallExpr("NewHandler", call.Args[0], sensorVar)

		return true
	case "StartWithContext":
		if recipe.argumentsAlreadyInstrumented(call.Args, sensorVar) {
			return false
		}

		call.Args[1] = recipe.instrumentationCallExpr("NewHandler", call.Args[1], sensorVar)

		return true
	default:
		return false
	}
}

func (recipe *Lambda) argumentsAlreadyInstrumented(args []ast.Expr, sensorVar string) bool {
	sensorFound := false
	for index := range args {
		ast.Inspect(args[index], func(node ast.Node) bool {
			if ident, ok := (node).(*ast.Ident); ok {
				if ident.Name == sensorVar {
					sensorFound = true

					return false
				}
			}

			return true
		})

		if sensorFound {
			return true
		}
	}

	return false
}

// generate instana instrumentation expression
func (recipe *Lambda) instrumentationCallExpr(funcName string, handler ast.Expr, sensorVar string) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(recipe.InstanaPkg),
			Sel: ast.NewIdent(funcName),
		},
		Args: []ast.Expr{
			handler,
			ast.NewIdent(sensorVar),
		},
	}
}
