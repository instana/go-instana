// (c) Copyright IBM Corp. 2022

package recipes

import (
	"go/ast"
	"go/token"

	"github.com/instana/go-instana/registry"
	"golang.org/x/tools/go/ast/astutil"
)

func init() {
	registry.Default.Register("github.com/julienschmidt/httprouter", NewGin())
}

// NewHttpRouter returns the HttpRouter recipe
func NewHttpRouter() *HttpRouter {
	return &HttpRouter{InstanaPkg: "instahttprouter", defaultRecipe: defaultRecipe{}}
}

// HttpRouter instruments the github.com/julienschmidt/httprouter package with Instana
type HttpRouter struct {
	InstanaPkg    string
	defaultRecipe defaultRecipe
}

// ImportPath returns the instrumentation import path
func (recipe *HttpRouter) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instahttprouter"
}

// Instrument applies the recipe to the ast Node
func (recipe *HttpRouter) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (result ast.Node, changed bool) {

	astutil.Apply(f, func(c *astutil.Cursor) bool {
		if c.Node() == nil {
			return false
		}

		return true
	}, func(c *astutil.Cursor) bool {

		switch node := c.Node().(type) {
		case *ast.SelectorExpr:
			nodeX, ok := node.X.(*ast.Ident)

			// We look for `var something *httprouter.Router` and replace by `var something *instahttprouter.WrappedRouter`

			if ok && nodeX.Name == "httprouter" && node.Sel.Name == "Router" {
				nodeX.Name = "instahttprouter"
				node.Sel.Name = "WrappedRouter"
			}
		case *ast.CallExpr:
			fn := node.Fun.(*ast.SelectorExpr)
			fnX := fn.X.(*ast.Ident)

			// Replacing httprouter.New() by instahttprouter.Wrap(httprouter.New(), __instanaSensor)

			if fnX.Name == "httprouter" {
				node.Args = append(node.Args, []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: "httprouter.New()",
					},
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: "__instanaSensor",
					},
				}...)
				fnX.Name = "instahttprouter"
				fn.Sel.Name = "Wrap"
			}
			// fmt.Printf(">>> POST: type: %v, %v\n", fnX.Name, fn.Sel.Name)
		default:
			// fmt.Printf(">>> NOT CALL EXPR: %T : %v\n", c.Node(), c.Node())
		}

		return true
	})

	return recipe.defaultRecipe.instrument(fset, f, targetPkg, sensorVar, recipe.InstanaPkg, recipe.ImportPath(), map[string]struct{}{
		"New": {},
	})
}
