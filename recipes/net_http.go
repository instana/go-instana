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
	registry.Default.Register("net/http", NewNetHTTP())
}

func NewNetHTTP() *NetHTTP {
	return &NetHTTP{
		InstanaPkg: "instana",
	}
}

// NetHTTP instruments net/http package with Instana
type NetHTTP struct {
	InstanaPkg string
}

// ImportPath returns instrumentation import path
func (recipe *NetHTTP) ImportPath() string {
	return "github.com/instana/go-sensor"
}

// Instrument instruments net/http.HandleFunc and net/http.Handle calls as well as (http.Client).Transport
func (recipe *NetHTTP) Instrument(fset *token.FileSet, node ast.Node, targetPkg, sensorVar string) (result ast.Node, changed bool) {
	return astutil.Apply(node, func(c *astutil.Cursor) bool {
		return true
	}, func(c *astutil.Cursor) bool {
		switch node := c.Node().(type) {
		case *ast.CallExpr:
			changed = recipe.instrumentMethodCall(node, targetPkg, sensorVar) || changed
		case *ast.CompositeLit:
			changed = recipe.instrumentCompositeLit(node, targetPkg, sensorVar) || changed
		}

		return true
	}), changed
}

func (recipe *NetHTTP) instrumentMethodCall(call *ast.CallExpr, targetPkg, sensorVar string) bool {
	pkgName, fnName, ok := extractFunctionName(call)
	if !ok {
		return false
	}

	if pkgName != targetPkg {
		return false
	}

	switch fnName {
	case "HandleFunc":
		handler := call.Args[1]

		// Double instrumentation check: handler is not an already instrumented http.HandlerFunc?
		if _, ok := assertFunctionName(handler, recipe.InstanaPkg, "TracingHandlerFunc"); ok {
			log.Println("skipping an already instrumented call to net/http.HandleFunc() at pos", call.Pos())

			return false
		}

		log.Println("instrumenting net/http.HandleFunc() call at pos", call.Pos())

		recipe.instrumentHandleFunc(call, handler, sensorVar)

		return true
	case "Handle":
		handler := call.Args[1]

		// Double instrumentation check: handler is not an already instrumented http.HandlerFunc?
		if call, ok := assertFunctionName(handler, targetPkg, "HandlerFunc"); ok {
			if len(call.Args) > 0 {
				if _, ok := assertFunctionName(call.Args[0], recipe.InstanaPkg, "TracingHandlerFunc"); ok {
					log.Println("skipping an already instrumented call to net/http.Handle() at pos", call.Pos())

					return false
				}
			}
		}

		log.Println("instrumenting net/http.Handle() call at pos", call.Pos())

		// Replace http.Handle with http.HandlerFunc, since instana.TracingHandleFunc() returns
		// a function instead of http.Handler
		call.Fun.(*ast.SelectorExpr).Sel.Name = "HandleFunc"
		recipe.instrumentHandleFunc(call, &ast.SelectorExpr{
			X:   handler,
			Sel: ast.NewIdent("ServeHTTP"),
		}, sensorVar)

		return true
	default:
		return false
	}
}

func (recipe *NetHTTP) instrumentHandleFunc(call *ast.CallExpr, handler ast.Expr, sensorVar string) {
	call.Args[1] = &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(recipe.InstanaPkg),
			Sel: ast.NewIdent("TracingHandlerFunc"),
		},
		Args: []ast.Expr{
			ast.NewIdent(sensorVar),
			call.Args[0], // pathTemplate
			handler,      // handler
		},
	}
}

func (recipe *NetHTTP) instrumentCompositeLit(lit *ast.CompositeLit, targetPkg, sensorVar string) bool {
	pkg, name, ok := extractSelectorPackageAndName(lit.Type)
	if !ok || pkg != targetPkg {
		return false
	}

	switch name {
	case "Client":
		// Check if this http.Client initializes its Transport already
		for _, el := range lit.Elts {
			kv, ok := el.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}

			if key.Name == "Transport" {
				// Double instrumentation check: is the transport already wrapped?
				if call, ok := kv.Value.(*ast.CallExpr); ok {
					if pkg, name, ok := extractFunctionName(call); ok && pkg == recipe.InstanaPkg && name == "RoundTripper" {
						log.Println("skipping an already instrumented (*http.Client).Transport at pos", kv.Value.Pos())

						return false
					}
				}

				log.Println("instrumenting (*http.Client).Transport at pos", kv.Value.Pos())
				kv.Value = recipe.instrumentTransport(kv.Value, sensorVar)

				return true
			}
		}

		// Initialize (http.Client).Transport otherwise with instana.RoundTripper
		log.Println("instrumenting *http.Client at pos", lit.Pos())
		lit.Elts = append(lit.Elts, &ast.KeyValueExpr{
			Key:   ast.NewIdent("Transport"),
			Value: recipe.instrumentTransport(ast.NewIdent("nil"), sensorVar),
		})

		return true
	}

	return false
}

func (recipe *NetHTTP) instrumentTransport(orig ast.Expr, sensorVar string) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(recipe.InstanaPkg),
			Sel: ast.NewIdent("RoundTripper"),
		},
		Args: []ast.Expr{
			ast.NewIdent(sensorVar),
			orig,
		},
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
