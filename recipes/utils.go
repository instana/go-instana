// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package recipes

import "go/ast"

func extractFunctionName(call *ast.CallExpr) (string, string, bool) {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		switch selector := fn.X.(type) {
		case *ast.Ident:
			return selector.Name, fn.Sel.Name, true
		default:
			return "", "", false
		}
	case *ast.Ident:
		return "", fn.Name, true
	default:
		return "", "", false
	}
}

func extractSelectorPackageAndName(typ ast.Expr) (string, string, bool) {
	switch typ := typ.(type) {
	case *ast.SelectorExpr:
		if pkg, ok := typ.X.(*ast.Ident); ok {
			return pkg.Name, typ.Sel.Name, true
		}
	case *ast.StarExpr:
		return extractSelectorPackageAndName(typ.X)
	}

	return "", "", false
}

type stack[E any] []*E

func (s *stack[E]) Push(val *E) {
	*s = append(*s, val)
}

func (s *stack[E]) Top() *E {
	index := len(*s) - 1
	if index >= 0 {
		return (*s)[index]
	}

	return nil
}

func (s *stack[E]) Pop() *E {
	index := len(*s) - 1
	if index >= 0 {
		res := (*s)[index]
		*s = (*s)[:index]

		return res
	}

	return nil
}
