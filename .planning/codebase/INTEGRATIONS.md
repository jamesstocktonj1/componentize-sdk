# External Integrations

**Analysis Date:** 2026-03-16

## APIs & External Services

**WASI HTTP:**
- WASI HTTP v0.2.0 - HTTP request/response handling via WASI standard
  - SDK Package: `go.bytecodealliance.org/pkg` (provides generated bindings)
  - Imports: `wasi:http/outgoing-handler@0.2.0` - Make outgoing HTTP requests
  - Exports: `wasi:http/incoming-handler@0.2.0` - Handle incoming HTTP requests
  - Implementation files:
    - `net/wasihttp/http.go` - Handler registration (SetHttpHandler, SetHandlerFunc)
    - `net/wasihttp/roundtripper.go` - HTTP RoundTripper for client requests
    - `net/wasihttp/request.go` - Parse http.Request to WASI format
    - `net/wasihttp/response.go` - Parse WASI response to http.Response
    - `net/wasihttp/mapper.go` - Type mapping utilities

**WASI Blobstore (Draft):**
- WASI Blobstore v0.2.0-draft - Object storage interface (experimental)
  - SDK Package: `go.bytecodealliance.org/pkg` (provides generated bindings)
  - Imports: `wasi:blobstore/blobstore@0.2.0-draft`
  - Generated bindings: `gen/wasi_blobstore_blobstore`, `gen/wasi_blobstore_container`, `gen/wasi_blobstore_types`
  - Example usage: `examples/blobby/main.go` demonstrates basic blobstore integration

## Data Storage

**Object Storage:**
- WASI Blobstore - Cloud-agnostic object storage interface
  - Access: Imported via WIT interface (no direct environment variables)
  - Client: Generated Go bindings from `wasi:blobstore/blobstore@0.2.0-draft`

**File Storage:**
- WASI I/O Streams (0.2.0, 0.2.1) - Standard stream-based I/O
  - No persistent file storage in current examples

## Streaming & I/O

**WASI I/O Interfaces:**
- `wasi:io/streams@0.2.0` - Stream-based input/output operations
- `wasi:io/poll@0.2.0, 0.2.1` - Async polling and event notification
- `wasi:io/error@0.2.0, 0.2.1` - I/O error handling
- Implementation: `internal/pollable/pollable.go` - Polling abstraction for async I/O

## System & Clock Access

**Time Services:**
- `wasi:clocks/monotonic-clock@0.2.0` - Monotonic clock for reliable timing
  - Used by: `internal/pollable/pollable.go` for polling interval delays
- `wasi:clocks/wall-clock@0.2.0` - Wall clock for current time

**Random:**
- `wasi:random/random@0.2.0` - Random number generation (imported but not actively used in core SDK)

## CLI & Command Line

**CLI Support:**
- `wasi:cli/stdin@0.2.0` - Standard input
- `wasi:cli/stdout@0.2.0` - Standard output
- `wasi:cli/stderr@0.2.0` - Standard error
- Available via WASI proxy world definition

## Generated Bindings

All WASI interfaces are code-generated from WIT definitions:
- Generated to: `gen/` directory
- Package prefix: `github.com/jamesstocktonj1/componentize-sdk/gen`
- Bindings modules:
  - `export_wasi_http_incoming_handler` - Export handlers for incoming HTTP
  - `wasi_http_outgoing_handler` - Import handler for outgoing HTTP
  - `wasi_http_incoming_handler` - Import definitions for incoming requests
  - `wasi_http_types` - HTTP type definitions
  - `wasi_blobstore_blobstore`, `wasi_blobstore_container`, `wasi_blobstore_types` - Blobstore interfaces
  - `wasi_io_poll`, `wasi_io_streams`, `wasi_io_error` - I/O interfaces
  - `wasi_clocks_monotonic_clock` - Clock interface
  - `wit_exports` - Centralized export definitions

## Authentication & Identity

**Auth Provider:**
- None detected - Authentication is handled by the host runtime/orchestration layer
- WASI components run within a confined environment controlled by the host
- No built-in credential management within SDK

## Build & Dependency Management

**External Tools:**
- `componentize-go` (from bytecodealliance/componentize-go:canary) - Compilation to WASM components
- `wash` (wasmCloud utility) - WIT dependency fetching and package management
- GitHub Actions (`.github/workflows/ci.yaml`) - CI/CD pipeline

**Build Steps:**
1. `wash wit fetch` - Downloads WIT dependencies from WASI registry (wasi.dev)
2. `componentize-go --world sdk bindings` - Generates Go bindings from WIT world definition
3. Go compilation to WASM target architecture

## Webhooks & Callbacks

**Incoming HTTP Webhooks:**
- Supported via `wasi:http/incoming-handler` export
- Handler pattern: Set handler function to process incoming requests (`wasihttp.HandleFunc()`)
- Example: `examples/hello/main.go` - Exports HTTP handlers for `/hello`, `/echo`, `/greet`

**Outgoing HTTP Callbacks:**
- Supported via `wasi:http/outgoing-handler` import
- Client pattern: Use `wasihttp.Transport` with standard Go `http.Client`
- Example: `examples/client/main.go` - Makes outgoing HTTP requests to external services

## External Package Registry

**WIT Package Registry:**
- Registry: `wasi.dev` (WASI official packages)
- Managed via: `wkg.lock` file (automatic version pinning)
- Fetched WIT packages stored in: `wit/deps/` directory

**Go Module Registry:**
- GitHub module: `github.com/jamesstocktonj1/componentize-sdk`
- Go Modules dependency: `go.bytecodealliance.org/pkg` from Bytecode Alliance

---

*Integration audit: 2026-03-16*
