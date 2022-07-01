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

func TestEchoClientRecipe(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
		Changed   bool
	}{
		"new engine instrumentation": {
			TargetPkg: "echo",
			Code: `package main

import (
	"github.com/labstack/echo/v4"
	"log"
)

func main() {
}
func RunEcho(addr string) {
	engine := echo.New()

	engine.GET("/echoEndpoint", func(c echo.Context) error {
		return nil
	})

	if err := engine.Start(addr); err != nil {
		log.Fatalln(err)
	}
}
`,
			Expected: `package main

import (
	instaecho "github.com/instana/go-sensor/instrumentation/instaecho"
	"github.com/labstack/echo/v4"
	"log"
)

func main() {
}
func RunEcho(addr string) {
	engine := instaecho.New(__instanaSensor)
	engine.GET("/echoEndpoint", func(c echo.Context) error {
		return nil
	})
	if err := engine.Start(addr); err != nil {
		log.Fatalln(err)
	}
}
`,
			Changed: true,
		},
		"already instrumented": {
			TargetPkg: "echo",
			Code: `package main

import (
	instaecho "github.com/instana/go-sensor/instrumentation/instaecho"
	echo "github.com/labstack/echo/v4"
	"log"
)

func main() {
}
func RunEcho(addr string) {
	engine := instaecho.New(__instanaSensor)
	engine.GET("/echoEndpoint", func(c echo.Context) error {
		return nil
	})
	if err := engine.Start(addr); err != nil {
		log.Fatalln(err)
	}
}
`,
			Expected: `package main

import (
	instaecho "github.com/instana/go-sensor/instrumentation/instaecho"
	echo "github.com/labstack/echo/v4"
	"log"
)

func main() {
}
func RunEcho(addr string) {
	engine := instaecho.New(__instanaSensor)
	engine.GET("/echoEndpoint", func(c echo.Context) error {
		return nil
	})
	if err := engine.Start(addr); err != nil {
		log.Fatalln(err)
	}
}
`,
			Changed: false,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)

			require.NoError(t, err)

			changed := recipes.NewEcho().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.Equal(t, example.Changed, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

			dumpExpectedCode(t, "echo", name, buf)

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}
