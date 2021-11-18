// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package registry_test

import (
	"testing"

	"github.com/instana/go-instana/recipes"
	"github.com/instana/go-instana/registry"
	"github.com/stretchr/testify/assert"
)

func TestRegistry(t *testing.T) {
	targetPkg := "http"

	expectedRecipe := recipes.NewNetHTTP()
	r := registry.NewRegistry()
	r.Register(targetPkg, expectedRecipe)

	assert.Equal(t, expectedRecipe.ImportPath(), r.InstrumentationImportPath(targetPkg))
	recipe := r.InstrumentationRecipe(targetPkg)
	assert.Equal(t, expectedRecipe, recipe)
}
