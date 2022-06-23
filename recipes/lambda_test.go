// (c) Copyright IBM Corp. 2022

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

func TestLambda(t *testing.T) {

	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"lambda.Start": {
			TargetPkg: "lambda",
			Code: `package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(func() {})
}
`,
			Expected: `package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	instalambda "github.com/instana/go-sensor/instrumentation/instalambda"
)

func main() {
	lambda.Start(instalambda.NewHandler(func() {
	}, __instanaSensor))
}
`,
		},
		"lambda.StartHandler": {
			TargetPkg: "lambda",
			Code: `package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.StartHandler(lambda.NewHandler(func() {}))
}
`,
			Expected: `package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	instalambda "github.com/instana/go-sensor/instrumentation/instalambda"
)

func main() {
	lambda.StartHandler(instalambda.WrapHandler(lambda.NewHandler(func() {
	}), __instanaSensor))
}
`,
		},
		"lambda.StartHandlerWithContext": {
			TargetPkg: "lambda",
			Code: `package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"

)

func main() {
	lambda.StartHandlerWithContext(context.Background(), lambda.NewHandler(func() {}))
}
`,
			Expected: `package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	instalambda "github.com/instana/go-sensor/instrumentation/instalambda"
)

func main() {
	lambda.StartHandlerWithContext(context.Background(), instalambda.WrapHandler(lambda.NewHandler(func() {
	}), __instanaSensor))
}
`,
		},
		"lambda.StartWithOptions": {
			TargetPkg: "lambda",
			Code: `package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.StartWithOptions(func() {}, lambda.WithContext(context.Background()))
}
`,
			Expected: `package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	instalambda "github.com/instana/go-sensor/instrumentation/instalambda"
)

func main() {
	lambda.StartWithOptions(instalambda.NewHandler(func() {
	}, __instanaSensor), lambda.WithContext(context.Background()))
}
`,
		},
		"lambda.StartWithContext": {
			TargetPkg: "lambda",
			Code: `package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.StartWithContext(context.Background(), func() {})
}
`,
			Expected: `package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	instalambda "github.com/instana/go-sensor/instrumentation/instalambda"
)

func main() {
	lambda.StartWithContext(context.Background(), instalambda.NewHandler(func() {
	}, __instanaSensor))
}
`,
		},
	}

	assertLambdaInstrumentation(t, examples)
}

func assertLambdaInstrumentation(t *testing.T, examples map[string]struct {
	TargetPkg string
	Code      string
	Expected  string
}) {
	for name, example := range examples {

		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)

			require.NoError(t, err)

			changed := recipes.NewLambda().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

			dumpExpectedCode(t, "lambda", name, buf)

			assert.Equal(t, example.Expected, buf.String())
		})

	}
}

func TestLambdaAlreadyInstrumented(t *testing.T) {

	examples := map[string]struct {
		TargetPkg string
		Code      string
	}{
		"lambda.Start": {
			TargetPkg: "lambda",
			Code: `package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	instalambda "github.com/instana/go-sensor/instrumentation/instalambda"
)

func main() {
	lambda.Start(instalambda.NewHandler(func() {
	}, __instanaSensor))
}
`,
		},
		"lambda.StartHandler": {
			TargetPkg: "lambda",
			Code: `package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	instalambda "github.com/instana/go-sensor/instrumentation/instalambda"
)

func main() {
	lambda.StartHandler(instalambda.WrapHandler(lambda.NewHandler(func() {
	}), __instanaSensor))
}
`,
		},
		"lambda.StartHandlerWithContext": {
			TargetPkg: "lambda",
			Code: `package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	instalambda "github.com/instana/go-sensor/instrumentation/instalambda"
)

func main() {
	lambda.StartHandlerWithContext(context.Background(), instalambda.WrapHandler(lambda.NewHandler(func() {
	}), __instanaSensor))
}
`,
		},
		"lambda.StartWithOptions": {
			TargetPkg: "lambda",
			Code: `package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	instalambda "github.com/instana/go-sensor/instrumentation/instalambda"
)

func main() {
	lambda.StartWithOptions(instalambda.NewHandler(func() {
	}, __instanaSensor), lambda.WithContext(context.Background()))
}
`,
		},
		"lambda.StartWithContext": {
			TargetPkg: "lambda",
			Code: `package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	instalambda "github.com/instana/go-sensor/instrumentation/instalambda"
)

func main() {
	lambda.StartWithContext(context.Background(), instalambda.NewHandler(func() {
	}, __instanaSensor))
}
`,
		},
	}

	for name, example := range examples {

		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)

			require.NoError(t, err)

			changed := recipes.NewLambda().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.False(t, changed)
		})
	}
}
