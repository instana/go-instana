// (c) Copyright IBM Corp. 2022

package recipes

import (
	"github.com/instana/go-instana/internal/registry"
	"go/ast"
	"go/token"
)

func init() {
	registry.Default.Register("github.com/aws/aws-sdk-go/aws/session", NewAWSSDK())
}

// NewAWSSDK returns AWSSDK recipe
func NewAWSSDK() *AWSSDK {
	return &AWSSDK{InstanaPkg: "instaawssdk", defaultRecipe: defaultRecipe{}}
}

// AWSSDK instruments github.com/aws/aws-sdk-go/aws/session package with Instana
type AWSSDK struct {
	InstanaPkg    string
	defaultRecipe defaultRecipe
}

// ImportPath returns instrumentation import path
func (recipe *AWSSDK) ImportPath() string {
	return "github.com/instana/go-sensor/instrumentation/instaawssdk"
}

// Instrument applies recipe to the ast Node
func (recipe *AWSSDK) Instrument(fset *token.FileSet, f ast.Node, targetPkg, sensorVar string) (changed bool) {
	return recipe.defaultRecipe.instrument(fset, f, targetPkg, sensorVar, recipe.InstanaPkg, recipe.ImportPath(), map[string]insertOption{
		"New":                   {},
		"NewSession":            {},
		"NewSessionWithOptions": {},
	})
}
