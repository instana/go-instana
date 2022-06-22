// (c) Copyright IBM Corp. 2022

package recipes_test

import (
	"bytes"
	"github.com/instana/go-instana/recipes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/format"
	"go/parser"
	"go/token"
	"testing"
)

func TestAWSSDKRecipe(t *testing.T) {

	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"New": {
			TargetPkg: "session",
			Code: `package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	session.New(&aws.Config{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")})
	session.New([]*aws.Config{
		{
			Region: aws.String("region"),
		},
		{
			Endpoint: aws.String("somestring"),
		},
	}...,
	)
}
`,
			Expected: `package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	instaawssdk "github.com/instana/go-sensor/instrumentation/instaawssdk"
)

func main() {
	instaawssdk.New(__instanaSensor, &aws.Config{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")})
	instaawssdk.New(__instanaSensor, []*aws.Config{{Region: aws.String("region")}, {Endpoint: aws.String("somestring")}}...)
}
`,
		},
		"NewSession": {
			TargetPkg: "session",
			Code: `package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	session.NewSession(&aws.Config{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")})
	session.NewSession([]*aws.Config{
		{
			Region: aws.String("region"),
		},
		{
			Endpoint: aws.String("somestring"),
		},
	}...,
	)
}
`,
			Expected: `package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	instaawssdk "github.com/instana/go-sensor/instrumentation/instaawssdk"
)

func main() {
	instaawssdk.NewSession(__instanaSensor, &aws.Config{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")})
	instaawssdk.NewSession(__instanaSensor, []*aws.Config{{Region: aws.String("region")}, {Endpoint: aws.String("somestring")}}...)
}
`,
		},
		"NewSessionWithOptions": {
			TargetPkg: "session",
			Code: `package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	session.NewSessionWithOptions(session.Options{})
}
`,
			Expected: `package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	instaawssdk "github.com/instana/go-sensor/instrumentation/instaawssdk"
)

func main() {
	instaawssdk.NewSessionWithOptions(__instanaSensor, session.Options{})
}
`,
		},
	}

	for name, example := range examples {

		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			fset.AddFile("tests", -1, 9)
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)

			require.NoError(t, err)

			instrumented, changed := recipes.NewAWSSDK().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			dumpExpectedCode(t, "awssdk", name, buf)

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func TestAWSSDKRecipeAlreadyInstrumented(t *testing.T) {

	examples := map[string]struct {
		TargetPkg string
		Code      string
	}{
		"New": {
			TargetPkg: "session",
			Code: `package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	instaawssdk "github.com/instana/go-sensor/instrumentation/instaawssdk"
)

func main() {
	instaawssdk.New(__instanaSensor, &aws.Config{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")})
	instaawssdk.New(__instanaSensor, []*aws.Config{{Region: aws.String("region")}, {Endpoint: aws.String("somestring")}}...)
}
`,
		},
		"NewSession": {
			TargetPkg: "session",
			Code: `package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	instaawssdk "github.com/instana/go-sensor/instrumentation/instaawssdk"
)

func main() {
	instaawssdk.NewSession(__instanaSensor, &aws.Config{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")})
	instaawssdk.NewSession(__instanaSensor, []*aws.Config{{Region: aws.String("region")}, {Endpoint: aws.String("somestring")}}...)
}
`,
		},
		"NewSessionWithOptions": {
			TargetPkg: "session",
			Code: `package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	instaawssdk "github.com/instana/go-sensor/instrumentation/instaawssdk"
)

func main() {
	instaawssdk.NewSessionWithOptions(__instanaSensor, session.Options{})
}
`,
		},
	}

	for name, example := range examples {

		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			fset.AddFile("tests", -1, 9)
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)

			require.NoError(t, err)

			instrumented, changed := recipes.NewAWSSDK().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Code, buf.String())
		})
	}
}
