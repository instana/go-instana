// (c) Copyright IBM Corp. 2022

package recipes

import (
	"go/ast"
	"go/token"
	"log"

	"github.com/instana/go-instana/registry"
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
	// defaultRecipe defaultRecipe
}

// ImportPath returns the instrumentation import path
func (recipe *HttpRouter) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instahttprouter"
}

// Instrument applies the recipe to the ast Node
func (recipe *HttpRouter) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (result ast.Node, changed bool) {
	// panic(3)
	// ast.Print(fset, f)
	// fmt.Fprint(os.Stdout, f)
	// fmt.Printf("%T, %v", f, f)
	// panic(1)
	result = astutil.Apply(f, func(c *astutil.Cursor) bool {
		// panic(111)
		// if c.Node() == nil {
		// 	return false
		// }

		return true
	}, func(c *astutil.Cursor) bool {
		// panic(222)
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
			fn, ok := node.Fun.(*ast.SelectorExpr)

			if !ok {
				return true
			}

			fnX, ok := fn.X.(*ast.Ident)

			if !ok {
				return true
			}

			if fnX.Name == targetPkg {
				// *node = ast.CallExpr{
				// 	Fun: &ast.SelectorExpr{
				// 		X:   ast.NewIdent(recipe.InstanaPkg),
				// 		Sel: ast.NewIdent("LALALANOME_DA_FUNCAO"),
				// 	},
				// 	// Args: []ast.Expr{
				// 	// 	handler,
				// 	// 	ast.NewIdent(sensorVar),
				// 	// },
				// }
				// panic(666)
				node.Args = append(node.Args, []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: "httprouter.New()",
					},
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: sensorVar,
					},
				}...)
				fnX.Name = "instahttprouter"
				fn.Sel.Name = "Wrap"

				changed = true
			}
		}

		return true
	})

	if changed {
		if val, ok := f.(*ast.File); ok {
			log.Printf("AddNamedImport: %s %s", recipe.InstanaPkg, recipe.ImportPath())
			astutil.AddNamedImport(fset, val, recipe.InstanaPkg, recipe.ImportPath())
		}
	}

	// fmt.Fprint(os.Stdout, "will return: ", changed, "\n")

	return result, changed
}
