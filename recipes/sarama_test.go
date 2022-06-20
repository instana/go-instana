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

func TestSarama_Instrument(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"NewAsyncProducer": {
			TargetPkg: "sarama",
			Code: `package main

import "github.com/Shopify/sarama"

func main() {
	config := sarama.NewConfig()
	_, _ = sarama.NewAsyncProducer([]string{"localhost"}, config)
}

`,
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	_, _ = instasarama.NewAsyncProducer([]string{"localhost"}, config, __instanaSensor)
}
`,
		},
		"NewAsyncProducerFromClient": {
			TargetPkg: "sarama",
			Code: `package main

import "github.com/Shopify/sarama"

func main() {
	config := sarama.NewConfig()
	c, _ := sarama.NewClient([]string{"localhost"}, config)
	_, _ = sarama.NewAsyncProducerFromClient(c)
}
`,
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	c, _ := sarama.NewClient([]string{"localhost"}, config)
	_, _ = instasarama.NewAsyncProducerFromClient(c, __instanaSensor)
}
`,
		},
		"NewConsumer": {
			TargetPkg: "sarama",
			Code: `package main

import (
	"github.com/Shopify/sarama"
)

func main() {
	config := sarama.NewConfig()
	_, _ = sarama.NewConsumer([]string{"localhost"}, config)
}
`,
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	_, _ = instasarama.NewConsumer([]string{"localhost"}, config, __instanaSensor)
}
`,
		},
		"NewConsumerFromClient": {
			TargetPkg: "sarama",
			Code: `package main

import "github.com/Shopify/sarama"

func main() {
	config := sarama.NewConfig()
	c, _ := sarama.NewClient([]string{"localhost"}, config)
	_, _ = sarama.NewConsumerFromClient(c)
}
`,
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	c, _ := sarama.NewClient([]string{"localhost"}, config)
	_, _ = instasarama.NewConsumerFromClient(c, __instanaSensor)
}
`,
		},
		"NewSyncProducer": {
			TargetPkg: "sarama",
			Code: `package main

import "github.com/Shopify/sarama"

func main() {
	config := sarama.NewConfig()
	_, _ = sarama.NewSyncProducer([]string{"localhost"}, config)
}

`,
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	_, _ = instasarama.NewSyncProducer([]string{"localhost"}, config, __instanaSensor)
}
`,
		},
		"NewSyncProducerFromClient": {
			TargetPkg: "sarama",
			Code: `package main

import "github.com/Shopify/sarama"

func main() {
	config := sarama.NewConfig()
	c, _ := sarama.NewClient([]string{"localhost"}, config)
	_, _ = sarama.NewSyncProducerFromClient(c)
}
`,
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	c, _ := sarama.NewClient([]string{"localhost"}, config)
	_, _ = instasarama.NewSyncProducerFromClient(c, __instanaSensor)
}
`,
		},
		"NewConsumerGroup": {
			TargetPkg: "sarama",
			Code: `package main

import "github.com/Shopify/sarama"

func main() {
	config := sarama.NewConfig()
	_, _ = sarama.NewConsumerGroup([]string{"localhost"}, "g1", config)
}
`,
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	_, _ = instasarama.NewConsumerGroup([]string{"localhost"}, "g1", config, __instanaSensor)
}
`,
		},
		"NewConsumerGroupFromClient": {
			TargetPkg: "sarama",
			Code: `package main

import "github.com/Shopify/sarama"

func main() {
	config := sarama.NewConfig()
	c, _ := sarama.NewClient([]string{"localhost"}, config)
	_, _ = sarama.NewConsumerGroupFromClient("g1", c)
}
`,
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	c, _ := sarama.NewClient([]string{"localhost"}, config)
	_, _ = instasarama.NewConsumerGroupFromClient("g1", c, __instanaSensor)
}
`,
		},
	}

	assertSaramaInstrumentation(t, examples)
}

func TestSarama_AlreadyInstrument(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Expected  string
	}{
		"NewAsyncProducer": {
			TargetPkg: "sarama",
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	_, _ = instasarama.NewAsyncProducer([]string{"localhost"}, config, __instanaSensor)
}
`,
		},
		"NewAsyncProducerFromClient": {
			TargetPkg: "sarama",
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	c, _ := sarama.NewClient([]string{"localhost"}, config)
	_, _ = instasarama.NewAsyncProducerFromClient(c, __instanaSensor)
}
`,
		},
		"NewConsumer": {
			TargetPkg: "sarama",
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	_, _ = instasarama.NewConsumer([]string{"localhost"}, config, __instanaSensor)
}
`,
		},
		"NewConsumerFromClient": {
			TargetPkg: "sarama",
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	c, _ := sarama.NewClient([]string{"localhost"}, config)
	_, _ = instasarama.NewConsumerFromClient(c, __instanaSensor)
}
`,
		},
		"NewSyncProducer": {
			TargetPkg: "sarama",
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	_, _ = instasarama.NewSyncProducer([]string{"localhost"}, config, __instanaSensor)
}
`,
		},
		"NewSyncProducerFromClient": {
			TargetPkg: "sarama",
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	c, _ := sarama.NewClient([]string{"localhost"}, config)
	_, _ = instasarama.NewSyncProducerFromClient(c, __instanaSensor)
}
`,
		},
		"NewConsumerGroup": {
			TargetPkg: "sarama",
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	_, _ = instasarama.NewConsumerGroup([]string{"localhost"}, "g1", config, __instanaSensor)
}
`,
		},
		"NewConsumerGroupFromClient": {
			TargetPkg: "sarama",
			Expected: `package main

import (
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	config := sarama.NewConfig()
	c, _ := sarama.NewClient([]string{"localhost"}, config)
	_, _ = instasarama.NewConsumerGroupFromClient("g1", c, __instanaSensor)
}
`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Expected, parser.AllErrors)

			require.NoError(t, err)

			instrumented, changed := recipes.NewSarama().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func assertSaramaInstrumentation(t *testing.T, examples map[string]struct {
	TargetPkg string
	Code      string
	Expected  string
}) {
	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)

			require.NoError(t, err)

			instrumented, changed := recipes.NewSarama().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), instrumented))

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}
