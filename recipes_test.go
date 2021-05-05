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

func TestNetHTTPRecipe(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"inline handler body": {
			TargetPkg: "http",
			Code: `http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNoContent)
})`,
			Expected: `http.HandleFunc("/", instana.TracingHandlerFunc(__instanaSensor, "/", func(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}))`,
		},
		"http.HandlerFunc variable": {
			TargetPkg: "http",
			Code:      `http.HandleFunc("/", http.NotFound)`,
			Expected:  `http.HandleFunc("/", instana.TracingHandlerFunc(__instanaSensor, "/", http.NotFound))`,
		},
		"http.Handle": {
			TargetPkg: "http",
			Code:      `http.Handle("/", http.FileServer(root))`,
			Expected:  `http.HandleFunc("/", instana.TracingHandlerFunc(__instanaSensor, "/", http.FileServer(root).ServeHTTP))`,
		},
		"aliased net/http": {
			TargetPkg: "custom",
			Code:      `custom.HandleFunc("/", custom.NotFound)`,
			Expected:  `custom.HandleFunc("/", instana.TracingHandlerFunc(__instanaSensor, "/", custom.NotFound))`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			node, err := parser.ParseExpr(example.Code)
			require.NoError(t, err)

			instrumented, changed := NetHTTPRecipe{
				InstanaPkg: "instana",
				TargetPkg:  example.TargetPkg,
				SensorVar:  "__instanaSensor",
			}.Instrument(node)

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func TestNetHTTPRecipe_Ignore(t *testing.T) {
	examples := map[string]string{
		"func differs":     `http.NewServeMux()`,
		"package differs":  `custom.HandleFunc("/", custom.NotFound)`,
		"non-package code": `log.Println("Hello")`,
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			node, err := parser.ParseExpr(example)
			require.NoError(t, err)

			result, changed := NetHTTPRecipe{
				InstanaPkg: "instana",
				TargetPkg:  "http",
				SensorVar:  "__instanaSensor",
			}.Instrument(node)

			require.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), result))

			assert.Equal(t, example, buf.String())
		})
	}
}

func TestNetHTTPRecipe_InstrumentedCode(t *testing.T) {
	examples := map[string]string{
		"inline handler body": `http.HandleFunc("/", instana.TracingHandlerFunc(__instanaSensor, "/", func(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}))`,
		"http.HandlerFunc variable": `http.HandleFunc("/", instana.TracingHandlerFunc(__instanaSensor, "/", http.NotFound))`,
		"http.Handle":               `http.HandleFunc("/", instana.TracingHandlerFunc(__instanaSensor, "/", http.FileServer(root).ServeHTTP))`,
		"aliased net/http":          `custom.HandleFunc("/", instana.TracingHandlerFunc(__instanaSensor, "/", custom.NotFound))`,
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			node, err := parser.ParseExpr(example)
			require.NoError(t, err)

			instrumented, changed := NetHTTPRecipe{
				InstanaPkg: "instana",
				TargetPkg:  "http",
				SensorVar:  "__instanaSensor",
			}.Instrument(node)

			assert.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example, buf.String())
		})
	}
}
