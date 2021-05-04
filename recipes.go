package main

import (
	"go/ast"
	"log"

	"golang.org/x/tools/go/ast/astutil"
)

type Instrumenter interface {
	Instrument(ast.Node) (ast.Node, bool)
}

// NetHTTPRecipe instruments net/http package with Instana
type NetHTTPRecipe struct {
	InstanaPkg string
	TargetPkg  string
	SensorVar  string
}

// Instrument instruments net/http.HandleFunc and net/http.Handle calls
func (recipe NetHTTPRecipe) Instrument(node ast.Node) (result ast.Node, changed bool) {
	result = astutil.Apply(node, func(c *astutil.Cursor) bool {
		call, ok := c.Node().(*ast.CallExpr)
		if !ok {
			return true
		}

		pkgName, fnName, ok := extractFunctionName(call)
		if !ok {
			log.Printf("failed to extract function name from %#v", call)
			return true
		}

		if pkgName != recipe.TargetPkg {
			return true
		}

		switch fnName {
		case "HandleFunc":
			handler := call.Args[len(call.Args)-1]
			call.Args[len(call.Args)-1] = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent(recipe.InstanaPkg),
					Sel: ast.NewIdent("TracingHandlerFunc"),
				},
				Args: []ast.Expr{
					ast.NewIdent(recipe.SensorVar),
					call.Args[0], // pathTemplate
					handler,      // handler
				},
			}

			changed = true
		case "Handle":
			handler := call.Args[len(call.Args)-1]
			call.Args[len(call.Args)-1] = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent(recipe.TargetPkg),
					Sel: ast.NewIdent("HandlerFunc"),
				},
				Args: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent(recipe.InstanaPkg),
							Sel: ast.NewIdent("TracingHandlerFunc"),
						},
						Args: []ast.Expr{
							ast.NewIdent(recipe.SensorVar),
							call.Args[0], // pathTemplate
							&ast.SelectorExpr{
								X:   handler,
								Sel: ast.NewIdent("ServeHTTP"),
							}, // wrap handler's ServeHTTP() method
						},
					},
				},
			}

			changed = true
		}

		return true
	}, nil)

	return result, changed
}
