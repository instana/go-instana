// (c) Copyright IBM Corp. 2022

package recipes

import (
	"github.com/instana/go-instana/internal/registry"
	"go/ast"
	"go/token"
	"log"

	"golang.org/x/tools/go/ast/astutil"
)

func init() {
	registry.Default.Register("github.com/julienschmidt/httprouter", NewHttpRouter())
}

// NewHttpRouter returns the HttpRouter recipe
func NewHttpRouter() *HttpRouter {
	return &HttpRouter{InstanaPkg: "instahttprouter"}
}

// HttpRouter instruments the github.com/julienschmidt/httprouter package with Instana
type HttpRouter struct {
	InstanaPkg string
}

// ImportPath returns the instrumentation import path
func (recipe *HttpRouter) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instahttprouter"
}

// Instrument applies the recipe to the ast Node
func (recipe *HttpRouter) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (changed bool) {
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		return c.Node() != nil
	}, func(c *astutil.Cursor) bool {
		switch node := c.Node().(type) {
		// We look for `var something *httprouter.Router` and replace by `var something *instahttprouter.WrappedRouter`
		case *ast.SelectorExpr:
			nodeX, ok := node.X.(*ast.Ident)

			if ok && nodeX.Name == "httprouter" && node.Sel.Name == "Router" {
				nodeX.Name = recipe.InstanaPkg
				node.Sel.Name = "WrappedRouter"
				changed = true
			}

		// Replacing httprouter.New() by instahttprouter.Wrap(httprouter.New(), __instanaSensor)
		case *ast.CallExpr:
			libPkg, libFunction, _ := extractFunctionName(node)

			if libPkg != "httprouter" || libFunction != "New" {
				return true
			}

			// If httprouter.New() is an argument of instahttprouter.Wrap(), it is already instrumented
			if parent, ok := c.Parent().(*ast.CallExpr); ok {
				instanaPkg, instanaFunction, found := extractFunctionName(parent)
				if found && instanaPkg == "instahttprouter" && instanaFunction == "Wrap" {
					return true
				}
			}

			fn, ok := node.Fun.(*ast.SelectorExpr)

			if !ok {
				return true
			}

			fnX, ok := fn.X.(*ast.Ident)

			if !ok {
				return true
			}

			node.Args = []ast.Expr{
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: "httprouter.New()",
				},
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: sensorVar,
				},
			}
			fnX.Name = "instahttprouter"
			fn.Sel.Name = "Wrap"

			changed = true
		}

		return true
	})

	if changed {
		if val, ok := f.(*ast.File); ok {
			log.Printf("AddNamedImport: %s %s", recipe.InstanaPkg, recipe.ImportPath())
			astutil.AddNamedImport(fset, val, recipe.InstanaPkg, recipe.ImportPath())
		}
	}

	return changed
}
