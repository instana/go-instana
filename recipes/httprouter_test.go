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

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)

	log.Fatal(http.ListenAndServe(":8080", router))
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

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}
func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}
func main() {
	router := instahttprouter.Wrap(httprouter.New(), __instanaSensor)
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)
	log.Fatal(http.ListenAndServe(":8080", router))
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

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func useRouterInstance(r httprouter.Router) {
	r.GET("/", Index)
}

func main() {
	var router *httprouter.Router
	router = httprouter.New()
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)

	log.Fatal(http.ListenAndServe(":8080", router))
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

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}
func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}
func useRouterInstance(r instahttprouter.WrappedRouter) {
	r.GET("/", Index)
}
func main() {
	var router *instahttprouter.WrappedRouter
	router = instahttprouter.Wrap(httprouter.New(), __instanaSensor)
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)
	log.Fatal(http.ListenAndServe(":8080", router))
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

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}
func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}
func useRouterInstance(r instahttprouter.WrappedRouter) {
	r.GET("/", Index)
}
func main() {
	var router *instahttprouter.WrappedRouter
	router = instahttprouter.Wrap(httprouter.New(), __instanaSensor)
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)
	log.Fatal(http.ListenAndServe(":8080", router))
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

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}
func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}
func useRouterInstance(r instahttprouter.WrappedRouter) {
	r.GET("/", Index)
}
func main() {
	var router *instahttprouter.WrappedRouter
	router = instahttprouter.Wrap(httprouter.New(), __instanaSensor)
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)
	log.Fatal(http.ListenAndServe(":8080", router))
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
