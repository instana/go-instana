// (c) Copyright IBM Corp. 2022

package internal

import "go/ast"

// InsertInBlockOnce inserts statement `stmt` in the block statements list after `node` if is not already there.
func InsertInBlockOnce(block *ast.BlockStmt, node *ast.CallExpr, stmt ast.Stmt) bool {
	indexToInsertInstrumentation := getIndex(node, block)

	if indexToInsertInstrumentation < 0 || block == nil || stmt == nil || alreadyInstrumented(stmt, block) {
		return false
	}

	if indexToInsertInstrumentation == len(block.List) {
		block.List = append(block.List, stmt)

		return true
	}

	block.List = append(block.List[:indexToInsertInstrumentation+1], block.List[indexToInsertInstrumentation:]...)
	block.List[indexToInsertInstrumentation] = stmt

	return true
}

// getIndex returns position in the block list to insert an instrumentation call. If there is none, then it returns -1.
// This method has two branches for the case when engine is created via AssignStmt and DeclStmt.
// Ex: `var a = gin.New` and `a := gin.New`
func getIndex(node *ast.CallExpr, block *ast.BlockStmt) int {
	if block == nil {
		return -1
	}

	for k, v := range block.List {
		switch stmt := (v).(type) {
		case *ast.AssignStmt:
			if len(stmt.Rhs) == 1 && stmt.Rhs[0] == node {
				return k + 1
			}
		case *ast.DeclStmt:
			genDecl, ok := (stmt.Decl).(*ast.GenDecl)
			if !ok || len(genDecl.Specs) != 1 {
				return -1
			}

			valuesSpec, ok := (genDecl.Specs[0]).(*ast.ValueSpec)
			if !ok || len(valuesSpec.Values) != 1 {
				return -1
			}

			callExpr, ok := (valuesSpec.Values[0]).(*ast.CallExpr)
			if !ok {
				return -1
			}

			if callExpr == node {
				return k + 1
			}
		}
	}

	return -1
}

// alreadyInstrumented checks if the engine variable is already instrumented. Because it compares string values of the
// idents, it has some limitation. For example: please check `TestGinRecipeLimitation`
func alreadyInstrumented(stmt ast.Stmt, block ast.Node) bool {
	alreadyInstrumented := false

	ast.Inspect(block, func(node ast.Node) bool {
		if s, ok := (node).(ast.Stmt); ok {
			if isEqual(stmt, s) {
				alreadyInstrumented = true
			}
		}

		return true
	})

	return alreadyInstrumented
}

// isEqual is used to compare instrumentation statements like `instagin.AddMiddleware(__instanaSensor, a)`.
// It is not accurate, because it compares only string representation of the idents.
func isEqual(one ast.Stmt, second ast.Stmt) bool {
	instrumentationPkg1, method1, args1 := getFunctionCallDetails(one)
	instrumentationPkg2, method2, args2 := getFunctionCallDetails(second)

	if instrumentationPkg1 != instrumentationPkg2 {
		return false
	}

	if method1 != method2 {
		return false
	}

	if len(args1) != len(args2) {
		return false
	}

	for k := range args1 {
		if args1[k] != args2[k] {
			return false
		}
	}

	return true
}

// getFunctionCallDetails returns from the statements like `instagin.AddMiddleware(__instanaSensor, a)` its idents:
// Ex: `instagin.AddMiddleware(__instanaSensor, a)` -> "instagin", "AddMiddleware", ["__instanaSensor", "a"]
func getFunctionCallDetails(stmt ast.Stmt) (string, string, []string) {
	instrumentationPkg := ""
	method := ""
	var args []string

	exprStmt, ok := (stmt).(*ast.ExprStmt)
	if !ok {
		return "", "", nil
	}

	callExpr, ok := (exprStmt.X).(*ast.CallExpr)
	if !ok {
		return "", "", nil
	}

	selectorExpr, ok := (callExpr.Fun).(*ast.SelectorExpr)
	if !ok {
		return "", "", nil
	}

	if x, ok := (selectorExpr.X).(*ast.Ident); ok {
		instrumentationPkg = x.String()
	}

	method = selectorExpr.Sel.String()

	for _, arg := range callExpr.Args {
		if a, ok := (arg).(*ast.Ident); ok {
			args = append(args, a.String())
		}
	}

	return instrumentationPkg, method, args
}
