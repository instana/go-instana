// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package recipes

import (
	"github.com/instana/go-instana/internal/registry"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
)

func init() {
	registry.Default.Register("google.golang.org/grpc", NewGRPC())
}

func NewGRPC() *GRPC {
	return &GRPC{InstanaPkg: "instagrpc"}
}

// GRPC instruments google.golang.org/grpc package with Instana
type GRPC struct {
	InstanaPkg string
}

// ImportPath returns instrumentation import path
func (recipe *GRPC) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instagrpc"
}

func (recipe *GRPC) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (changed bool) {
	astutil.Apply(f,
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

	return changed
}

func (recipe *GRPC) instrumentMethodCall(call *ast.CallExpr, targetPkg, sensorVar string) bool {
	pkgName, fnName, ok := extractFunctionName(call)
	if !ok {
		return false
	}

	if pkgName != targetPkg {
		return false
	}

	switch fnName {
	case "NewServer":
		if recipe.argumentsAlreadyInstrumented(call.Args, sensorVar) {
			return false
		}

		originalArgsLen := len(call.Args)
		call.Args = append([]ast.Expr{
			recipe.targetCallExpr(targetPkg, "ChainStreamInterceptor", recipe.instrumentationCallExpr("StreamServerInterceptor", sensorVar)),
			recipe.targetCallExpr(targetPkg, "ChainUnaryInterceptor", recipe.instrumentationCallExpr("UnaryServerInterceptor", sensorVar)),
		}, call.Args...)

		return originalArgsLen != len(call.Args)
	case "Dial":
		if recipe.argumentsAlreadyInstrumented(call.Args, sensorVar) {
			return false
		}

		originalArgsLen := len(call.Args)

		if originalArgsLen == 0 {
			return false
		}

		call.Args = append([]ast.Expr{call.Args[0]}, append([]ast.Expr{
			recipe.targetCallExpr(targetPkg, "WithChainStreamInterceptor", recipe.instrumentationCallExpr("StreamClientInterceptor", sensorVar)),
			recipe.targetCallExpr(targetPkg, "WithChainUnaryInterceptor", recipe.instrumentationCallExpr("UnaryClientInterceptor", sensorVar)),
		}, call.Args[1:]...)...)

		return originalArgsLen != len(call.Args)
	default:
		return false
	}
}

// argumentsAlreadyInstrumented returns two parameters
func (recipe *GRPC) argumentsAlreadyInstrumented(args []ast.Expr, sensorVar string) bool {
	sensorFound := false
	for index := range args {
		ast.Inspect(args[index], func(node ast.Node) bool {
			if ident, ok := (node).(*ast.Ident); ok {
				if ident.Name == sensorVar {
					sensorFound = true

					return false
				}
			}

			return true
		})

		if sensorFound {
			return true
		}
	}

	return false
}

// generate grpcExpr expression
func (recipe *GRPC) targetCallExpr(targetPkg, funcName string, expr ast.Expr) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(targetPkg),
			Sel: ast.NewIdent(funcName),
		},
		Args: []ast.Expr{
			expr,
		},
	}
}

// generate instana instrumentation expression
func (recipe *GRPC) instrumentationCallExpr(funcName, sensorVar string) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(recipe.InstanaPkg),
			Sel: ast.NewIdent(funcName),
		},
		Args: []ast.Expr{
			ast.NewIdent(sensorVar),
		},
	}
}
