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

func (recipe NetHTTPRecipe) instrumentMethodCall(call *ast.CallExpr) bool {
	pkgName, fnName, ok := extractFunctionName(call)
	if !ok {
		log.Printf("failed to extract function name from %#v", call)
		return false
	}

	if pkgName != recipe.TargetPkg {
		return false
	}

	switch fnName {
	case "HandleFunc":
		handler := call.Args[len(call.Args)-1]

		// Double instrumentation check: handler is not an already insturmented http.HandlerFunc?
		if _, ok := assertFunctionName(handler, recipe.InstanaPkg, "TracingHandlerFunc"); ok {
			log.Println("skipping an already instrumented call to net/http.HandleFunc() at pos", call.Pos())

			return false
		}

		log.Println("instrumenting net/http.HandleFunc() call at pos", call.Pos())

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

		return true
	case "Handle":
		handler := call.Args[len(call.Args)-1]

		// Double instrumentation check: handler is not an already insturmented http.HandlerFunc?
		if call, ok := assertFunctionName(handler, recipe.TargetPkg, "HandlerFunc"); ok {
			if len(call.Args) > 0 {
				if _, ok := assertFunctionName(call.Args[0], recipe.InstanaPkg, "TracingHandlerFunc"); ok {
					log.Println("skipping an already instrumented call to net/http.Handle() at pos", call.Pos())

					return false
				}
			}
		}

		log.Println("instrumenting net/http.Handle() call at pos", call.Pos())

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

		return true
	default:
		return false
	}
}

func assertFunctionName(node ast.Expr, pkg, fn string) (*ast.CallExpr, bool) {
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return nil, false
	}

	fnPkg, fnName, ok := extractFunctionName(call)

	return call, ok && fnPkg == pkg && fnName == fn
}
