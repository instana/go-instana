// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package recipes_test

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"testing"

	"github.com/instana/go-instana/recipes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRPCServerRecipe(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"new server with no parameters": {
			TargetPkg: "grpc",
			Code:      `grpc.NewServer()`,
			Expected:  `grpc.NewServer(grpc.ChainStreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)), grpc.ChainUnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)))`,
		},
		"new server only with extra UnaryInterceptor": {
			TargetPkg: "grpc",
			Code:      `grpc.NewServer(grpc.UnaryInterceptor(someFunc()))`,
			Expected:  `grpc.NewServer(grpc.ChainStreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)), grpc.ChainUnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.UnaryInterceptor(someFunc()))`,
		},
		"new server only with StreamInterceptor parameter": {
			TargetPkg: "grpc",
			Code:      `grpc.NewServer(grpc.StreamInterceptor(someFunc()))`,
			Expected:  `grpc.NewServer(grpc.ChainStreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)), grpc.ChainUnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(someFunc()))`,
		},
		"new server only with UnaryInterceptor parameter and wrong sensor variable name": {
			TargetPkg: "grpc",
			Code:      `grpc.NewServer(grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(WRONG_VAR_NAME)))`,
			Expected:  `grpc.NewServer(grpc.ChainStreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)), grpc.ChainUnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(WRONG_VAR_NAME)))`,
		},
		"new server only with StreamServerInterceptor parameter and wrong sensor variable name": {
			TargetPkg: "grpc",
			Code:      `grpc.NewServer(grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(WRONG_VAR_NAME)))`,
			Expected:  `grpc.NewServer(grpc.ChainStreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)), grpc.ChainUnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(WRONG_VAR_NAME)))`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			node, err := parser.ParseExpr(example.Code)
			require.NoError(t, err)

			instrumented, changed := recipes.NewGRPC().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func TestGRPCServerRecipe_Ignore_Already_Instrumented(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
		Changed   bool
	}{
		"new server only with UnaryInterceptor parameter": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	grpc.NewServer(grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)))
}`,
			Expected: `package main

func main() {
	grpc.NewServer(grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)))
}
`,
			Changed: false,
		},
		"new server only with StreamServerInterceptor parameter": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	grpc.NewServer(grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))
}`,
			Expected: `package main

func main() {
	grpc.NewServer(grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))
}
`,
			Changed: false,
		},
		"Already instrumented 1": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	grpc.NewServer(grpc.ChainUnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))
}
`,
			Expected: `package main

func main() {
	grpc.NewServer(grpc.ChainUnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))
}
`,
			Changed: false,
		},
		"Already instrumented 2": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	grpc.NewServer(grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))
}
`,
			Expected: `package main

func main() {
	grpc.NewServer(grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))
}
`,
			Changed: false,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)
			require.NoError(t, err)

			instrumented, changed := recipes.NewGRPC().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.Equal(t, example.Changed, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func TestGRPCClientRecipe(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"new client with no extra parameters": {
			TargetPkg: "grpc",
			Code:      `grpc.Dial("localhost")`,
			Expected:  `grpc.Dial("localhost", grpc.WithChainStreamInterceptor(instagrpc.StreamClientInterceptor(__instanaSensor)), grpc.WithChainUnaryInterceptor(instagrpc.UnaryClientInterceptor(__instanaSensor)))`,
		},
		"new client only with extra WithUnaryInterceptor and wrong parameter": {
			TargetPkg: "grpc",
			Code:      `grpc.Dial("localhost", grpc.WithUnaryInterceptor(someFunc()))`,
			Expected:  `grpc.Dial("localhost", grpc.WithChainStreamInterceptor(instagrpc.StreamClientInterceptor(__instanaSensor)), grpc.WithChainUnaryInterceptor(instagrpc.UnaryClientInterceptor(__instanaSensor)), grpc.WithUnaryInterceptor(someFunc()))`,
		},
		"new client only with WithStreamInterceptor and wrong parameter": {
			TargetPkg: "grpc",
			Code:      `grpc.Dial("localhost", grpc.WithStreamInterceptor(someFunc()))`,
			Expected:  `grpc.Dial("localhost", grpc.WithChainStreamInterceptor(instagrpc.StreamClientInterceptor(__instanaSensor)), grpc.WithChainUnaryInterceptor(instagrpc.UnaryClientInterceptor(__instanaSensor)), grpc.WithStreamInterceptor(someFunc()))`,
		},
		"new client only with WithUnaryInterceptor parameter and wrong sensor variable name": {
			TargetPkg: "grpc",
			Code:      `grpc.Dial("localhost", grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(WRONG_VAR_NAME)))`,
			Expected:  `grpc.Dial("localhost", grpc.WithChainStreamInterceptor(instagrpc.StreamClientInterceptor(__instanaSensor)), grpc.WithChainUnaryInterceptor(instagrpc.UnaryClientInterceptor(__instanaSensor)), grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(WRONG_VAR_NAME)))`,
		},
		"new server only with WithStreamInterceptor parameter and wrong sensor variable name": {
			TargetPkg: "grpc",
			Code:      `grpc.Dial("localhost", grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(WRONG_VAR_NAME)))`,
			Expected:  `grpc.Dial("localhost", grpc.WithChainStreamInterceptor(instagrpc.StreamClientInterceptor(__instanaSensor)), grpc.WithChainUnaryInterceptor(instagrpc.UnaryClientInterceptor(__instanaSensor)), grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(WRONG_VAR_NAME)))`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			node, err := parser.ParseExpr(example.Code)
			require.NoError(t, err)

			instrumented, changed := recipes.NewGRPC().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func TestGRPCClientRecipe_Ignore_Already_Instrumented(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
		Changed   bool
	}{
		"new client only with WithUnaryInterceptor parameter": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	grpc.Dial("localhost", grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(__instanaSensor)))
}`,
			Expected: `package main

func main() {
	grpc.Dial("localhost", grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(__instanaSensor)))
}
`,
			Changed: false,
		},
		"new client only with WithStreamInterceptor parameter": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	grpc.Dial("localhost", grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(__instanaSensor)))
}
`,
			Expected: `package main

func main() {
	grpc.Dial("localhost", grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(__instanaSensor)))
}
`,
			Changed: false,
		},
		"Already instrumented 1": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	conn, err := grpc.Dial(listenAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(__instanaSensor)), grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(__instanaSensor)))
}
`,
			Expected: `package main

func main() {
	conn, err := grpc.Dial(listenAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(__instanaSensor)), grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(__instanaSensor)))
}
`,
			Changed: false,
		},
		"Already instrumented 2": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	conn, err := grpc.Dial(listenAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithChainUnaryInterceptor(instagrpc.UnaryClientInterceptor(__instanaSensor)), grpc.WithChainStreamInterceptor(instagrpc.StreamClientInterceptor(__instanaSensor)))
}
`,
			Expected: `package main

func main() {
	conn, err := grpc.Dial(listenAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithChainUnaryInterceptor(instagrpc.UnaryClientInterceptor(__instanaSensor)), grpc.WithChainStreamInterceptor(instagrpc.StreamClientInterceptor(__instanaSensor)))
}
`,
			Changed: false,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)
			require.NoError(t, err)

			instrumented, changed := recipes.NewGRPC().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.Equal(t, example.Changed, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}
