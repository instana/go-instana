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

func TestGinRecipeWithDecl(t *testing.T) {
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
	var a = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
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
	var a = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
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
		var a = gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
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
		var a = gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
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
		var a = gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
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
		var a = gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},

		"multiple engines 1": {
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
	var a = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	var b = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, b)
}
`,
		},
		"multiple engines 2": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	fmt.Println("empty line")
	var a = gin.^^CONSTRUCTOR_NAME^^()
	fmt.Println("empty line")
	fmt.Println("empty line")
	var b = gin.^^CONSTRUCTOR_NAME^^()
	fmt.Println("empty line")
	var c = gin.^^CONSTRUCTOR_NAME^^()
	fmt.Println("empty line")
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	fmt.Println("empty line")
	var a = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	var b = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	var c = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
		"multiple engines 3": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	fmt.Println("empty line")
	var a = gin.^^CONSTRUCTOR_NAME^^()
	fmt.Println("empty line")
	fmt.Println("empty line")
	if true {
		var a1 = gin.^^CONSTRUCTOR_NAME^^()
		fmt.Println("empty line")
		fmt.Println("empty line")
		if true {
			var a2 = gin.^^CONSTRUCTOR_NAME^^()
			fmt.Println("empty line")
			fmt.Println("empty line")
			var a3 = gin.^^CONSTRUCTOR_NAME^^()
		}
		fmt.Println("empty line")
		var a4 = gin.^^CONSTRUCTOR_NAME^^()
		fmt.Println("empty line")
	}
	var b = gin.^^CONSTRUCTOR_NAME^^()
	fmt.Println("empty line")
	var c = gin.^^CONSTRUCTOR_NAME^^()
	fmt.Println("empty line")
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	fmt.Println("empty line")
	var a = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	if true {
		var a1 = gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a1)
		fmt.Println("empty line")
		fmt.Println("empty line")
		if true {
			var a2 = gin.^^CONSTRUCTOR_NAME^^()
			instagin.AddMiddleware(__instanaSensor, a2)
			fmt.Println("empty line")
			fmt.Println("empty line")
			var a3 = gin.^^CONSTRUCTOR_NAME^^()
			instagin.AddMiddleware(__instanaSensor, a3)
		}
		fmt.Println("empty line")
		var a4 = gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a4)
		fmt.Println("empty line")
	}
	var b = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	var c = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
	}

	assertInstrumentation(t, examples, constructorsNames)
}

func TestGinRecipeWithAssignment(t *testing.T) {
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
	a := gin.^^CONSTRUCTOR_NAME^^()
	
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
	a := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	return true
}
func main() {
	foo()
}
`},
		"new engine": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	a := gin.^^CONSTRUCTOR_NAME^^()
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	a := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
}
`,
		},
		"new engine within if statement": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	if true {
		a := gin.^^CONSTRUCTOR_NAME^^()
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
		a := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
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
		a := gin.^^CONSTRUCTOR_NAME^^()
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
		a := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
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
		a := gin.^^CONSTRUCTOR_NAME^^()
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
		a := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
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
		a := gin.^^CONSTRUCTOR_NAME^^()
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
		a := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},

		"multiple engines 1": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	a := gin.^^CONSTRUCTOR_NAME^^()
	b := gin.^^CONSTRUCTOR_NAME^^()
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	a := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	b := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, b)
}
`,
		},
		"multiple engines 2": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	fmt.Println("empty line")
	a := gin.^^CONSTRUCTOR_NAME^^()
	fmt.Println("empty line")
	fmt.Println("empty line")
	b := gin.^^CONSTRUCTOR_NAME^^()
	fmt.Println("empty line")
	c := gin.^^CONSTRUCTOR_NAME^^()
	fmt.Println("empty line")
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	fmt.Println("empty line")
	a := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	b := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	c := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
		"multiple engines 3": {
			TargetPkg: "gin",
			Code: `package main

import "github.com/gin-gonic/gin"

func main() {
	fmt.Println("empty line")
	a := gin.^^CONSTRUCTOR_NAME^^()
	fmt.Println("empty line")
	fmt.Println("empty line")
	if true {
		a1 := gin.^^CONSTRUCTOR_NAME^^()
		fmt.Println("empty line")
		fmt.Println("empty line")
		if true {
			a2 := gin.^^CONSTRUCTOR_NAME^^()
			fmt.Println("empty line")
			fmt.Println("empty line")
			a3 := gin.^^CONSTRUCTOR_NAME^^()
		}
		fmt.Println("empty line")
		a4 := gin.^^CONSTRUCTOR_NAME^^()
		fmt.Println("empty line")
	}
	b := gin.^^CONSTRUCTOR_NAME^^()
	fmt.Println("empty line")
	c := gin.^^CONSTRUCTOR_NAME^^()
	fmt.Println("empty line")
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	fmt.Println("empty line")
	a := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	if true {
		a1 := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a1)
		fmt.Println("empty line")
		fmt.Println("empty line")
		if true {
			a2 := gin.^^CONSTRUCTOR_NAME^^()
			instagin.AddMiddleware(__instanaSensor, a2)
			fmt.Println("empty line")
			fmt.Println("empty line")
			a3 := gin.^^CONSTRUCTOR_NAME^^()
			instagin.AddMiddleware(__instanaSensor, a3)
		}
		fmt.Println("empty line")
		a4 := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a4)
		fmt.Println("empty line")
	}
	b := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	c := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
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

func TestAlreadyInstrumented_Decl(t *testing.T) {
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
	var a = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
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
	var a = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
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
		var a = gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
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
		var a = gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
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
		var a = gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
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
		var a = gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},

		"multiple engines 1": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	var a = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	var b = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, b)
}
`,
		},
		"multiple engines 2": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	fmt.Println("empty line")
	var a = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	var b = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	var c = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
		"multiple engines 3": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	fmt.Println("empty line")
	var a = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	if true {
		var a1 = gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a1)
		fmt.Println("empty line")
		fmt.Println("empty line")
		if true {
			var a2 = gin.^^CONSTRUCTOR_NAME^^()
			instagin.AddMiddleware(__instanaSensor, a2)
			fmt.Println("empty line")
			fmt.Println("empty line")
			var a3 = gin.^^CONSTRUCTOR_NAME^^()
			instagin.AddMiddleware(__instanaSensor, a3)
		}
		fmt.Println("empty line")
		var a4 = gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a4)
		fmt.Println("empty line")
	}
	var b = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	var c = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
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

func TestAlreadyInstrumented_Assignment(t *testing.T) {
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
	a := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
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
	a := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
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
		a := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
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
		a := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
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
		a := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
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
		a := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},

		"multiple engines 1": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	a := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	b := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, b)
}
`,
		},
		"multiple engines 2": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	fmt.Println("empty line")
	a := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	b := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	var c = gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
		"multiple engines 3": {
			TargetPkg: "gin",

			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	fmt.Println("empty line")
	a := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	if true {
		a1 := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a1)
		fmt.Println("empty line")
		fmt.Println("empty line")
		if true {
			a2 := gin.^^CONSTRUCTOR_NAME^^()
			instagin.AddMiddleware(__instanaSensor, a2)
			fmt.Println("empty line")
			fmt.Println("empty line")
			a3 := gin.^^CONSTRUCTOR_NAME^^()
			instagin.AddMiddleware(__instanaSensor, a3)
		}
		fmt.Println("empty line")
		a4 := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a4)
		fmt.Println("empty line")
	}
	b := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	c := gin.^^CONSTRUCTOR_NAME^^()
	instagin.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
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

func TestGinRecipeLimitation(t *testing.T) {
	constructorsNames := []string{
		"Default",
		"New",
	}

	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"new engine": {
			TargetPkg: "gin",
			Code: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	var a = gin.^^CONSTRUCTOR_NAME^^()
	func() {
		a := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
	}()
}
`,
			Expected: `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
)

func main() {
	var a = gin.^^CONSTRUCTOR_NAME^^()
	func() {
		a := gin.^^CONSTRUCTOR_NAME^^()
		instagin.AddMiddleware(__instanaSensor, a)
	}()
}
`,
		},
	}

	for name, example := range examples {
		for _, constructorName := range constructorsNames {
			t.Run(name, func(t *testing.T) {
				fset := token.NewFileSet()
				node, err := parser.ParseFile(fset, "test", strings.ReplaceAll(example.Code, "^^CONSTRUCTOR_NAME^^", constructorName), parser.AllErrors)

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
