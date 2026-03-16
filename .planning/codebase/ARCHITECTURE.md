# Architecture

**Analysis Date:** 2026-03-16

## Pattern Overview

**Overall:** Adapter/Bridge pattern for WebAssembly Component HTTP binding

This is a Go SDK that bridges standard Go `net/http` handlers and clients to WASI (WebAssembly System Interface) HTTP components. The architecture enables Go developers to write HTTP services and clients that run as WebAssembly components without needing to understand the underlying WASI interface definitions.

**Key Characteristics:**
- Two-way HTTP binding: supports both HTTP request handling (incoming) and HTTP client calls (outgoing)
- WASI type mapping layer that converts between Go's standard library and WIT (WebAssembly Interface Types)
- Non-blocking async I/O handling via WASI streams and pollable resources
- Component exports/imports via generated WIT bindings
- Clean separation between user code and generated bindings

## Layers

**User Application Layer:**
- Purpose: User-written HTTP handlers and client code using standard Go net/http
- Location: `examples/` (example applications)
- Contains: Custom HTTP handler functions, request/response logic
- Depends on: `net/wasihttp` package
- Used by: The wasihttp layer for request/response handling

**WASI HTTP Binding Layer:**
- Purpose: Bridges Go's net/http to WASI HTTP interfaces
- Location: `net/wasihttp/`
- Contains: HTTP handler registration, HTTP client transport implementation
- Depends on: Generated WIT bindings in `gen/`, internal helpers in `internal/`
- Used by: User application layer, generated export handlers

**HTTP Type Mapping Layer:**
- Purpose: Converts between Go HTTP types and WASI HTTP types
- Location: `internal/httptypes/` and `net/wasihttp/`
- Contains: Request/response parsing, header mapping, body wrapping
- Depends on: WASI HTTP type definitions in `gen/wasi_http_types/`
- Used by: Both incoming request handler and outgoing response handler

**WASI I/O Abstraction Layer:**
- Purpose: Handles async I/O operations through WASI streams
- Location: `internal/pollable/`, `internal/httptypes/`
- Contains: Pollable waiting mechanism, stream reading/writing
- Depends on: WASI IO interfaces in `gen/wasi_io_*`
- Used by: HTTP type mapping layer for body streaming

**Generated WIT Bindings Layer:**
- Purpose: Raw bindings between Go and WebAssembly component interfaces
- Location: `gen/`
- Contains: Auto-generated struct types, method handlers, resource wrappers
- Depends on: `go.bytecodealliance.org/pkg/wit/types` for Option/Result types
- Used by: All other layers for WASI interface access

**WIT World Definition:**
- Purpose: Declares component's imports and exports
- Location: `wit/world.wit`
- Contains: Component name, imported WASI interfaces, exported handler interface

## Data Flow

**Incoming Request Flow (Server Mode):**

1. WASI runtime invokes the exported `wasi:http/incoming-handler#handle` function
2. Generated `wit_exports.go` entry point dispatches to `gen/export_wasi_http_incoming_handler/handler.go`
3. Handler receives `IncomingRequest` and `ResponseOutparam` (result container)
4. `newHttpRequest()` in `request.go` converts WASI request to Go `http.Request`:
   - Parses method, authority, path via type mappers in `mapper.go`
   - Opens incoming body stream via `NewIncomingBodyReader()` in `internal/httptypes/incoming_body.go`
   - Populates headers, trailers, content-length
5. `newHttpResponseWriter()` in `response.go` creates Go `http.ResponseWriter`:
   - Pre-allocates WASI response headers map
   - Defers body creation until first Write (lazy initialization via `sync.Once`)
6. User handler (`handler(res, req)`) processes request normally using Go net/http API
7. ResponseWriter flushes headers on first Write or WriteHeader:
   - Converts Go headers to WASI Fields via `MapHttpHeader()` in `internal/httptypes/header.go`
   - Creates OutgoingResponse with mapped headers and status code
   - Opens outgoing body stream via `NewOutgoingBodyWriter()` in `internal/httypes/outgoing_body.go`
8. User writes response body via standard `io.Writer` interface
9. On Close, response writer finishes body with trailers (if any)
10. Result set in `ResponseOutparam` via `ResponseOutparamSet()`

**Outgoing Request Flow (Client Mode):**

1. User creates standard Go `http.Client` with `&wasihttp.Transport{}`
2. Client calls `Do(req)` which calls `Transport.RoundTrip(req)`
3. `RoundTrip()` in `roundtripper.go`:
   - Calls `parseHttpRequest()` to convert Go request to WASI OutgoingRequest
   - Maps method, headers, scheme, authority, path via mappers in `mapper.go` and `request.go`
   - Opens outgoing body stream
   - Sends request via `handler.Handle()` (imported WASI interface) with `mapRequestOptions()`
   - Gets back FutureIncomingResponse
4. Writes request body via `finishRequestBody()` (calls `NewOutgoingBodyWriter()`)
5. Waits for response via `pollable.Await()` in `internal/pollable/pollable.go`:
   - Polls FutureResponse via `Subscribe()` on pollable resource
   - Sleeps 10µs between polls to allow host async I/O
   - Respects Go context cancellation
6. Parses incoming response via `parseFutureResponse()` and `parseIncomingResponse()` in `response.go`
7. Converts WASI IncomingResponse to Go `http.Response`:
   - Extracts status code, headers
   - Opens incoming body stream via `NewIncomingBodyReader()`
   - Handles content-length limiting and trailers
8. Returns `http.Response` to user code

**State Management:**

- **Request/Response Lifecycle:** Managed by `sync.Once` primitives to ensure headers are flushed exactly once
- **Stream Resources:** WASI streams are explicit resources that must be dropped; embedded in body wrappers (`incomingBody`, `outgoingBody`)
- **Pollable Waiting:** Context-aware polling with host-level timer integration allows cooperative scheduling
- **Trailers:** Captured lazily after body consumption via `IncomingBodyFinish()` / `OutgoingBodyFinish()` futures

## Key Abstractions

**HTTP Handler Registration:**
- Purpose: Wire user handlers into WASI export interface
- Examples: `net/wasihttp/http.go`
- Pattern: Global handler variable set via `Handle()` or `HandleFunc()`, invoked by generated export handler

**HTTP Transport RoundTripper:**
- Purpose: Implement `http.RoundTripper` interface for WASI outgoing requests
- Examples: `net/wasihttp/roundtripper.go`, `net/wasihttp/request.go`, `net/wasihttp/response.go`
- Pattern: Standard library pattern - users replace default transport with `&wasihttp.Transport{}`

**Body Wrappers:**
- Purpose: Adapt WASI streams to Go `io.ReadCloser` / `io.WriteCloser` interfaces
- Examples:
  - `internal/httptypes/incoming_body.go` - wraps IncomingBody stream as ReadCloser
  - `internal/httptypes/outgoing_body.go` - wraps OutgoingBody stream as WriteCloser
- Pattern: Implement standard Go io interfaces, manage resource lifecycle (Drop/Close)

**Type Mappers:**
- Purpose: Convert between Go standard types and WASI WIT types
- Examples: `net/wasihttp/mapper.go` (method, scheme, error code)
- Pattern: Conversion functions with exhaustive case matching for enum types

**Pollable Awaiter:**
- Purpose: Bridge sync Go code with async WASI operations
- Examples: `internal/pollable/pollable.go`
- Pattern: Spin-wait with periodic host timer subscription to allow scheduler advancement

## Entry Points

**Incoming Handler Export:**
- Location: `gen/wit_exports/wit_exports.go` (generated)
- Triggers: WASI runtime calling exported handler when HTTP request arrives
- Responsibilities:
  1. Dispatch to `export_wasi_http_incoming_handler.Handle()`
  2. Convert WASI types to function arguments
  3. Handle panic recovery and type marshaling

**User Handler Registration (Handler):**
- Location: `net/wasihttp/http.go`
- Triggers: Called during `init()` in user's main.go
- Responsibilities:
  1. Accept standard Go `http.Handler` or `http.HandlerFunc`
  2. Store globally for dispatch by export handler
  3. Enable standard Go mux usage

**User Handler Registration (HandleFunc):**
- Location: `net/wasihttp/http.go`
- Triggers: Called during `init()` in user's main.go (alternative to Handle)
- Responsibilities:
  1. Accept single `http.HandlerFunc`
  2. Store globally for dispatch by export handler
  3. Simpler API for single-endpoint handlers

**HTTP Client Transport:**
- Location: `net/wasihttp/roundtripper.go`
- Triggers: Called via `http.Client.Do(req)` when transport is `&wasihttp.Transport{}`
- Responsibilities:
  1. Convert Go request to WASI types
  2. Invoke WASI outgoing handler
  3. Poll for response and convert back to Go types

## Error Handling

**Strategy:** Propagate errors through Go standard patterns and WASI Result types

**Patterns:**
- **Incoming Handler Errors:** Converted to WASI `ErrorCode` via `MakeErrorCodeInternalError()`, set in ResponseOutparam
- **WASI Operation Errors:** Result types checked with `IsErr()`, converted to Go error via `fmt.Errorf()` with debug info
- **Stream I/O Errors:** Returned through `io.ReadCloser`/`io.WriteCloser` interfaces, checked for `StreamErrorClosed` tag
- **Context Cancellation:** Propagated from `ctx.Err()` through `pollable.Await()`, stops polling and returns error
- **Body Consumption Errors:** Wrapped with context (e.g., "failed to consume incoming request")

## Cross-Cutting Concerns

**Logging:** Not implemented - errors returned as Go errors for caller handling

**Validation:**
- Method validation via exhaustive case matching in `mapMethod()`
- Header key/value validation delegated to WASI Fields.Set()
- Content-Length parsing with bounds checking (`n < 0`)

**Authentication:** Not handled in SDK - delegated to user handler or middleware

**Resource Cleanup:** Explicit resource management via `.Drop()` calls on WASI resources, `defer` statements in body wrappers to ensure cleanup even on error

---

*Architecture analysis: 2026-03-16*
