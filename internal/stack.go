// (c) Copyright IBM Corp. 2022

package internal

import "go/ast"

type BlockStmtStack []*ast.BlockStmt

func (s *BlockStmtStack) Push(val *ast.BlockStmt) {
	*s = append(*s, val)
}

func (s *BlockStmtStack) Top() *ast.BlockStmt {
	index := len(*s) - 1
	if index >= 0 {
		return (*s)[index]
	}

	return nil
}

func (s *BlockStmtStack) Pop() *ast.BlockStmt {
	index := len(*s) - 1
	if index >= 0 {
		res := (*s)[index]
		*s = (*s)[:index]

		return res
	}

	return nil
}
