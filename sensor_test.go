// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package main_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	main "github.com/instana/go-instana"
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
				assert.Equal(t, example.Expected, main.LookupInstanaSensor(pkg))
			}
		})
	}
}

func TestWriteInstanaGoFile(t *testing.T) {
	const fixturePath = "./testdata/http/"

	defer resetDir(fixturePath)()

	filePath := filepath.Join(fixturePath, "instana.go")
	_, err := os.Stat(filePath)
	assert.True(t, os.IsNotExist(err))

	instanaGoFD, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	assert.NoError(t, err)

	notEmpty, err := main.WriteInstanaGoFile(instanaGoFD, "main", true, []string{})
	require.NoError(t, err)
	assert.True(t, notEmpty)
	defer instanaGoFD.Close()

	code, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.AllErrors)
	require.NoError(t, err)

	if assertImportsPackage(t, code.Imports, "instana", main.SensorPackage) {
		t.Run("instana.go exists", func(t *testing.T) {
			contentBefore, err := os.ReadFile(filePath)
			require.NoError(t, err)

			fd, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
			assert.NoError(t, err)
			defer fd.Close()

			notEmpty, err = main.WriteInstanaGoFile(fd, "main", true, []string{})
			assert.NoError(t, err)
			assert.True(t, notEmpty)

			contentAfter, err := os.ReadFile(filePath)
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
