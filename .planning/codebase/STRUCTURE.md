# Codebase Structure

**Analysis Date:** 2026-03-16

## Directory Layout

```
componentize-sdk/
├── examples/              # Example applications demonstrating SDK usage
│   ├── hello/            # Simple HTTP handler example with mux routing
│   │   ├── main.go       # Handler registration and endpoint definitions
│   │   ├── go.mod        # Module definition
│   │   ├── wit/          # WIT dependencies for this example
│   │   └── build/        # Build output directory (generated)
│   ├── client/           # HTTP client example
│   │   ├── main.go       # Handler that makes outgoing HTTP requests
│   │   ├── go.mod        # Module definition
│   │   └── wit/          # WIT dependencies for this example
│   └── blobby/           # Blobstore example (basic handler)
│       ├── main.go       # Simple handler
│       ├── go.mod        # Module definition
│       └── wit/          # WIT dependencies for this example
├── gen/                  # Generated WIT bindings (DO NOT EDIT)
│   ├── export_wasi_http_incoming_handler/  # Incoming HTTP handler exports
│   │   ├── handler.go    # Handler dispatch and setup
│   │   ├── request.go    # WASI -> Go request conversion
│   │   └── response.go   # Go -> WASI response conversion
│   ├── wasi_http_*/      # WASI HTTP type definitions (multiple versions)
│   │   └── wit_bindings.go
│   ├── wasi_io_*/        # WASI I/O types (streams, poll, error)
│   │   └── wit_bindings.go
│   ├── wasi_clocks_monotonic_clock/  # Clock subscription for polling
│   │   └── wit_bindings.go
│   ├── wit_exports/      # Component export entry points (generated)
│   │   └── wit_exports.go  # WASM export functions
│   └── wasi_blobstore_*/  # Blobstore types (draft support)
│       └── wit_bindings.go
├── internal/             # Internal helper packages
│   ├── httptypes/        # HTTP type conversions and stream wrappers
│   │   ├── header.go     # Go Header <-> WASI Fields conversion
│   │   ├── incoming_body.go  # IncomingBody stream wrapper (ReadCloser)
│   │   └── outgoing_body.go  # OutgoingBody stream wrapper (WriteCloser)
│   └── pollable/         # Async I/O polling utilities
│       └── pollable.go   # Context-aware pollable waiting
├── net/                  # Networking packages
│   └── wasihttp/         # WASI HTTP handler and transport (public API)
│       ├── http.go       # Handle() and HandleFunc() registration
│       ├── roundtripper.go  # http.RoundTripper implementation
│       ├── request.go    # Go -> WASI request conversion
│       ├── response.go   # WASI -> Go response conversion
│       └── mapper.go     # Type mapping utilities (method, scheme, errors)
├── wit/                  # WebAssembly Interface Types definitions
│   ├── world.wit         # Component world declaration (imports/exports)
│   └── deps/             # Fetched WASI interface dependencies
│       ├── wasi-http-0.2.0/
│       ├── wasi-io-0.2.0/
│       ├── wasi-clocks-0.2.0/
│       ├── wasi-random-0.2.0/
│       └── wasi-cli-0.2.0/
├── .planning/            # GSD planning artifacts
│   └── codebase/         # Codebase analysis documents
├── .github/              # GitHub configuration
│   └── workflows/        # CI/CD workflows
├── go.mod               # Root module definition
├── go.sum               # Dependency versions
├── wkg.lock             # WIT package lock file
├── Makefile             # Build targets (clean, fetch, generate)
└── README.md            # Project overview
```

## Directory Purposes

**examples/:**
- Purpose: Runnable example applications showing SDK usage
- Contains: User code demonstrating handler registration, request handling, HTTP client usage
- Key files:
  - `examples/hello/main.go` - Multi-endpoint HTTP handler with mux routing
  - `examples/client/main.go` - Single endpoint that makes outgoing requests
  - `examples/blobby/main.go` - Basic handler with blobstore support

**gen/:**
- Purpose: Auto-generated bindings between Go and WebAssembly components (created by `componentize-go` tool)
- Contains: WASI type definitions, struct methods, resource handlers
- Key files:
  - `gen/wit_exports/wit_exports.go` - Component export entry point (calls user handlers)
  - `gen/export_wasi_http_incoming_handler/handler.go` - Main incoming handler dispatcher
  - `gen/wasi_http_types/wit_bindings.go` - HTTP type definitions (OutgoingRequest, IncomingResponse, etc.)
- Important: **DO NOT EDIT** - regenerate via `make generate` when WIT changes

**internal/httptypes/:**
- Purpose: Adapter layer converting between Go net/http types and WASI stream types
- Contains: Stream-based body wrappers, header converters
- Key implementations:
  - `incoming_body.go` - Wraps WASI IncomingBody as `io.ReadCloser`
  - `outgoing_body.go` - Wraps WASI OutgoingBody as `io.WriteCloser`
  - `header.go` - Converts `http.Header` to/from WASI Fields

**internal/pollable/:**
- Purpose: Handle waiting for async WASI operations in synchronous Go code
- Contains: Polling loop with context support and host timer integration
- Key file: `pollable.go` - `Await()` function blocks until pollable is ready

**net/wasihttp/:**
- Purpose: Public SDK API - HTTP handler and client transport
- Contains: User-facing handler registration and HTTP client transport
- Key files:
  - `http.go` - `Handle()` and `HandleFunc()` entry points
  - `roundtripper.go` - `http.RoundTripper` implementation for outgoing requests
  - `request.go` - Converts Go `http.Request` to WASI OutgoingRequest
  - `response.go` - Converts WASI IncomingResponse to Go `http.Response`
  - `mapper.go` - Utility functions for type mapping (HTTP methods, schemes, errors)

**wit/:**
- Purpose: WebAssembly component interface definitions
- Contains: WIT world declaration and fetched WASI interface specs
- Key file: `wit/world.wit` - Declares component name, imports (WASI handlers), exports (user handler)

## Key File Locations

**Entry Points:**

**For Incoming Requests (Server):**
- Primary: `gen/wit_exports/wit_exports.go` - WASM export function registered by runtime
- User Setup: `net/wasihttp/http.go` - Where user calls `Handle()` or `HandleFunc()`
- Handler Dispatch: `gen/export_wasi_http_incoming_handler/handler.go` - Routes to user handler

**For Outgoing Requests (Client):**
- User Code: `examples/client/main.go` - Creates `http.Client{Transport: &wasihttp.Transport{}}`
- Transport: `net/wasihttp/roundtripper.go` - Implements `http.RoundTripper.RoundTrip()`

**Configuration:**

- Component Interface: `wit/world.wit` - Defines what's exported/imported
- Module Definition: `go.mod` - Go module name and dependencies
- Build Config: `Makefile` - Build targets and tool invocations

**Core Logic:**

**Request/Response Conversion:**
- Incoming: `gen/export_wasi_http_incoming_handler/request.go` - WASI request to Go request
- Incoming: `gen/export_wasi_http_incoming_handler/response.go` - Go response to WASI response
- Outgoing: `net/wasihttp/request.go` - Go request to WASI request
- Outgoing: `net/wasihttp/response.go` - WASI response to Go response

**Body Streaming:**
- Read: `internal/httptypes/incoming_body.go` - Stream reading with trailer support
- Write: `internal/httptypes/outgoing_body.go` - Stream writing with trailer support
- Async Wait: `internal/pollable/pollable.go` - Polling loop for stream readiness

**Type Mapping:**
- Method Mapping: `net/wasihttp/mapper.go` - HTTP method string <-> WASI Method enum
- Header Mapping: `internal/httptypes/header.go` - Go Header <-> WASI Fields
- Scheme Mapping: `net/wasihttp/request.go` - URL scheme <-> WASI Scheme

**Testing:**

- Example Tests: Located alongside examples (no dedicated test files visible in current state)
- Integration via: Running examples with component runtime

## Naming Conventions

**Files:**

- `*_types.go` - Type definitions and constructors (generated)
- `*.go` - Single-package implementation files
- `wit_bindings.go` - Generated bindings for WIT interfaces
- `wit_exports.go` - Generated WASM export function wrapper

**Directories:**

- `gen/` prefix - Generated code (one subdirectory per WASI interface)
- `internal/` prefix - Internal packages not part of public API
- `net/` prefix - Networking-related packages
- `examples/` prefix - Runnable example applications
- `wit/` prefix - WIT interface definitions

**Packages:**

- `export_wasi_http_incoming_handler` - Handler for incoming requests (from WASI spec)
- `wasi_http_*` - HTTP types (request, response, headers)
- `wasi_io_*` - I/O primitives (streams, poll, error)
- `httptypes` - SDK's HTTP type helpers (not WASI-generated)
- `pollable` - SDK's async polling utilities
- `wasihttp` - SDK's public net/http adapter

## Where to Add New Code

**New HTTP Handler Endpoint:**
- Primary code: Add handler function to `examples/hello/main.go` (or new example)
- Registration: Add route in `mux.HandleFunc()` within `init()`
- Pattern: Standard Go `http.HandlerFunc` signature, use `net/wasihttp.Handle(mux)`

**New Utility Function for Body Processing:**
- Location: Create in `internal/httypes/` if dealing with header/body conversion
- Naming: Follow existing pattern (`MapHttpHeader`, `NewIncomingBodyReader`)
- Testing: Verify via example application usage

**New WASI Interface Support:**
- WIT Definition: Update `wit/world.wit` with new import/export
- Type Bindings: Run `make fetch && make generate` (creates in `gen/`)
- Wrapper Code: Add helpers to `internal/` if needed for Go adaptation
- Example: Add new example in `examples/` demonstrating usage

**New HTTP Client Feature:**
- Transport Extension: Modify `net/wasihttp/roundtripper.go` or related files
- Request Mapping: Update `net/wasihttp/request.go` for new request properties
- Response Parsing: Update `net/wasihttp/response.go` for new response properties
- Testing: Verify with `examples/client/main.go` example

## Special Directories

**gen/**
- Purpose: Auto-generated code from `componentize-go` tool
- Generated: Yes - entire directory regenerated by `make generate`
- Committed: Yes - bindings are part of the repository
- Regenerate When: WIT definitions in `wit/` change
- Command: `make generate`
- Do NOT Edit: All files in this directory are auto-generated and will be overwritten

**wit/deps/**
- Purpose: Cached WASI interface specifications
- Generated: Yes - fetched by `wash wit fetch` (called by `make fetch`)
- Committed: Typically NOT committed (depends on project policy)
- Fetch When: `wit/` references change or first time setup
- Command: `make fetch` (part of `make generate`)

**examples/*/build/**
- Purpose: Build output directory for each example
- Generated: Yes - compiled WebAssembly modules
- Committed: Typically NOT committed
- Generated By: Cargo/Go build tools when building examples

---

*Structure analysis: 2026-03-16*
