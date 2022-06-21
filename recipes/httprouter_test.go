// (c) Copyright IBM Corp. 2022

package recipes

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpRouterRecipe(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
		Changed   bool
	}{
		"with type inference": {
			TargetPkg: "httprouter",
			Changed:   true,
			Code: `package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()
}
`,
			Expected: `package main

import (
	"fmt"
	instahttprouter "github.com/instana/go-sensor/instrumentation/instahttprouter"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func main() {
	router := instahttprouter.Wrap(httprouter.New(), __instanaSensor)
}
`,
		},
		"with var declaration": {
			TargetPkg: "httprouter",
			Changed:   true,
			Code: `package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func useRouterInstance(r httprouter.Router) {
	fmt.Println(r)
}

func main() {
	var router *httprouter.Router
	router = httprouter.New()
}
`,
			Expected: `package main

import (
	"fmt"
	instahttprouter "github.com/instana/go-sensor/instrumentation/instahttprouter"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func useRouterInstance(r instahttprouter.WrappedRouter) {
	fmt.Println(r)
}
func main() {
	var router *instahttprouter.WrappedRouter
	router = instahttprouter.Wrap(httprouter.New(), __instanaSensor)
}
`,
		},
		"code already instrumented": {
			TargetPkg: "httprouter",
			Changed:   false,
			Code: `package main

import (
	"fmt"
	instahttprouter "github.com/instana/go-sensor/instrumentation/instahttprouter"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func useRouterInstance(r instahttprouter.WrappedRouter) {
	fmt.Println(r)
}
func main() {
	var router *instahttprouter.WrappedRouter
	router = instahttprouter.Wrap(httprouter.New(), __instanaSensor)
}
`,
			Expected: `package main

import (
	"fmt"
	instahttprouter "github.com/instana/go-sensor/instrumentation/instahttprouter"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func useRouterInstance(r instahttprouter.WrappedRouter) {
	fmt.Println(r)
}
func main() {
	var router *instahttprouter.WrappedRouter
	router = instahttprouter.Wrap(httprouter.New(), __instanaSensor)
}
`,
		},
		"instrument httprouter.New only": {
			TargetPkg: "httprouter",
			Changed:   true,
			Code: `package main

import (
	"log"
	"os"

	"github.com/julienschmidt/httprouter"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	_ = httprouter.Router{}
	httprouter.CleanPath("")
	httprouter.New()
}
`,
			Expected: `package main

import (
	instahttprouter "github.com/instana/go-sensor/instrumentation/instahttprouter"
	"github.com/julienschmidt/httprouter"
	"log"
	"os"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	_ = instahttprouter.WrappedRouter{}
	httprouter.CleanPath("")
	instahttprouter.Wrap(httprouter.New(), __instanaSensor)
}
`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()

			node, err := parser.ParseFile(fset, "", example.Code, 0)
			require.NoError(t, err)

			recipe := NewHttpRouter()

			instrumented, changed := recipe.Instrument(fset, node, example.TargetPkg, "__instanaSensor")
			assert.Equal(t, example.Changed, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}
