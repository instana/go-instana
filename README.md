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