package main

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
	"golang.org/x/tools/go/ast/astutil"
)

func TestInstrument(t *testing.T) {
	node, err := parser.ParseExpr(simpleHTTPServer)
	require.NoError(t, err)

	res := astutil.Apply(node, Instrument(map[string]string{}), nil)

	buf := bytes.NewBuffer(nil)
	require.NoError(t, format.Node(buf, token.NewFileSet(), res))

	assert.Equal(t, strings.ReplaceAll(simpleHTTPServer, "http.HandleFunc", "instana.HandleFunc"), buf.String())
}
