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
	registry.Default.Register("go.mongodb.org/mongo-driver/mongo", NewMongo())
}

func NewMongo() *Mongo {
	return &Mongo{InstanaPkg: "instamongo"}
}

// Mongo instruments go.mongodb.org/mongo-driver/mongo package with Instana
type Mongo struct {
	InstanaPkg string
}

// ImportPath returns instrumentation import path
func (recipe *Mongo) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instamongo"
}

// Instrument applies recipe to the ast Node
func (recipe *Mongo) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (result ast.Node, changed bool) {
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

func (recipe *Mongo) instrumentMethodCall(call *ast.CallExpr, targetPkg, sensorVar string) bool {
	pkgName, fnName, ok := extractFunctionName(call)
	if !ok {
		return false
	}

	if pkgName != targetPkg {
		return false
	}

	switch fnName {
	case "Connect":
		var args []ast.Expr
		i := 0
		for i < len(call.Args) {
			args = append(args, call.Args[i])
			i++
			if i == 1 {
				args = append(args, ast.NewIdent(sensorVar))
			}
		}

		*call = ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(recipe.InstanaPkg),
				Sel: ast.NewIdent("Connect"),
			},
			Args: args,
		}

		return true

	case "NewClient":
		var args []ast.Expr
		args = append(args, ast.NewIdent(sensorVar))
		args = append(args, call.Args...)

		*call = ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(recipe.InstanaPkg),
				Sel: ast.NewIdent("NewClient"),
			},
			Args: args,
		}

		return true
	}

	return false
}
