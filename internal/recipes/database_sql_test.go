// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package recipes_test

import (
	"bytes"
	"github.com/instana/go-instana/internal/recipes"
	"go/format"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseSQLRecipe(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"sql.Open": {
			TargetPkg: "sql",
			Code: `package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:example@tcp(127.0.0.1:3306)/hello")
	fmt.Println(db, err)
}
`,
			Expected: `package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	instana "github.com/instana/go-sensor"
)

func main() {
	db, err := instana.SQLInstrumentAndOpen(__instanaSensor, "mysql", "root:example@tcp(127.0.0.1:3306)/hello")
	fmt.Println(db, err)
}
`,
		}}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)

			require.NoError(t, err)

			changed := recipes.NewDatabaseSQL().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

			dumpExpectedCode(t, "databasesql", name, buf)

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func TestDatabaseSQLRecipe_AlreadyInstrumented(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Expected  string
	}{
		"sql.Open already instrumented": {
			TargetPkg: "sql",
			Expected: `package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	instana "github.com/instana/go-sensor"
)

func main() {
	db, err := instana.SQLInstrumentAndOpen(__instanaSensor, "mysql", "root:example@tcp(127.0.0.1:3306)/hello")
	fmt.Println(db, err)
}
`,
		}}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Expected, parser.AllErrors)

			require.NoError(t, err)

			changed := recipes.NewDatabaseSQL().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

			dumpExpectedCode(t, "databasesql", name, buf)

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}
