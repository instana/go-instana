// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package recipes_test

import (
	"bytes"
	"github.com/instana/go-instana/internal/recipes"
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
		"http.Client init": {
			TargetPkg: "http",
			Code:      `http.Client{Timeout: 5 * time.Second}`,
			Expected:  `http.Client{Timeout: 5 * time.Second, Transport: instana.RoundTripper(__instanaSensor, nil)}`,
		},
		"http.Client init with custom transport": {
			TargetPkg: "http",
			Code:      `http.Client{Timeout: 5 * time.Second, Transport: custom}`,
			Expected:  `http.Client{Timeout: 5 * time.Second, Transport: instana.RoundTripper(__instanaSensor, custom)}`,
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

			changed := recipes.NewNetHTTP().
				Instrument(nil, node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

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

			changed := recipes.NewNetHTTP().
				Instrument(nil, node, "http", "__instanaSensor")

			require.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

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
		"http.Client init":          `http.Client{Timeout: 5 * time.Second, Transport: instana.RoundTripper(__instanaSensor, custom)}`,
		"aliased net/http":          `custom.HandleFunc("/", instana.TracingHandlerFunc(__instanaSensor, "/", custom.NotFound))`,
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			node, err := parser.ParseExpr(example)
			require.NoError(t, err)

			changed := recipes.NewNetHTTP().
				Instrument(nil, node, "http", "__instanaSensor")

			assert.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

			assert.Equal(t, example, buf.String())
		})
	}
}
