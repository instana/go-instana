// (c) Copyright IBM Corp. 2022

package main

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstrument(t *testing.T) {
	availableInstrumentationPkgs := map[string]string{
		"github.com/instana/go-sensor":                                 "_",
		"github.com/instana/go-sensor/instrumentation/instagin":        "_",
		"github.com/instana/go-sensor/instrumentation/instahttprouter": "_",
	}

	originalCode := `package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func main() {

	gin.New()

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)

	log.Fatal(http.ListenAndServe(":8080", router))
}
`

	instrumentedCode := `package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
	instahttprouter "github.com/instana/go-sensor/instrumentation/instahttprouter"
	"github.com/julienschmidt/httprouter"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func main() {
	instagin.New(__instanaSensor)

	router := instahttprouter.Wrap(httprouter.New(), __instanaSensor)
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)

	log.Fatal(http.ListenAndServe(":8080", router))
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", originalCode, parser.AllErrors)

	require.NoError(t, err)

	Instrument(fset, f, "__instanaSensor", availableInstrumentationPkgs)

	buf := bytes.NewBuffer(nil)

	format.Node(buf, fset, f)

	assert.Equal(t, instrumentedCode, buf.String())
}
