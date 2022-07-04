package recipes_test

import (
	"bytes"
	"github.com/instana/go-instana/internal/recipes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func dumpExpectedCode(t *testing.T, pkgName, name string, buf *bytes.Buffer) {
	if _, ok := os.LookupEnv("GO_INSTANA_TEST_DUMP"); ok {
		pkgName = strings.ReplaceAll(pkgName, " ", "_")
		name = strings.ReplaceAll(name, " ", "_")

		p := filepath.Join("../../testdata/tmp", pkgName, name)
		assert.NoError(t, os.MkdirAll(p, 0700))

		f, err := os.OpenFile(filepath.Join(p, name+".go"), os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
		assert.NoError(t, err)

		_, err = f.Write(buf.Bytes())
		assert.NoError(t, err)

		assert.NoError(t, f.Close())

		f2, err := os.OpenFile(filepath.Join(p, "gen.go"), os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
		assert.NoError(t, err)

		_, err = f2.Write([]byte(`//go:generate go mod init example.com/m
//` + `go:generate go mod tidy
//` + `go:generate go-instana add
//` + `go:generate goimports -w .
//` + `go:generate go build .

package main`))
		assert.NoError(t, err)

		assert.NoError(t, f2.Close())
	}
}

func TestGetPackageImportName(t *testing.T) {
	code := `package main

import (
	"log"
	"github.com/labstack/echo/v4"
	db "database/sql"
	. "point"
	_ "dash"
)

func main() {
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test", code, parser.AllErrors)

	require.NoError(t, err)

	var p string

	p, err = recipes.GetPackageImportName(fset, node, "log")
	assert.NoError(t, err)
	assert.Equal(t, "log", p)

	p, err = recipes.GetPackageImportName(fset, node, "github.com/labstack/echo/v4")
	assert.NoError(t, err)
	assert.Equal(t, "echo", p)

	p, err = recipes.GetPackageImportName(fset, node, "database/sql")
	assert.NoError(t, err)
	assert.Equal(t, "db", p)

	_, err = recipes.GetPackageImportName(fset, node, "point")
	assert.Error(t, err)

	_, err = recipes.GetPackageImportName(fset, node, "dash")
	assert.Error(t, err)

	_, err = recipes.GetPackageImportName(fset, node, "???")
	assert.Error(t, err)
}
