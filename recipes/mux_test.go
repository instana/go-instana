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

func TestMuxRecipeWithDecl(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"new within a function router": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func foo() bool {
	var a = mux.NewRouter()
	
	return true
}

func main() {
	foo()
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func foo() bool {
	var a = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	return true
}
func main() {
	foo()
}
`,
		},
		"new router": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	var a = mux.NewRouter()
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	var a = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
}
`,
		},
		"new router within if statement": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	if true {
		var a = mux.NewRouter()
	}
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	if true {
		var a = mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},
		"new router within for statement": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	for {
		var a = mux.NewRouter()
	}
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	for {
		var a = mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},

		"new router within goroutine": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	go func() {
		var a = mux.NewRouter()
	}()
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	go func() {
		var a = mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}()
}
`,
		},

		"new router within block": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	{
		var a = mux.NewRouter()
	}
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	{
		var a = mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},

		"multiple routers 1": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	var a = mux.NewRouter()
	var b = mux.NewRouter()
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	var a = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	var b = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, b)
}
`,
		},
		"multiple routers 2": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	fmt.Println("empty line")
	var a = mux.NewRouter()
	fmt.Println("empty line")
	fmt.Println("empty line")
	var b = mux.NewRouter()
	fmt.Println("empty line")
	var c = mux.NewRouter()
	fmt.Println("empty line")
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	fmt.Println("empty line")
	var a = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	var b = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	var c = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
		"multiple routers 3": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	fmt.Println("empty line")
	var a = mux.NewRouter()
	fmt.Println("empty line")
	fmt.Println("empty line")
	if true {
		var a1 = mux.NewRouter()
		fmt.Println("empty line")
		fmt.Println("empty line")
		if true {
			var a2 = mux.NewRouter()
			fmt.Println("empty line")
			fmt.Println("empty line")
			var a3 = mux.NewRouter()
		}
		fmt.Println("empty line")
		var a4 = mux.NewRouter()
		fmt.Println("empty line")
	}
	var b = mux.NewRouter()
	fmt.Println("empty line")
	var c = mux.NewRouter()
	fmt.Println("empty line")
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	fmt.Println("empty line")
	var a = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	if true {
		var a1 = mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a1)
		fmt.Println("empty line")
		fmt.Println("empty line")
		if true {
			var a2 = mux.NewRouter()
			instamux.AddMiddleware(__instanaSensor, a2)
			fmt.Println("empty line")
			fmt.Println("empty line")
			var a3 = mux.NewRouter()
			instamux.AddMiddleware(__instanaSensor, a3)
		}
		fmt.Println("empty line")
		var a4 = mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a4)
		fmt.Println("empty line")
	}
	var b = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	var c = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
	}

	assertInstrumentationMux(t, examples)
}

func TestMuxRecipeWithAssignment(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"new within a function router": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func foo() bool {
	a := mux.NewRouter()
	
	return true
}

func main() {
	foo()
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func foo() bool {
	a := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	return true
}
func main() {
	foo()
}
`},
		"new router": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	a := mux.NewRouter()
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	a := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
}
`,
		},
		"new router within if statement": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	if true {
		a := mux.NewRouter()
	}
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	if true {
		a := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},
		"new router within for statement": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	for {
		a := mux.NewRouter()
	}
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	for {
		a := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},

		"new router within goroutine": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	go func() {
		a := mux.NewRouter()
	}()
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	go func() {
		a := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}()
}
`,
		},

		"new router within block": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	{
		a := mux.NewRouter()
	}
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	{
		a := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},

		"multiple routers 1": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	a := mux.NewRouter()
	b := mux.NewRouter()
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	a := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	b := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, b)
}
`,
		},
		"multiple routers 2": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	fmt.Println("empty line")
	a := mux.NewRouter()
	fmt.Println("empty line")
	fmt.Println("empty line")
	b := mux.NewRouter()
	fmt.Println("empty line")
	c := mux.NewRouter()
	fmt.Println("empty line")
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	fmt.Println("empty line")
	a := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	b := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	c := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
		"multiple routers 3": {
			TargetPkg: "mux",
			Code: `package main

import "github.com/gorilla/mux"

func main() {
	fmt.Println("empty line")
	a := mux.NewRouter()
	fmt.Println("empty line")
	fmt.Println("empty line")
	if true {
		a1 := mux.NewRouter()
		fmt.Println("empty line")
		fmt.Println("empty line")
		if true {
			a2 := mux.NewRouter()
			fmt.Println("empty line")
			fmt.Println("empty line")
			a3 := mux.NewRouter()
		}
		fmt.Println("empty line")
		a4 := mux.NewRouter()
		fmt.Println("empty line")
	}
	b := mux.NewRouter()
	fmt.Println("empty line")
	c := mux.NewRouter()
	fmt.Println("empty line")
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	fmt.Println("empty line")
	a := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	if true {
		a1 := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a1)
		fmt.Println("empty line")
		fmt.Println("empty line")
		if true {
			a2 := mux.NewRouter()
			instamux.AddMiddleware(__instanaSensor, a2)
			fmt.Println("empty line")
			fmt.Println("empty line")
			a3 := mux.NewRouter()
			instamux.AddMiddleware(__instanaSensor, a3)
		}
		fmt.Println("empty line")
		a4 := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a4)
		fmt.Println("empty line")
	}
	b := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	c := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
	}

	assertInstrumentationMux(t, examples)
}

func assertInstrumentationMux(t *testing.T, examples map[string]struct {
	TargetPkg string
	Code      string
	Expected  string
}) {
	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)

			require.NoError(t, err)

			instrumented, changed := recipes.NewMux().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func TestAlreadyInstrumentedMux_Decl(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Expected  string
	}{
		"new within a function router": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func foo() bool {
	var a = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	return true
}
func main() {
	foo()
}
`,
		},
		"new router": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	var a = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
}
`,
		},
		"new router within if statement": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	if true {
		var a = mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},
		"new router within for statement": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	for {
		var a = mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},

		"new router within goroutine": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	go func() {
		var a = mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}()
}
`,
		},

		"new router within block": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	{
		var a = mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},

		"multiple routers 1": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	var a = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	var b = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, b)
}
`,
		},
		"multiple routers 2": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	fmt.Println("empty line")
	var a = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	var b = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	var c = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
		"multiple routers 3": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	fmt.Println("empty line")
	var a = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	if true {
		var a1 = mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a1)
		fmt.Println("empty line")
		fmt.Println("empty line")
		if true {
			var a2 = mux.NewRouter()
			instamux.AddMiddleware(__instanaSensor, a2)
			fmt.Println("empty line")
			fmt.Println("empty line")
			var a3 = mux.NewRouter()
			instamux.AddMiddleware(__instanaSensor, a3)
		}
		fmt.Println("empty line")
		var a4 = mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a4)
		fmt.Println("empty line")
	}
	var b = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	var c = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Expected, parser.AllErrors)

			require.NoError(t, err)

			instrumented, changed := recipes.NewMux().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})

	}
}

func TestAlreadyInstrumentedMux_Assignment(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Expected  string
	}{
		"new within a function router": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func foo() bool {
	a := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	return true
}
func main() {
	foo()
}
`,
		},
		"new router": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	a := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
}
`,
		},
		"new router within if statement": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	if true {
		a := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},
		"new router within for statement": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	for {
		a := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},

		"new router within goroutine": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	go func() {
		a := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}()
}
`,
		},

		"new router within block": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	{
		a := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}
}
`,
		},

		"multiple routers 1": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	a := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	b := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, b)
}
`,
		},
		"multiple routers 2": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	fmt.Println("empty line")
	a := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	b := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	var c = mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
		"multiple routers 3": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	fmt.Println("empty line")
	a := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, a)
	fmt.Println("empty line")
	fmt.Println("empty line")
	if true {
		a1 := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a1)
		fmt.Println("empty line")
		fmt.Println("empty line")
		if true {
			a2 := mux.NewRouter()
			instamux.AddMiddleware(__instanaSensor, a2)
			fmt.Println("empty line")
			fmt.Println("empty line")
			a3 := mux.NewRouter()
			instamux.AddMiddleware(__instanaSensor, a3)
		}
		fmt.Println("empty line")
		a4 := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a4)
		fmt.Println("empty line")
	}
	b := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, b)
	fmt.Println("empty line")
	c := mux.NewRouter()
	instamux.AddMiddleware(__instanaSensor, c)
	fmt.Println("empty line")
}
`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Expected, parser.AllErrors)

			require.NoError(t, err)

			instrumented, changed := recipes.NewMux().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func TestMuxRecipeLimitation(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"new router": {
			TargetPkg: "mux",
			Code: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	var a = mux.NewRouter()
	func() {
		a := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}()
}
`,
			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	var a = mux.NewRouter()
	func() {
		a := mux.NewRouter()
		instamux.AddMiddleware(__instanaSensor, a)
	}()
}
`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)

			require.NoError(t, err)

			instrumented, changed := recipes.NewMux().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}
