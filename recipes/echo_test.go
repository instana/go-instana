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

func TestEchoClientRecipe(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
		Changed   bool
	}{
		"new engine instrumentation": {
			TargetPkg: "echo",
			Code:      `echo.New()`,
			Expected:  `instaecho.New(__instanaSensor)`,
			Changed:   true,
		},
		"already instrumented": {
			TargetPkg: "echo",
			Code:      `instaecho.New(__instanaSensor)`,
			Expected:  `instaecho.New(__instanaSensor)`,
			Changed:   false,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			node, err := parser.ParseExpr(example.Code)
			require.NoError(t, err)

			instrumented, changed := recipes.NewEcho().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.Equal(t, example.Changed, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}
