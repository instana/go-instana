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
	"strings"
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
	m := map[string]insertOption{
		"NewAsyncProducer":           {sensorPosition: lastInsertPosition},
		"NewAsyncProducerFromClient": {sensorPosition: lastInsertPosition},
		"NewConsumer":                {sensorPosition: lastInsertPosition},
		"NewConsumerFromClient":      {sensorPosition: lastInsertPosition},
		"NewSyncProducer":            {sensorPosition: lastInsertPosition},
		"NewSyncProducerFromClient":  {sensorPosition: lastInsertPosition},
		"NewConsumerGroup":           {sensorPosition: lastInsertPosition},
		"NewConsumerGroupFromClient": {sensorPosition: lastInsertPosition},
	}
	changed = recipe.defaultRecipe.instrument(fset, f, targetPkg, sensorVar, recipe.InstanaPkg, recipe.ImportPath(), m)

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
			changed = recipe.tryToInstrumentProducerMessageCreation(cursor, contextImportName, funcDeclStack) || changed
			changed = recipe.tryToInstrumentSendingMessage(cursor, m, contextImportName, funcDeclStack) || changed

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

	if changed {
		if val, ok := f.(*ast.File); ok {
			log.Printf("AddNamedImport: %s %s", recipe.InstanaPkg, recipe.ImportPath())
			astutil.AddNamedImport(fset, val, recipe.InstanaPkg, recipe.ImportPath())
		}
	}

	return changed
}

func (recipe *Sarama) tryToInstrumentSendingMessage(cursor *astutil.Cursor, m map[string]insertOption, contextImportName string, funcDeclStack *stack[ast.FuncDecl]) bool {
	if callExpr, ok := (cursor.Node()).(*ast.CallExpr); ok {
		if recipe.isItCorrectSendMessageCall(callExpr, m) {
			//todo: check already instrumented
			if len(callExpr.Args) == 1 {
				if ctxName, ok := recipe.tryGetContextVariableNameInTheScope(contextImportName, funcDeclStack.Top()); ok {
					callExpr.Args[0] = &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "instasarama"},
							Sel: &ast.Ident{Name: "ProducerMessageWithSpanFromContext"},
						},
						Args: []ast.Expr{
							&ast.Ident{Name: ctxName},
							callExpr.Args[0],
						},
					}

					return true
				}
			}
		}
	}

	return false
}

func (recipe *Sarama) isItCorrectSendMessageCall(callExpr *ast.CallExpr, m map[string]insertOption) bool {
	if selExpr, ok := (callExpr.Fun).(*ast.SelectorExpr); ok {
		if selExpr.Sel.Name == "SendMessage" {
			if ident, ok := selExpr.X.(*ast.Ident); ok {
				t := recipe.getObjType(ident)
				instasaramaPrefix := "instasarama."
				saramaPrefix := "sarama."

				saramaTypesAndConstructors := map[string]struct{}{
					"SyncProducer": {},
					//"AsyncProducer":{}, doesn't have a SendMessage method
					"NewAsyncProducer":           {},
					"NewAsyncProducerFromClient": {},
					"NewConsumer":                {},
					"NewConsumerFromClient":      {},
					"NewSyncProducer":            {},
					"NewSyncProducerFromClient":  {},
					"NewConsumerGroup":           {},
					"NewConsumerGroupFromClient": {},
				}
				if strings.HasPrefix(t, instasaramaPrefix) {
					if _, ok := saramaTypesAndConstructors[strings.TrimPrefix(t, instasaramaPrefix)]; ok {
						return true
					}
				} else if strings.HasPrefix(t, saramaPrefix) {
					if _, ok := saramaTypesAndConstructors[strings.TrimPrefix(t, saramaPrefix)]; ok {
						return true
					}
				}
			}
		}
	}

	return false
}

// we are checking obj type, trying to use a `Obj.Decl` for this.
func (recipe *Sarama) getObjType(node any) string {
	switch t := node.(type) {
	case *ast.Ident:
		if t.Obj != nil && t.Obj.Decl != nil {
			return recipe.getObjType(t.Obj.Decl)
		} else {
			return t.String()
		}
	case *ast.Field:
		return recipe.getObjType(t.Type)
	case *ast.ValueSpec:
		return recipe.getObjType(t.Type)
	case *ast.AssignStmt:
		return recipe.getObjType(t.Rhs)
	case []ast.Expr:
		if len(t) == 1 {
			return recipe.getObjType(t[0])
		}
	case *ast.CallExpr:
		return recipe.getObjType(t.Fun)
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", recipe.getObjType(t.X), recipe.getObjType(t.Sel))
	}

	return ""
}

func (recipe *Sarama) tryToInstrumentProducerMessageCreation(cursor *astutil.Cursor, contextImportName string, funcDeclStack *stack[ast.FuncDecl]) bool {
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

			return true
		}
	}

	return false
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
