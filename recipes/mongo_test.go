// (c) Copyright IBM Corp. 2022

package recipes_test

import (
	"bytes"
	"github.com/instana/go-instana/recipes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/format"
	"go/parser"
	"go/token"
	"testing"
)

func TestMongoClientRecipe(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
		Changed   bool
	}{
		"Connect instrumentation": {
			TargetPkg: "mongo",
			Code:      `mongo.Connect(ctx, options.Client().ApplyURI("mongodb://foo:bar@localhost:27017"))`,
			Expected:  `instamongo.Connect(ctx, __instanaSensor, options.Client().ApplyURI("mongodb://foo:bar@localhost:27017"))`,
			Changed:   true,
		},
		"already instrumented Connect": {
			TargetPkg: "mongo",
			Code:      `instamongo.Connect(ctx, sensor, options.Client().ApplyURI("mongodb://foo:bar@localhost:27017"))`,
			Expected:  `instamongo.Connect(ctx, sensor, options.Client().ApplyURI("mongodb://foo:bar@localhost:27017"))`,
			Changed:   false,
		},
		"NewClient instrumentation": {
			TargetPkg: "mongo",
			Code:      `mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))`,
			Expected:  `instamongo.NewClient(__instanaSensor, options.Client().ApplyURI("mongodb://localhost:27017"))`,
			Changed:   true,
		},
		"already instrumented NewClient": {
			TargetPkg: "mongo",
			Code:      `instamongo.NewClient(__instanaSensor, options.Client().ApplyURI("mongodb://localhost:27017"))`,
			Expected:  `instamongo.NewClient(__instanaSensor, options.Client().ApplyURI("mongodb://localhost:27017"))`,
			Changed:   false,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			node, err := parser.ParseExpr(example.Code)
			require.NoError(t, err)

			changed := recipes.NewMongo().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.Equal(t, example.Changed, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}
