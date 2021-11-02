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

func TestGRPCRecipe(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"new server with no parameters": {
			TargetPkg: "grpc",
			Code:      `grpc.NewServer()`,
			Expected:  `grpc.NewServer(grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))`,
		},
		"new server only with UnaryInterceptor parameter": {
			TargetPkg: "grpc",
			Code:      `grpc.NewServer(grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)))`,
			Expected:  `grpc.NewServer(grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))`,
		},
		"new server only with StreamServerInterceptor parameter": {
			TargetPkg: "grpc",
			Code:      `grpc.NewServer(grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))`,
			Expected:  `grpc.NewServer(grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)), grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)))`,
		},
		"new server only with extra UnaryInterceptor": {
			TargetPkg: "grpc",
			Code:      `grpc.NewServer(grpc.UnaryInterceptor(someFunc()))`,
			Expected:  `grpc.NewServer(grpc.UnaryInterceptor(someFunc()), grpc.ChainUnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))`,
		},
		"new server only with StreamInterceptor parameter": {
			TargetPkg: "grpc",
			Code:      `grpc.NewServer(grpc.StreamInterceptor(someFunc()))`,
			Expected:  `grpc.NewServer(grpc.StreamInterceptor(someFunc()), grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.ChainStreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))`,
		},
		"new server only with UnaryInterceptor parameter and wrong sensor variable name": {
			TargetPkg: "grpc",
			Code:      `grpc.NewServer(grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(WRONG_VAR_NAME)))`,
			Expected:  `grpc.NewServer(grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(WRONG_VAR_NAME)), grpc.ChainUnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))`,
		},
		"new server only with StreamServerInterceptor parameter and wrong sensor variable name": {
			TargetPkg: "grpc",
			Code:      `grpc.NewServer(grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(WRONG_VAR_NAME)))`,
			Expected:  `grpc.NewServer(grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(WRONG_VAR_NAME)), grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.ChainStreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor)))`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			node, err := parser.ParseExpr(example.Code)
			require.NoError(t, err)

			instrumented, changed := recipes.GRPC{
				InstanaPkg: "instagrpc",
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

func TestGRPCRecipe_Ignore_Already_Instrumented(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
		Changed   bool
	}{
		"Interceptor defined as external var 1": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	var a = grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor))
	var b = grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))
	grpc.NewServer(a, b)
}
`,
			Expected: `package main

func main() {
	var a = grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor))
	var b = grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))
	grpc.NewServer(a, b)
}
`,
			Changed: false,
		},
		"Interceptor defined as external var 2": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	var a = grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor))
	var b = grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))
	var c = a
	grpc.NewServer(c, b)
}
`,
			Expected: `package main

func main() {
	var a = grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor))
	var b = grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))
	var c = a
	grpc.NewServer(c, b)
}
`,
			Changed: false,
		},
		"Interceptor defined as external var 3": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	var a = grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor))
	var b = grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))
	var c = a
	var d = b
	grpc.NewServer(c, d)
}
`,
			Expected: `package main

func main() {
	var a = grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor))
	var b = grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))
	var c = a
	var d = b
	grpc.NewServer(c, d)
}
`,
			Changed: false,
		},

		"Interceptor defined as external var 4": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	var a = grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor))
	var b = grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))
	grpc.NewServer(func() grpc.ServerOption {
		return a
	}(), b)
}
`,
			Expected: `package main

func main() {
	var a = grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor))
	var b = grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))
	grpc.NewServer(func() grpc.ServerOption {
		return a
	}(), b)
}
`,
			Changed: false,
		},
		"Interceptor defined as external var 5": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	var a = grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor))
	var b = grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))
	grpc.NewServer(func() grpc.ServerOption {
		return *(&a)
	}(), *(&b))
}
`,
			Expected: `package main

func main() {
	var a = grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor))
	var b = grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))
	grpc.NewServer(func() grpc.ServerOption {
		return *(&a)
	}(), *(&b))
}
`,
			Changed: false,
		},
		"Interceptor defined as external var 6": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	var a = grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor))
	var b = grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))
	grpc.NewServer(func() grpc.ServerOption {
		return *&a
	}(), *&b)
}
`,
			Expected: `package main

func main() {
	var a = grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor))
	var b = grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))
	grpc.NewServer(func() grpc.ServerOption {
		return *&a
	}(), *&b)
}
`,
			Changed: false,
		},
		"Interceptor defined as external var 7": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	arr := []grpc.ServerOption{grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))}
	srv := grpc.NewServer(arr[0], *&arr[1])
}
`,
			Expected: `package main

func main() {
	arr := []grpc.ServerOption{grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))}
	srv := grpc.NewServer(arr[0], *&arr[1])
}
`,
			Changed: false,
		},
		"Interceptor defined as external var 8": {
			TargetPkg: "grpc",
			Code: `package main

func main() {
	arr := []grpc.ServerOption{grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))}
	a := func() grpc.ServerOption {
		return arr[1]
	}
	b := func() *grpc.ServerOption {
		return &arr[0]
	}
	srv := grpc.NewServer(a(), *b())
}
`,
			Expected: `package main

func main() {
	arr := []grpc.ServerOption{grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))}
	a := func() grpc.ServerOption {
		return arr[1]
	}
	b := func() *grpc.ServerOption {
		return &arr[0]
	}
	srv := grpc.NewServer(a(), *b())
}
`,
			Changed: false,
		},
		"Interceptor defined as external var 9": {
			TargetPkg: "grpc",
			Code: `package main

import (
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	"google.golang.org/grpc"
)

var arr = []grpc.ServerOption{grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))}

func a() grpc.ServerOption {
	return arr[0]
}
func b() *grpc.ServerOption {
	return &arr[1]
}
func main() {
	_ = grpc.NewServer(a(), *b())
}
`,
			Expected: `package main

import (
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	"google.golang.org/grpc"
)

var arr = []grpc.ServerOption{grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(__instanaSensor)), grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(__instanaSensor))}

func a() grpc.ServerOption {
	return arr[0]
}
func b() *grpc.ServerOption {
	return &arr[1]
}
func main() {
	_ = grpc.NewServer(a(), *b())
}
`,
			Changed: false,
		},
		"Interceptor defined as external var 10": {
			TargetPkg: "grpc",
			Code: `package main

import (
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	"google.golang.org/grpc"
)

func a() grpc.UnaryServerInterceptor {
	return instagrpc.UnaryServerInterceptor(__instanaSensor)
}
func b() grpc.StreamServerInterceptor {
	return instagrpc.StreamServerInterceptor(__instanaSensor)
}
func main() {
	grpc.NewServer(grpc.UnaryInterceptor(a()), grpc.StreamInterceptor(b()))
}
`,
			Expected: `package main

import (
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	"google.golang.org/grpc"
)

func a() grpc.UnaryServerInterceptor {
	return instagrpc.UnaryServerInterceptor(__instanaSensor)
}
func b() grpc.StreamServerInterceptor {
	return instagrpc.StreamServerInterceptor(__instanaSensor)
}
func main() {
	grpc.NewServer(grpc.UnaryInterceptor(a()), grpc.StreamInterceptor(b()))
}
`,
			Changed: false,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			node, err := parser.ParseFile(token.NewFileSet(), "test", example.Code, parser.AllErrors)
			require.NoError(t, err)

			instrumented, changed := recipes.GRPC{
				InstanaPkg: "instagrpc",
				TargetPkg:  example.TargetPkg,
				SensorVar:  "__instanaSensor",
			}.Instrument(node)

			assert.Equal(t, example.Changed, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}
