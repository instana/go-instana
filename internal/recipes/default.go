// (c) Copyright IBM Corp. 2022

package recipes

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
)

const firstInsertPosition = 0
const lastInsertPosition = -1

type defaultRecipe struct {
}

type insertOption struct {
	sensorPosition int
	functionName   string
}

// instrument applies recipe to the ast Node
func (recipe *defaultRecipe) instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar, instanaPkg, importPath string, methods map[string]insertOption) (changed bool) {
	astutil.Apply(f,
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
		addNamedImport(fset, f, instanaPkg, importPath)
	}

	return changed
}

func (recipe *defaultRecipe) instrumentMethodCall(call *ast.CallExpr, targetPkg, sensorVar, instanaPkg string, methods map[string]insertOption) bool {
	pkgName, fnName, ok := extractFunctionName(call)
	if !ok {
		return false
	}

	if pkgName != targetPkg {
		return false
	}

	if opt, ok := methods[fnName]; ok {
		args := call.Args
		ep := call.Ellipsis

		var newArgs []ast.Expr
		switch opt.sensorPosition {
		case firstInsertPosition:
			newArgs = append([]ast.Expr{
				ast.NewIdent(sensorVar),
			}, args...)
		case lastInsertPosition:
			newArgs = append(args, ast.NewIdent(sensorVar))
		default:
			index := opt.sensorPosition

			newArgs = append(args[:index+1], args[index:]...)
			newArgs[index] = ast.NewIdent(sensorVar)
		}

		if opt.functionName != "" {
			fnName = opt.functionName
		}

		*call = ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(instanaPkg),
				Sel: ast.NewIdent(fnName),
			},
			Args:     newArgs,
			Ellipsis: ep,
		}

		return true
	}

	return false
}
