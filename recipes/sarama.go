// (c) Copyright IBM Corp. 2022

package recipes

import (
	"errors"
	"fmt"
	"github.com/instana/go-instana/registry"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"strconv"
)

func init() {
	registry.Default.Register("github.com/Shopify/sarama", NewSarama())
}

// NewSarama returns Sarama recipe
func NewSarama() *Sarama {
	return &Sarama{InstanaPkg: "instasarama", defaultRecipe: defaultRecipe{}}
}

// Sarama instruments github.com/Shopify/sarama package with Instana
type Sarama struct {
	InstanaPkg    string
	defaultRecipe defaultRecipe
}

// ImportPath returns instrumentation import path
func (recipe *Sarama) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instasarama"
}

// Instrument applies recipe to the ast Node
func (recipe *Sarama) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (changed bool) {
	changed = recipe.defaultRecipe.instrument(fset, f, targetPkg, sensorVar, recipe.InstanaPkg, recipe.ImportPath(), map[string]insertOption{
		"NewAsyncProducer":           {sensorPosition: lastInsertPosition},
		"NewAsyncProducerFromClient": {sensorPosition: lastInsertPosition},
		"NewConsumer":                {sensorPosition: lastInsertPosition},
		"NewConsumerFromClient":      {sensorPosition: lastInsertPosition},
		"NewSyncProducer":            {sensorPosition: lastInsertPosition},
		"NewSyncProducerFromClient":  {sensorPosition: lastInsertPosition},
		"NewConsumerGroup":           {sensorPosition: lastInsertPosition},
		"NewConsumerGroupFromClient": {sensorPosition: lastInsertPosition},
	})

	funcDeclStack := &stack[ast.FuncDecl]{}
	if v, ok := f.(*ast.File); ok {
		contextImportName, err := recipe.getContextImportName(fset, v)

		if err != nil {
			log.Println(err)
			// use goto to simplify flow
			goto EXIT
		}

		astutil.Apply(f, func(cursor *astutil.Cursor) bool {
			if cursor.Node() == nil {
				return false
			}

			if fd, ok := (cursor.Node()).(*ast.FuncDecl); ok {
				funcDeclStack.Push(fd)
			}

			// check if this producer message creation
			recipe.tryToInstrumentProducerMessageCreation(cursor, contextImportName, funcDeclStack)

			return true
		}, func(cursor *astutil.Cursor) bool {
			if cursor.Node() == nil {
				return true
			}

			if _, ok := (cursor.Node()).(*ast.FuncDecl); ok {
				funcDeclStack.Pop()
			}

			return true
		})
	}

EXIT:
	return changed
}

func (recipe *Sarama) tryToInstrumentProducerMessageCreation(cursor *astutil.Cursor, contextImportName string, funcDeclStack *stack) {
	if unaryExp := recipe.isProducerMessageCreation(cursor); unaryExp != nil {
		if ctxName, ok := recipe.tryGetContextVariableNameInTheScope(contextImportName, funcDeclStack.Top()); ok {
			//todo: check if is instrumented

			cursor.Replace(
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "instasarama"},
						Sel: &ast.Ident{Name: "ProducerMessageWithSpanFromContext"},
					},
					Args: []ast.Expr{
						&ast.Ident{Name: ctxName},
						unaryExp,
					},
				})
		}
	}
}

// checking for this `msg := &sarama.ProducerMessage{...`
// todo: currently assumes that imported as sarama
func (recipe *Sarama) isProducerMessageCreation(cursor *astutil.Cursor) *ast.UnaryExpr {
	if unaryExp, ok := (cursor.Node()).(*ast.UnaryExpr); ok {
		if compositeLit, ok := (unaryExp.X).(*ast.CompositeLit); ok {
			if selExp, ok := (compositeLit.Type).(*ast.SelectorExpr); ok {
				xName := ""
				selName := ""
				if ident, ok := selExp.X.(*ast.Ident); ok {
					xName = ident.Name
				}

				if selExp.Sel != nil {
					selName = selExp.Sel.Name
				}

				if xName == "sarama" && selName == "ProducerMessage" && unaryExp.Op == token.AND {
					return unaryExp
				}
			}
		}
		return nil
	}

	return nil
}

func (recipe *Sarama) getContextImportName(fset *token.FileSet, f *ast.File) (string, error) {
	for _, importGroup := range astutil.Imports(fset, f) {
		for _, imp := range importGroup {
			if imp.Path == nil {
				return "", errors.New("import path is <nil>")
			}

			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return "", fmt.Errorf("can not unquote import path:%w", err)
			}

			if p == "context" {
				if imp.Name != nil {
					return imp.Name.Name, nil
				}

				return "context", nil
			}
		}
	}

	return "", errors.New("no context import found")
}

func (recipe *Sarama) tryGetContextVariableNameInTheScope(contextImportName string, fdcl *ast.FuncDecl) (string, bool) {
	// if there is no function declaration
	if fdcl == nil {
		return "", false
	}

	// check "special" import cases
	switch contextImportName {
	case "_":
		return "", false
	case ".":
		contextImportName = ""
	}

	// store all context variables from the declaration
	var ctxNames []string

	if fdcl.Type != nil && fdcl.Type.Params != nil {
		for _, field := range fdcl.Type.Params.List {
			if len(field.Names) != 1 {
				log.Println("declaration has more than 1 field.Names")
				continue
			}

			hasCorrectType := false
			hasCorrectImport := false
			if selExpr, ok := (field.Type).(*ast.SelectorExpr); ok {
				//case when imported as "context" or as named import

				//check type
				hasCorrectType = selExpr.Sel.Name == "Context"

				//check import name
				if ident, ok := (selExpr.X).(*ast.Ident); ok {
					hasCorrectImport = ident.Name == contextImportName
				}

				if hasCorrectImport && hasCorrectType {
					ctxNames = append(ctxNames, field.Names[0].Name)
				}

			} else if ident, ok := (field.Type).(*ast.Ident); ok {
				//case . "context"

				//check type
				hasCorrectType = ident.Name == "Context"

				if hasCorrectType {
					ctxNames = append(ctxNames, field.Names[0].Name)
				}
			}
		}
	}

	if len(ctxNames) != 1 {
		log.Println("expecting only one context in the declaration")
		return "", false
	}

	return ctxNames[0], true
}
