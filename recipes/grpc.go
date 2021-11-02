package recipes

import (
	"go/ast"
	"golang.org/x/tools/go/ast/astutil"
)

// GRPC instruments google.golang.org/grpc package with Instana
type GRPC struct {
	InstanaPkg string
	TargetPkg  string
	SensorVar  string
}

const UnaryInterceptor = "UnaryInterceptor"
const ChainUnaryInterceptor = "ChainUnaryInterceptor"

const StreamInterceptor = "StreamInterceptor"
const ChainStreamInterceptor = "ChainStreamInterceptor"

const UnaryServerInterceptor = "UnaryServerInterceptor"
const StreamServerInterceptor = "StreamServerInterceptor"

func (recipe GRPC) Instrument(node ast.Node) (result ast.Node, changed bool) {
	return astutil.Apply(node,
		func(c *astutil.Cursor) bool {
			return true
		},
		func(c *astutil.Cursor) bool {
			switch node := c.Node().(type) {
			case *ast.CallExpr:
				changed = recipe.instrumentMethodCall(node) || changed
			}

			return true
		}), changed
}

func (recipe GRPC) instrumentMethodCall(call *ast.CallExpr) bool {
	pkgName, fnName, ok := extractFunctionName(call)
	if !ok {
		return false
	}

	if pkgName != recipe.TargetPkg {
		return false
	}

	switch fnName {
	case "NewServer":
		originalArgsLen := len(call.Args)
		call.Args = recipe.assertArguments(call.Args, UnaryInterceptor, ChainUnaryInterceptor, UnaryServerInterceptor)
		call.Args = recipe.assertArguments(call.Args, StreamInterceptor, ChainStreamInterceptor, StreamServerInterceptor)

		return originalArgsLen != len(call.Args)
	default:
		return false
	}
}

// checkIfArgumentsAlreadyInstrumented returns two parameters
// - if instrumentation detected
// - if any UnaryInterceptor found
func (recipe GRPC) checkIfArgumentsAlreadyInstrumented(args []ast.Expr, grpcFuncName, chainGRPCFuncName, instanaInstrumentationFunctionName string) (bool, bool) {
	var needToChain bool

	for index := range args {
		sr := &searchResult{
			interceptorFuncName:                grpcFuncName,
			chainInterceptorFuncName:           chainGRPCFuncName,
			instanaInstrumentationFunctionName: instanaInstrumentationFunctionName,
			sensorVarName:                      recipe.SensorVar,
		}

		recipe.checkNode(args[index], sr)
		needToChain = sr.needToChain || needToChain

		if sr.isFound() {
			return true, needToChain
		}
	}

	return false, needToChain
}

// checkNode goes through AST tree and tries to find artifacts that are signs of the instrumentation
func (recipe GRPC) checkNode(node interface{}, result *searchResult) {
	switch elem := node.(type) {
	case *ast.CallExpr:
		recipe.checkNode(elem.Fun, result)
		for index := range elem.Args {
			recipe.checkNode(elem.Args[index], result)
		}
	case *ast.Ident:
		if result.assertName(elem.Name) {
			return
		}

		if elem.Obj != nil {
			recipe.checkNode(elem.Obj.Decl, result)
		}
	case *ast.ValueSpec:
		for index := range elem.Values {
			recipe.checkNode(elem.Values[index], result)
		}
	case *ast.SelectorExpr:
		if pkgName, ok := (elem.X).(*ast.Ident); ok && pkgName.Name == recipe.TargetPkg {
			result.assertGRPCFuncName(elem.Sel.Name)
		}

		if pkgName, ok := (elem.X).(*ast.Ident); ok && pkgName.Name == recipe.InstanaPkg {
			result.assertInstanaInstrumentationFuncName(elem.Sel.Name)
		}

		if elem.Sel.Obj != nil {
			recipe.checkNode(elem.Sel.Obj.Decl, result)
		}
	case *ast.BinaryExpr:
		recipe.checkNode(elem.X, result)
		recipe.checkNode(elem.Y, result)
	case *ast.BasicLit:
		return
	case *ast.FuncLit:
		for key := range elem.Body.List {
			recipe.checkNode(elem.Body.List[key], result)
		}
	case *ast.ReturnStmt:
		for key := range elem.Results {
			recipe.checkNode(elem.Results[key], result)
		}
	case *ast.StarExpr:
		recipe.checkNode(elem.X, result)
	case *ast.UnaryExpr:
		recipe.checkNode(elem.X, result)
	case *ast.ParenExpr:
		recipe.checkNode(elem.X, result)
	case *ast.IndexExpr:
		recipe.checkNode(elem.X, result)
	case *ast.AssignStmt:
		recipe.checkNode(elem.Rhs, result)
	case []ast.Expr:
		for key := range elem {
			recipe.checkNode(elem[key], result)
		}
	case *ast.CompositeLit:
		for k := range elem.Elts {
			recipe.checkNode(elem.Elts[k], result)
		}
	case *ast.FuncDecl:
		for key := range elem.Body.List {
			recipe.checkNode(elem.Body.List[key], result)
		}
	default:
		//todo: log here. Panic is only for development
		panic("a")
	}
}

// assertArguments ensures that function call is instrumented
func (recipe GRPC) assertArguments(args []ast.Expr, grpcFuncName, chainGRPCFuncName, instanaInstrumentationFunctionName string) []ast.Expr {
	found, needToChainInterceptor := recipe.checkIfArgumentsAlreadyInstrumented(args, grpcFuncName, chainGRPCFuncName, instanaInstrumentationFunctionName)

	if !found {
		if needToChainInterceptor {
			return append(args, recipe.grpcExpr(chainGRPCFuncName, recipe.instrumentationExpr(instanaInstrumentationFunctionName)))
		} else {
			return append(args, recipe.grpcExpr(grpcFuncName, recipe.instrumentationExpr(instanaInstrumentationFunctionName)))
		}
	}

	return args
}

// generate grpcExpr expression
func (recipe GRPC) grpcExpr(funcName string, expr ast.Expr) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(recipe.TargetPkg),
			Sel: ast.NewIdent(funcName),
		},
		Args: []ast.Expr{
			expr,
		},
	}
}

// generate instana instrumentation expression
func (recipe GRPC) instrumentationExpr(funcName string) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(recipe.InstanaPkg),
			Sel: ast.NewIdent(funcName),
		},
		Args: []ast.Expr{
			ast.NewIdent(recipe.SensorVar),
		},
	}
}

// searchResult will be passed through recursive calls and collect information about founded artifacts
type searchResult struct {
	interceptorFuncName                string
	chainInterceptorFuncName           string
	instanaInstrumentationFunctionName string
	sensorVarName                      string

	sensorVarFound                  bool
	grpcFuncFound                   bool
	instanaInstrumentationFuncFound bool

	needToChain bool
}

func (s *searchResult) assertName(varName string) bool {
	if varName == s.sensorVarName {
		s.sensorVarFound = true

		return true
	}

	return false
}

func (s *searchResult) assertGRPCFuncName(funcName string) bool {
	if funcName == s.interceptorFuncName {
		s.needToChain = true
	}

	if funcName == s.interceptorFuncName || funcName == s.chainInterceptorFuncName {
		s.grpcFuncFound = true

		return true
	}

	return false
}

func (s *searchResult) assertInstanaInstrumentationFuncName(funcName string) bool {
	if funcName == s.instanaInstrumentationFunctionName {
		s.instanaInstrumentationFuncFound = true

		return true
	}

	return false
}

func (s *searchResult) isFound() bool {
	return s.sensorVarFound && s.grpcFuncFound && s.instanaInstrumentationFuncFound
}
