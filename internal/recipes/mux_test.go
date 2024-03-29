// (c) Copyright IBM Corp. 2022

package recipes_test

import (
	"bytes"
	"github.com/instana/go-instana/internal/recipes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/format"
	"go/parser"
	"go/token"
	"testing"
)

func TestMuxRecipeWith(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"new router": {
			TargetPkg: "mux",
			Code: `package main

import "fmt"
import "github.com/gorilla/mux"

func main() {
	var a = mux.NewRouter()
	fmt.Println(a)
}
`,
			Expected: `package main

import (
	"fmt"
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	var a = instamux.NewRouter(__instanaSensor)
	fmt.Println(a)
}
`,
		},
		"multiple routers": {
			TargetPkg: "mux",
			Code: `package main

import "fmt"
import "github.com/gorilla/mux"

func main() {
	var a = mux.NewRouter()
	var b = mux.NewRouter()
	fmt.Println(a)
	fmt.Println(b)
}
`,
			Expected: `package main

import (
	"fmt"
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	var a = instamux.NewRouter(__instanaSensor)
	var b = instamux.NewRouter(__instanaSensor)
	fmt.Println(a)
	fmt.Println(b)
}
`,
		},
	}

	assertMuxInstrumentation(t, examples)
}

func assertMuxInstrumentation(t *testing.T, examples map[string]struct {
	TargetPkg string
	Code      string
	Expected  string
}) {
	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)

			require.NoError(t, err)

			changed := recipes.NewMux().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

			dumpExpectedCode(t, "mux", name, buf)

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func TestAlreadyInstrumentedMux(t *testing.T) {
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
	var a = instamux.NewRouter(__instanaSensor)
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
	var a = instamux.NewRouter(__instanaSensor)
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
		var a = instamux.NewRouter(__instanaSensor)
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
		var a = instamux.NewRouter(__instanaSensor)
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
		var a = instamux.NewRouter(__instanaSensor)
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
		var a = instamux.NewRouter(__instanaSensor)
	}
}
`,
		},

		"multiple routers": {
			TargetPkg: "mux",

			Expected: `package main

import (
	"github.com/gorilla/mux"
	instamux "github.com/instana/go-sensor/instrumentation/instamux"
)

func main() {
	var a = instamux.NewRouter(__instanaSensor)
	var b = instamux.NewRouter(__instanaSensor)
}
`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Expected, parser.AllErrors)

			require.NoError(t, err)

			changed := recipes.NewMux().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

			assert.Equal(t, example.Expected, buf.String())
		})

	}
}
