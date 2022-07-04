Go Instana
==========

An Instana instrumentation tool for Go applications. `go-instana` is designed to work as a part of
Go toolchain, applying instrumentation patches during the compile time.

Installation
------------

`go-instana` requires go1.18+ and can be installed to `$GOPATH/bin` with

```bash
$ go install github.com/instana/go-instana
```

You might consider adding `$GOPATH/bin` to your `PATH` to avoid typing it every time to invoke
`go-instana`:

```bash
$ export PATH="$GOPATH/bin:$PATH"
```

Usage
-----

The instrumentation process is done in two stages. Both steps are idempotent and can be added to your
CI build pipeline:

1. **Initialization**, when `go-instana` ensures there is an Instana sensor present in all packages
   of the project. It also adds imports for all instrumentation packages that are necessary. To apply this step run the 
   following command from the root folder of your project

   ```bash
   $ go-instana add
   ```

   At this point your application is already instrumented and ready to send metrics to Instana.
   However, in order to collect trace information some minor modifications to your code are still
   required. These changes are applied during the instrumentation stage, which is the second step.

2. **Download added dependencies** `go-instana` will add necessary instrumentation packages to your code. Run `go mod tidy` to download and add them to your `go.mod`.

3. **Instrumentation**, when `go-instana` searches for Instana sensor in the package global scope
   and applies instrumentation patches that add Instana code wrappers where necessary.

   This step is a part of `go build` command sequence and done by specifying `go-instana` as a
   value for `-toolexec` flag:

   ``` bash
   $ go build -toolexec="go-instana"
   ```
   In case there is no Instana sensor available in the global scope, no changes are applied.

   You can also provide the `-toolexec` for all `go build` commands by adding it to the `GOFLAGS`
   environment variable:

   ```bash
   $ export GOFLAGS='-toolexec="go-instana"'
   $ go build # will use go-instana to build your app
   ```
   To apply instrumentation without building the binary, run `go-instana instrument` from the module's root directory.

To see which packages might be instrumented, use `go-instana list`. 

To exclude packages from the instrumentation list use `e` flag. For example: `go-instana -e db -e sql list`.

# Instrumentations

This section describes which libraries are supported by this tool. Also, it gives an understanding of which transformation
will be applied to the code and which patterns can be instrumented.

## `database/sql`

This will reuse the already registered DB driver and wrap it with the necessary code to instrument.

It will replace `Open` calls from the original package with `SQLInstrumentAndOpen` from the instrumentation.

## `github.com/Shopify/sarama`

This instrumentation will replace, with instrumented analog, following functions:

* `NewAsyncProducer`
* `NewAsyncProducerFromClient`
* `NewConsumer`
* `NewConsumerFromClient`
* `NewSyncProducer`
* `NewSyncProducerFromClient`
* `NewConsumerGroup`
* `NewConsumerGroupFromClient`

This will not provide a continuation of the trace but will be sufficient to display the correlation between
producer and consumer in isolation, if both parts are instrumented.

To enable an auto trace propagation to the consumer, please ensure that at least one of the following is true:

1. Producer messages are created within a function that has a tracing context in the parameters list.
   Example :
```
// ctx should contain a tracing information. 
// For instance: it can be a received context from the instrumented http handler.
func constructMyMessage(ctx context.Context) {
    ...
    msg := &sarama.ProducerMessage{
        Topic:  "test-topic-1",
        Offset: sarama.OffsetNewest,
        Value:  sarama.StringEncoder("I am a message"),
    }
    ...
}
```
2. Producer's method `SendMessage` is called within a function that has a tracing context in the parameters list.
   Example :
```
// ctx should contain a tracing information. 
// For instance: it can be a received context from the instrumented http handler.
func Produce(ctx context.Context) {
	...
	producer.SendMessage(msg)
	...
}
```

Tracing context will not be automatically extracted on the consumer side. This has to be done manually.
Check this [example](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama#example-package-Consumer)
for details.

## `github.com/aws/aws-lambda-go/lambda`

This will instrument the following methods from the package:

* `Start`
* `StartHandler`
* `StartHandlerWithContext`
* `StartWithOptions`
* `StartWithContext`

It will wrap the handler in the argument's list with an instrumentation call.

## `github.com/aws/aws-sdk-go/aws/session`

It substitutes with instrumented original calls:

* `New`
* `NewSession`
* `NewSessionWithOptions`

## `github.com/gin-gonic/gin`

Replaces call `New` and/or `Default` with calls that return instrumented gin instance.

## `github.com/gorilla/mux`

Replaces call `NewRouter` with a call that returns instrumented mux instance.

## `github.com/julienschmidt/httprouter`

Replaces everywhere `httprouter.Router` type with `instahttprouter.WrappedRouter`.
Also, wraps `New` with an instrumentation call.

## `github.com/labstack/echo/v4`

Replaces call `New` with a call that returns instrumented echo instance.

## `go.mongodb.org/mongo-driver/mongo`

Replaces `NewClient` and `Connect` calls with instrumented analog.

## `google.golang.org/grpc`

The server is instrumented by modifying `NewServer` parameters.
The client is instrumented by modifying `Dial` parameters.

## `net/http`

Http handlers are instrumented by wrapping `http.HandleFunc` and/or `http.Handle` handler parameter.

Clients are instrumented only if they are created with instantiated like this `client := http.Client{}` in the code.
The default client will be not instrumented.