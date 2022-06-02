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
	"strings"
	"testing"
)

func TestGinRecipe(t *testing.T) {
	constructorsNames := []string{
		"Default",
		"New",
	}

	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"new within a function engine": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func foo() bool {
	var a = gin.^^CONSTRUCTOR_NAME^^()
	
	return true
}

func main() {
	foo()
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func foo() bool {
	var a = instagin.^^CONSTRUCTOR_NAME^^(__instanaSensor)
	return true
}
func main() {
	foo()
}
`,
		},
		"new engine": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	var a = gin.^^CONSTRUCTOR_NAME^^()
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	var a = instagin.^^CONSTRUCTOR_NAME^^(__instanaSensor)
}
`,
		},
		"new engine within if statement": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	if true {
		var a = gin.^^CONSTRUCTOR_NAME^^()
	}
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	if true {
		var a = instagin.^^CONSTRUCTOR_NAME^^(__instanaSensor)
	}
}
`,
		},
		"new engine within for statement": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	for {
		var a = gin.^^CONSTRUCTOR_NAME^^()
	}
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	for {
		var a = instagin.^^CONSTRUCTOR_NAME^^(__instanaSensor)
	}
}
`,
		},
		"new engine within goroutine": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	go func() {
		var a = gin.^^CONSTRUCTOR_NAME^^()
	}()
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	go func() {
		var a = instagin.^^CONSTRUCTOR_NAME^^(__instanaSensor)
	}()
}
`,
		},
		"new engine within block": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	{
		var a = gin.^^CONSTRUCTOR_NAME^^()
	}
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	{
		var a = instagin.^^CONSTRUCTOR_NAME^^(__instanaSensor)
	}
}
`,
		},
		"multiple engines": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	var a = gin.^^CONSTRUCTOR_NAME^^()
	var b = gin.^^CONSTRUCTOR_NAME^^()
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	var a = instagin.^^CONSTRUCTOR_NAME^^(__instanaSensor)
	var b = instagin.^^CONSTRUCTOR_NAME^^(__instanaSensor)
}
`,
		},
	}

	assertInstrumentation(t, examples, constructorsNames)
}

func assertInstrumentation(t *testing.T, examples map[string]struct {
	TargetPkg string
	Code      string
	Expected  string
}, constructorsNames []string) {
	for name, example := range examples {
		for _, constructorName := range constructorsNames {
			t.Run(name, func(t *testing.T) {
				fset := token.NewFileSet()
				node, err := parser.ParseFile(fset, "test", strings.ReplaceAll(example.Code, "^^CONSTRUCTOR_NAME^^", constructorName), parser.AllErrors)

				require.NoError(t, err)

				instrumented, changed := recipes.NewGin().
					Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

				assert.True(t, changed)

				buf := bytes.NewBuffer(nil)
				require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

				assert.Equal(t, strings.ReplaceAll(example.Expected, "^^CONSTRUCTOR_NAME^^", constructorName), buf.String())
			})
		}

	}
}

func TestAlreadyInstrumented(t *testing.T) {
	constructorsNames := []string{
		"Default",
		"New",
	}

	examples := map[string]struct {
		TargetPkg string
		Expected  string
	}{
		"new within a function engine": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func foo() bool {
	var a = instagin.^^CONSTRUCTOR_NAME^^()
	return true
}
func main() {
	foo()
}
`,
		},
		"new engine": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	var a = instagin.^^CONSTRUCTOR_NAME^^()
}
`,
		},
		"new engine within if statement": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	if true {
		var a = instagin.^^CONSTRUCTOR_NAME^^()
	}
}
`,
		},
		"new engine within for statement": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	for {
		var a = instagin.^^CONSTRUCTOR_NAME^^()
	}
}
`,
		},

		"new engine within goroutine": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	go func() {
		var a = instagin.^^CONSTRUCTOR_NAME^^()
	}()
}
`,
		},

		"new engine within block": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	{
		var a = instagin.^^CONSTRUCTOR_NAME^^()
	}
}
`,
		},

		"multiple engines": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	var a = instagin.^^CONSTRUCTOR_NAME^^()
	var b = instagin.^^CONSTRUCTOR_NAME^^()
}
`,
		},
	}

	for name, example := range examples {
		for _, constructorName := range constructorsNames {
			t.Run(name, func(t *testing.T) {
				fset := token.NewFileSet()
				node, err := parser.ParseFile(fset, "test", strings.ReplaceAll(example.Expected, "^^CONSTRUCTOR_NAME^^", constructorName), parser.AllErrors)

				require.NoError(t, err)

				instrumented, changed := recipes.NewGin().
					Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

				assert.False(t, changed)

				buf := bytes.NewBuffer(nil)
				require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

				assert.Equal(t, strings.ReplaceAll(example.Expected, "^^CONSTRUCTOR_NAME^^", constructorName), buf.String())
			})
		}
	}
}