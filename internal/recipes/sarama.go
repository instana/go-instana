// (c) Copyright IBM Corp. 2022

package recipes

import (
	"fmt"
	"github.com/instana/go-instana/internal/registry"
	"github.com/rs/zerolog/log"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
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
	changed = recipe.instrumentMessagesAndSending(fset, f) || changed

	if changed {
		addNamedImport(fset, f, recipe.InstanaPkg, recipe.ImportPath())
	}

	return changed
}

// instrumentMessagesAndSending iterates over ast tree and track current function declaration. If the current function
// has a "context.Context" type, it tries to instrument "sarama.ProducerMessage" type creation and/or "SendMessage" call
// if that is done by "sarama.SyncProducer". Important: this auto instrumentation assumes that sarama library
// ("github.com/Shopify/sarama") is imported as "sarama" and "context" is not imported via "_" or ".".
func (recipe *Sarama) instrumentMessagesAndSending(fset *token.FileSet, f ast.Node) (changed bool) {
	// stack to store current function declaration
	funcDeclStack := &stack[ast.FuncDecl]{}

	if v, ok := f.(*ast.File); ok {
		// try to get context variable name
		contextImportName, err := GetPackageImportName(fset, v, "context")

		// if no proper context import found
		if err != nil {
			log.Debug().Msgf("sarama instrumentation : %s", err.Error())
			return false
		}

		// traverse a tree
		astutil.Apply(f, func(cursor *astutil.Cursor) bool {

			// stop checking children
			if cursor.Node() == nil {
				return false
			}

			// if current node is a function declaration add it to the stack
			if fd, ok := (cursor.Node()).(*ast.FuncDecl); ok {
				funcDeclStack.Push(fd)

				return true
			}

			// try to instrument "sarama.ProducerMessage" type creation
			changed = recipe.tryToInstrumentProducerMessageCreation(cursor, contextImportName, funcDeclStack) || changed

			// try to instrument "SendMessage" call
			changed = recipe.tryToInstrumentSendingMessage(cursor, contextImportName, funcDeclStack) || changed

			return true
		}, func(cursor *astutil.Cursor) bool {
			if cursor.Node() == nil {
				return true
			}

			// remove function declaration from the stack
			if _, ok := (cursor.Node()).(*ast.FuncDecl); ok {
				funcDeclStack.Pop()
			}

			return true
		})
	}

	return changed
}

// tryToInstrumentSendingMessage instruments first and only argument of the "sarama.SyncProducer" "SendMessage" call
func (recipe *Sarama) tryToInstrumentSendingMessage(cursor *astutil.Cursor, contextImportName string, funcDeclStack *stack[ast.FuncDecl]) bool {
	if callExpr, ok := (cursor.Node()).(*ast.CallExpr); ok {
		if recipe.isItCorrectSendMessageCall(callExpr) {
			if len(callExpr.Args) == 1 {
				// check if already instrumented
				if ce, ok := callExpr.Args[0].(*ast.CallExpr); ok {
					if recipe.getObjType(ce) == "instasarama.ProducerMessageWithSpanFromContext" {
						return false
					}
				}

				// check the name of the context variable name in the function declaration
				if ctxName, ok := recipe.tryGetContextVariableNameInTheFunctionDeclaration(contextImportName, funcDeclStack.Top()); ok {
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

// isItCorrectSendMessageCall checks if current call is "SendMessage" and belongs to the kafka publishing.
// It assumes that instrumentation and sarama library were imported with their default names.
// It does not support async publisher.
func (recipe *Sarama) isItCorrectSendMessageCall(callExpr *ast.CallExpr) bool {
	if selExpr, ok := (callExpr.Fun).(*ast.SelectorExpr); ok {
		if selExpr.Sel.Name == "SendMessage" {
			if ident, ok := selExpr.X.(*ast.Ident); ok {
				instasaramaPrefix := "instasarama."
				saramaPrefix := "sarama."

				//"sarama.AsyncProducer" doesn't have a SendMessage method
				saramaProducerTypesAndConstructors := map[string]struct{}{
					"SyncProducer":               {},
					"NewAsyncProducer":           {},
					"NewAsyncProducerFromClient": {},
					"NewSyncProducer":            {},
					"NewSyncProducerFromClient":  {},
				}

				// expected to have types like "instasarama.SyncProducer" or "sarama.SyncProducer" and so on.
				t := recipe.getObjType(ident)
				if strings.HasPrefix(t, instasaramaPrefix) {
					if _, ok := saramaProducerTypesAndConstructors[strings.TrimPrefix(t, instasaramaPrefix)]; ok {
						return true
					}
				} else if strings.HasPrefix(t, saramaPrefix) {
					if _, ok := saramaProducerTypesAndConstructors[strings.TrimPrefix(t, saramaPrefix)]; ok {
						return true
					}
				}
			}
		}
	}

	return false
}

// getObjType originally was created to extract select expression from the ident object declaration.
// Use this method only if you understand what it does :).
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

// tryToInstrumentProducerMessageCreation wraps "&sarama.ProducerMessage{...}"
// with "instasarama.ProducerMessageWithSpanFromContext"
func (recipe *Sarama) tryToInstrumentProducerMessageCreation(cursor *astutil.Cursor, contextImportName string, funcDeclStack *stack[ast.FuncDecl]) bool {
	// check if it is unary expression that creates "&sarama.ProducerMessage"
	if unaryExp := recipe.isProducerMessageCreation(cursor.Node()); unaryExp != nil {

		// search for the "context.Context" variable name in the current function declaration
		if ctxName, ok := recipe.tryGetContextVariableNameInTheFunctionDeclaration(contextImportName, funcDeclStack.Top()); ok {
			// check if is already instrumented
			if recipe.getObjType(cursor.Parent()) == "instasarama.ProducerMessageWithSpanFromContext" {
				return false
			}

			// wrap message creation
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

// isProducerMessageCreation checking if current node is unary expression like `msg := &sarama.ProducerMessage{...`
func (recipe *Sarama) isProducerMessageCreation(node ast.Node) *ast.UnaryExpr {
	if unaryExp, ok := node.(*ast.UnaryExpr); ok {
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

// tryGetContextVariableNameInTheFunctionDeclaration check if current FuncDecl has "context.Context" type among
// the parameters and returns its name.
func (recipe *Sarama) tryGetContextVariableNameInTheFunctionDeclaration(contextImportName string, fdcl *ast.FuncDecl) (string, bool) {
	// if there is no function declaration, returns
	if fdcl == nil {
		return "", false
	}

	// store all context variables from the declaration
	var ctxNames []string

	if fdcl.Type != nil && fdcl.Type.Params != nil {
		for _, field := range fdcl.Type.Params.List {
			if len(field.Names) != 1 {
				log.Warn().Msg("declaration has more than 1 field.Names. Skipping...")
				continue
			}

			if recipe.getObjType(field) == contextImportName+".Context" {
				ctxNames = append(ctxNames, field.Names[0].Name)
			}
		}
	}

	if len(ctxNames) != 1 {
		log.Warn().Msg("expecting only one context in the function declaration. Skipping...")
		return "", false
	}

	return ctxNames[0], true
}
