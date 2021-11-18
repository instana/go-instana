// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package recipes_test

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"testing"

	"github.com/instana/go-instana/recipes"
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
			Code:      `sql.Open("sqlite", "sqlite:///var/data/db.sqlite")`,
			Expected:  `instana.SQLOpen("sqlite", "sqlite:///var/data/db.sqlite")`,
		},
		"custom import": {
			TargetPkg: "db",
			Code:      `db.Open("sqlite", "sqlite:///var/data/db.sqlite")`,
			Expected:  `instana.SQLOpen("sqlite", "sqlite:///var/data/db.sqlite")`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			node, err := parser.ParseExpr(example.Code)
			require.NoError(t, err)

			instrumented, changed := recipes.NewDatabaseSQL().
				Instrument(nil, node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}
