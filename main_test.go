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
	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
)

func main() {
	gin.New()

	router := httprouter.New()
}
`

	instrumentedCode := `package main

import (
	"github.com/gin-gonic/gin"
	instagin "github.com/instana/go-sensor/instrumentation/instagin"
	instahttprouter "github.com/instana/go-sensor/instrumentation/instahttprouter"
	"github.com/julienschmidt/httprouter"
)

func main() {
	instagin.New(__instanaSensor)

	router := instahttprouter.Wrap(httprouter.New(), __instanaSensor)
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", originalCode, parser.AllErrors)

	require.NoError(t, err)

	instrument(fset, f, "__instanaSensor", availableInstrumentationPkgs)

	buf := bytes.NewBuffer(nil)

	assert.NoError(t, format.Node(buf, fset, f))

	assert.Equal(t, instrumentedCode, buf.String())
}
