// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package recipes

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var verRegexp = regexp.MustCompile(`v\d+$`)

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

// GetPackageImportName extracts from the imports name of the import
func GetPackageImportName(fset *token.FileSet, f *ast.File, importPath string) (string, error) {
	for _, importGroup := range astutil.Imports(fset, f) {
		for _, importSpec := range importGroup {
			if importSpec.Path == nil {
				return "", errors.New("import path is <nil>")
			}

			// imports paths in the look like ""context.Context""
			rawPath, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				return "", fmt.Errorf("can not unquote import path:%w", err)
			}

			if rawPath == importPath {
				if importSpec.Name != nil {

					// check "special" import cases:
					// _
					// .
					if importSpec.Name.Name == "." || importSpec.Name.Name == "_" {
						return "", errors.New("does not support import as " + importSpec.Name.Name)
					}

					return importSpec.Name.Name, nil
				}

				return ExtractLocalImportName(importPath), nil
			}
		}
	}

	return "", errors.New("no import found for " + importPath)
}

func tryGetPackageImportName(fset *token.FileSet, f *ast.File, importPath string) string {
	v, err := GetPackageImportName(fset, f, importPath)
	if err != nil {
		log.Println(err)
		return ""
	}

	return v
}

// ExtractLocalImportName returns last part of the import path ignoring a version suffix
func ExtractLocalImportName(impPath string) string {
	localName := path.Base(impPath)
	if verRegexp.MatchString(localName) {
		imp := strings.Split(impPath, "/")
		if len(imp) > 1 {
			localName = imp[len(imp)-2]
		}
	}

	return localName
}
