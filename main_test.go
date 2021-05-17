// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupInstanaSensor(t *testing.T) {
	examples := map[string]struct {
		Path     string
		Expected string
	}{
		"non-instrumented": {"./testdata/http/", ""},
		"instrumented":     {"./testdata/http-instrumented/", "sensor"},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			pkgs, err := parser.ParseDir(token.NewFileSet(), example.Path, func(fInfo os.FileInfo) bool {
				return !strings.HasSuffix(fInfo.Name(), "_test.go")
			}, 0)
			require.NoError(t, err)

			require.Len(t, pkgs, 1)

			for _, pkg := range pkgs {
				assert.Equal(t, example.Expected, LookupInstanaSensor(pkg))
			}
		})
	}
}

func TestAddInstanaSensor(t *testing.T) {
	const fixturePath = "./testdata/http/"

	defer resetDir(fixturePath)()

	expectedFilePath := filepath.Join(fixturePath, "instana.go")

	sensorName, err := AddInstanaSensor("main", fixturePath)
	require.NoError(t, err)

	assert.NotEmpty(t, sensorName)

	code, err := parser.ParseFile(token.NewFileSet(), expectedFilePath, nil, parser.AllErrors)
	require.NoError(t, err)

	if assertImportsPackage(t, code.Imports, "instana", goSensorPackage) {
		t.Run("instana.go exists", func(t *testing.T) {
			contentBefore, err := os.ReadFile(expectedFilePath)
			require.NoError(t, err)

			_, err = AddInstanaSensor("main", fixturePath)
			assert.Error(t, err)

			contentAfter, err := os.ReadFile(expectedFilePath)
			require.NoError(t, err)

			assert.Equal(t, string(contentBefore), string(contentAfter))
		})
	}
}

func assertImportsPackage(t *testing.T, imports []*ast.ImportSpec, name, path string) bool {
	t.Helper()

	m := make(map[string]string)
	for _, imp := range imports {
		pkgName, _ := strconv.Unquote(imp.Path.Value)
		m[pkgName] = imp.Name.Name
	}

	return assert.Contains(t, m, path) && assert.Equal(t, name, m[path], "%s import name", path)
}

func resetDir(dir string) func() {
	return func() {
		if err := os.Remove(filepath.Join(dir, "instana.go")); err != nil {
			if !os.IsNotExist(err) {
				panic(err)
			}
		}
	}
}
