// (c) Copyright IBM Corp. 2022

package recipes

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
)

type defaultRecipe struct {
}

// instrument applies recipe to the ast Node
func (recipe *defaultRecipe) instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar, instanaPkg, importPath string, methods map[string]struct{}) (result ast.Node, changed bool) {
	result = astutil.Apply(f,
		func(c *astutil.Cursor) bool {
			return true
		},
		func(c *astutil.Cursor) bool {
			switch node := c.Node().(type) {
			case *ast.CallExpr:
				changed = recipe.instrumentMethodCall(node, targetPkg, sensorVar, instanaPkg, methods) || changed
			}

			return true
		},
	)

	if changed {
		if val, ok := f.(*ast.File); ok {
			log.Printf("AddNamedImport: %s %s", instanaPkg, importPath)
			astutil.AddNamedImport(fset, val, instanaPkg, importPath)
		}
	}

	return result, changed
}

func (recipe *defaultRecipe) instrumentMethodCall(call *ast.CallExpr, targetPkg, sensorVar, instanaPkg string, methods map[string]struct{}) bool {
	pkgName, fnName, ok := extractFunctionName(call)
	if !ok {
		return false
	}

	if pkgName != targetPkg {
		return false
	}

	if _, ok := methods[fnName]; ok {
		args := call.Args
		*call = ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(instanaPkg),
				Sel: ast.NewIdent(fnName),
			},
			Args: append([]ast.Expr{
				ast.NewIdent(sensorVar),
			}, args...),
		}
		return true
	}

	return false
}
