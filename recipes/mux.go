// (c) Copyright IBM Corp. 2022

package recipes

import (
	"github.com/instana/go-instana/internal"
	"github.com/instana/go-instana/registry"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
)

func init() {
	registry.Default.Register("github.com/gorilla/mux", NewMux())
}

// NewMux returns Mux recipe
func NewMux() *Mux {
	return &Mux{InstanaPkg: "instamux"}
}

// Mux instruments github.com/gorilla/mux package with Instana
type Mux struct {
	InstanaPkg string
}

// ImportPath returns instrumentation import path
func (recipe *Mux) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instamux"
}

// Instrument applies recipe to the ast Node
func (recipe *Mux) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (ast.Node, bool) {
	// blocks are used to track current block while traversing a tree
	var blocks internal.BlockStmtStack

	var changed bool

	result := astutil.Apply(f,
		func(c *astutil.Cursor) bool {
			switch node := c.Node().(type) {
			case *ast.BlockStmt:
				blocks.Push(node)
			case *ast.CallExpr:
				if recipe.eligibleForInstrumentation(node, targetPkg) {
					switch val := c.Parent().(type) {
					case *ast.ValueSpec:
						if len(val.Names) != 0 {
							changed = internal.InsertInBlockOnce(blocks.Top(), node, recipe.createInstrumentationStatement(sensorVar, val.Names[0]))
						}
					case *ast.AssignStmt:
						if len(val.Lhs) == 1 {
							if ident, ok := (val.Lhs[0]).(*ast.Ident); ok {
								changed = internal.InsertInBlockOnce(blocks.Top(), node, recipe.createInstrumentationStatement(sensorVar, ident))
							}
						}
					}
				}
			}

			return true
		},
		func(c *astutil.Cursor) bool {
			if _, ok := (c.Node()).(*ast.BlockStmt); ok {
				blocks.Pop()
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

// eligibleForInstrumentation checks if `call` is the new Mux router creation method invocation
func (recipe *Mux) eligibleForInstrumentation(call *ast.CallExpr, targetPkg string) bool {
	pkgName, fnName, ok := extractFunctionName(call)

	if !ok {
		return false
	}

	if pkgName != targetPkg {
		return false
	}

	if fnName != "NewRouter" {
		return false
	}

	return true
}

// createInstrumentationStatement returns instrumentation statements to insert after the Mux router creation statement
func (recipe *Mux) createInstrumentationStatement(sensorVar string, routerVarName *ast.Ident) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(recipe.InstanaPkg),
				Sel: ast.NewIdent("AddMiddleware"),
			},
			Args: []ast.Expr{
				ast.NewIdent(sensorVar),
				routerVarName,
			},
		},
	}
}
