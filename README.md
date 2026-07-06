# Componentize SDK

A Go utility package for building [WebAssembly components](https://component-model.bytecodealliance.org/) using the [`componentize-go`](https://github.com/bytecodealliance/componentize-go) and `wit-bindgen go` toolchain. It provides idiomatic Go abstractions over WASI interfaces including HTTP servers and clients, TCP sockets, and blob storage.

## Overview

The SDK wraps low-level WASI bindings so you can write WebAssembly components in standard Go without dealing with the generated glue code directly. Components built with this SDK can run on any WASI-compatible runtime such as [wasmtime](https://wasmtime.dev/).

Key capabilities:

- **HTTP server** — register a standard `http.Handler` that runs inside a WASM component
- **HTTP client** — make outbound HTTP requests via `http.Client` with a WASI transport
- **TCP sockets** — dial and listen using Go's `net.Conn` / `net.Listener` interfaces
- **Blob storage** — read and write objects via a WASI blobstore abstraction

## Getting Started

### Prerequisites

Install the following tools before working with this SDK or its examples:

- **Go** 1.21+ — <https://go.dev/dl/>
- **componentize-go** — builds Go source into a `.wasm` component: `go install go.bytecodealliance.org/cmd/componentize-go@latest`
- **wasmtime** — runs and serves WASM components: <https://wasmtime.dev/>
- **wkg** *(optional, for regenerating bindings)* — WIT package manager: <https://github.com/bytecodealliance/wkg>

### Installation

Add the SDK to your Go module:

```bash
go get github.com/jamesstocktonj1/componentize-sdk
```

### Writing your first HTTP component

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/jamesstocktonj1/componentize-sdk/net/wasihttp"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello, World!")
    })
    wasihttp.Handle(mux)
}
```

Build and run it:

```bash
componentize-go build --world http --output hello.wasm .
wasmtime serve --addr 0.0.0.0:8000 hello.wasm
curl http://localhost:8000/hello
```

Your `go.mod` must reference a WIT world that exports the WASI HTTP incoming-handler interface. See the `wit/` directory in this repo for the world definitions used by the examples.

## Running the Examples

Each example lives under `examples/` and has its own `Makefile` with `build` and `run` targets.

### hello — HTTP server

A simple HTTP server with `/hello`, `/echo`, and `/greet` endpoints.

```bash
cd examples/hello
make build   # produces build/hello.wasm
make run     # serves on http://localhost:8000
```

```bash
curl http://localhost:8000/hello
curl -X POST http://localhost:8000/echo -d "ping"
curl -X POST http://localhost:8000/greet -d '{"name":"Alice"}'
```

### client — outbound HTTP client

Demonstrates making outbound HTTP requests from within a WASM component.

```bash
cd examples/client
make build
make run
```

### socket-server — TCP socket server

Listens on TCP port 7777 and transforms incoming text to leet speak.

```bash
cd examples/socket-server
make build
make run     # listens on TCP :7777
```

```bash
echo "hello" | nc localhost 7777
```

### socket-client — TCP socket client

An HTTP endpoint that connects to a TCP socket and relays data.

```bash
cd examples/socket-client
make build
make run
```

### blobby — blob storage

Stores the HTTP request body in a WASI blobstore and reads it back.

```bash
cd examples/blobby
make build
make run
```

### P3 variants

The `hello-p3` and `client-p3` examples use the older WASI Preview 3 bindings found in `p3/`. Build and run them the same way as their non-P3 counterparts.

## Project Structure

```
cli/            CLI entrypoint helper (SetRun / SetRunE)
net/
  wasihttp/     HTTP server handler and outbound client transport
  socket/       TCP Dial and Listen backed by WASI sockets
file/
  blobstore/    Container and object abstractions for WASI blobstore
internal/       Shared async I/O polling and stream utilities
gen/            Auto-generated WASI bindings (do not edit by hand)
wit/            WIT world and interface definitions
p3/             Bindings and examples for WASI Preview 3
examples/       Runnable example components
```

## Regenerating Bindings

If you update the WIT definitions, regenerate the Go bindings:

```bash
make fetch      # pull latest WIT dependencies via wkg
make generate   # run componentize-go bindings to regenerate gen/
```

## Contributing

Contributions are welcome. Please follow the steps below:

1. Fork the repository and create a feature branch from `main`.
2. Make your changes and ensure the examples still build (`make build` in each relevant example directory).
3. Keep commits focused and write clear commit messages.
4. Open a pull request describing what changed and why.

There are currently no automated unit tests; the examples act as functional tests. If you add a new feature, consider adding or extending an example that exercises it.

For larger changes or new WASI interface support, open an issue first to discuss the approach.

## License

MIT — see [LICENSE](LICENSE).
