package recipes_test

import (
	"bytes"
	"github.com/instana/go-instana/internal/recipes"
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

			changed := recipes.NewSarama().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

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

			changed := recipes.NewSarama().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

			dumpExpectedCode(t, "sarama", name, buf)

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func TestSarama_InstrumentUsingContext(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"UseContextForInstrumentation": {
			TargetPkg: "sarama",
			Code: `package main

import (
	"context"
	"github.com/Shopify/sarama"
)

func main() {
}
func Produce(ctx context.Context, useless int) {
	brokers := []string{"localhost:9092"}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_11_0_0
	producer, _ := sarama.NewSyncProducer(brokers, config)
	msg := &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")}
	producer.SendMessage(msg)
}
`,
			Expected: `package main

import (
	"context"
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
}
func Produce(ctx context.Context, useless int) {
	brokers := []string{"localhost:9092"}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_11_0_0
	producer, _ := instasarama.NewSyncProducer(brokers, config, __instanaSensor)
	msg := instasarama.ProducerMessageWithSpanFromContext(ctx, &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")})
	producer.SendMessage(instasarama.ProducerMessageWithSpanFromContext(ctx, msg))
}
`,
		},
		"UseContextForInstrumentationWithNamedContextImport": {
			TargetPkg: "sarama",
			Code: `package main

import (
	co "context"
	"github.com/Shopify/sarama"
)

func main() {
}
func Produce(ctx co.Context, useless int) {
	brokers := []string{"localhost:9092"}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_11_0_0
	producer, _ := sarama.NewSyncProducer(brokers, config)
	msg := &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")}
	producer.SendMessage(msg)
}
`,
			Expected: `package main

import (
	co "context"
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
}
func Produce(ctx co.Context, useless int) {
	brokers := []string{"localhost:9092"}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_11_0_0
	producer, _ := instasarama.NewSyncProducer(brokers, config, __instanaSensor)
	msg := instasarama.ProducerMessageWithSpanFromContext(ctx, &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")})
	producer.SendMessage(instasarama.ProducerMessageWithSpanFromContext(ctx, msg))
}
`,
		},
		"UseContextForInstrumentationProducerIsProvidedViaInterfaceSyncProducer": {
			TargetPkg: "sarama",
			Code: `package main

import (
	"context"

	"github.com/Shopify/sarama"
)

func main() {
	producer, _ := sarama.NewSyncProducer([]string{"localhost:9092"}, sarama.NewConfig())
	Produce(context.Background(), producer)
}
func Produce(ctx context.Context, producer sarama.SyncProducer) {
	msg := &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")}
	producer.SendMessage(msg)
}
`,
			Expected: `package main

import (
	"context"
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	producer, _ := instasarama.NewSyncProducer([]string{"localhost:9092"}, sarama.NewConfig(), __instanaSensor)
	Produce(context.Background(), producer)
}
func Produce(ctx context.Context, producer sarama.SyncProducer) {
	msg := instasarama.ProducerMessageWithSpanFromContext(ctx, &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")})
	producer.SendMessage(instasarama.ProducerMessageWithSpanFromContext(ctx, msg))
}
`,
		},
		"UseContextForInstrumentationProducerInTheGlobalScope": {
			TargetPkg: "sarama",
			Code: `package main

import (
	"context"

	"github.com/Shopify/sarama"
)

var producer sarama.SyncProducer

func main() {
	producer, _ = sarama.NewSyncProducer([]string{"localhost:9092"}, sarama.NewConfig())
	Produce(context.Background())
}
func Produce(ctx context.Context) {
	msg := &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")}
	producer.SendMessage(msg)
}
`,
			Expected: `package main

import (
	"context"
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

var producer sarama.SyncProducer

func main() {
	producer, _ = instasarama.NewSyncProducer([]string{"localhost:9092"}, sarama.NewConfig(), __instanaSensor)
	Produce(context.Background())
}
func Produce(ctx context.Context) {
	msg := instasarama.ProducerMessageWithSpanFromContext(ctx, &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")})
	producer.SendMessage(instasarama.ProducerMessageWithSpanFromContext(ctx, msg))
}
`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)

			require.NoError(t, err)

			changed := recipes.NewSarama().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

			dumpExpectedCode(t, "sarama", name, buf)

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func TestSarama_InstrumentUsingContext_AlreadyInstrumented(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Expected  string
	}{
		"AlreadyInstrumented 1": {
			TargetPkg: "sarama",
			Expected: `package main

import (
	"context"
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
}
func Produce(ctx context.Context, useless int) {
	brokers := []string{"localhost:9092"}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_11_0_0
	producer, _ := instasarama.NewSyncProducer(brokers, config, __instanaSensor)
	msg := instasarama.ProducerMessageWithSpanFromContext(ctx, &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")})
	producer.SendMessage(instasarama.ProducerMessageWithSpanFromContext(ctx, msg))
}
`,
		},
		"AlreadyInstrumented 2": {
			TargetPkg: "sarama",
			Expected: `package main

import (
	"context"
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
	producer, _ := instasarama.NewSyncProducer([]string{"localhost:9092"}, sarama.NewConfig(), __instanaSensor)
	Produce(context.Background(), producer)
}
func Produce(ctx context.Context, producer sarama.SyncProducer) {
	msg := instasarama.ProducerMessageWithSpanFromContext(ctx, &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")})
	producer.SendMessage(instasarama.ProducerMessageWithSpanFromContext(ctx, msg))
}
`,
		},
		"AlreadyInstrumented 3": {
			TargetPkg: "sarama",
			Expected: `package main

import (
	"context"
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

var producer sarama.SyncProducer

func main() {
	producer, _ = instasarama.NewSyncProducer([]string{"localhost:9092"}, sarama.NewConfig(), __instanaSensor)
	Produce(context.Background())
}
func Produce(ctx context.Context) {
	msg := instasarama.ProducerMessageWithSpanFromContext(ctx, &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")})
	producer.SendMessage(instasarama.ProducerMessageWithSpanFromContext(ctx, msg))
}
`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Expected, parser.AllErrors)

			require.NoError(t, err)

			changed := recipes.NewSarama().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.False(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

			dumpExpectedCode(t, "sarama", name, buf)

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}

func TestSarama_InstrumentUsingContextWithUnsupportedImport(t *testing.T) {
	examples := map[string]struct {
		TargetPkg string
		Code      string
		Expected  string
	}{
		"UseContextForInstrumentationWithUnsupportedContextImport_Blank": {
			TargetPkg: "sarama",
			Code: `package main

import (
	_ "context"
	"context"
	"github.com/Shopify/sarama"
)

func main() {
}
func Produce(ctx context.Context, useless int) {
	brokers := []string{"localhost:9092"}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_11_0_0
	producer, _ := sarama.NewSyncProducer(brokers, config)
	msg := &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")}
	producer.SendMessage(msg)
}
`,
			Expected: `package main

import (
	"context"
	_ "context"
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
}
func Produce(ctx context.Context, useless int) {
	brokers := []string{"localhost:9092"}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_11_0_0
	producer, _ := instasarama.NewSyncProducer(brokers, config, __instanaSensor)
	msg := &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")}
	producer.SendMessage(msg)
}
`,
		},
		"UseContextForInstrumentationWithUnsupportedContextImport_Dot": {
			TargetPkg: "sarama",
			Code: `package main

import (
	. "context"
	"github.com/Shopify/sarama"
)

func main() {
}
func Produce(ctx Context, useless int) {
	brokers := []string{"localhost:9092"}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_11_0_0
	producer, _ := sarama.NewSyncProducer(brokers, config)
	msg := &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")}
	producer.SendMessage(msg)
}
`,
			Expected: `package main

import (
	. "context"
	"github.com/Shopify/sarama"
	instasarama "github.com/instana/go-sensor/instrumentation/instasarama"
)

func main() {
}
func Produce(ctx Context, useless int) {
	brokers := []string{"localhost:9092"}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_11_0_0
	producer, _ := instasarama.NewSyncProducer(brokers, config, __instanaSensor)
	msg := &sarama.ProducerMessage{Topic: "test-topic-1", Offset: sarama.OffsetNewest, Value: sarama.StringEncoder("I am a message")}
	producer.SendMessage(msg)
}
`,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test", example.Code, parser.AllErrors)

			require.NoError(t, err)

			changed := recipes.NewSarama().
				Instrument(token.NewFileSet(), node, example.TargetPkg, "__instanaSensor")

			assert.True(t, changed)

			buf := bytes.NewBuffer(nil)
			require.NoError(t, format.Node(buf, token.NewFileSet(), node))

			dumpExpectedCode(t, "sarama", name, buf)

			assert.Equal(t, example.Expected, buf.String())
		})
	}
}
