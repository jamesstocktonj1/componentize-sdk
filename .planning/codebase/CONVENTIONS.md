# Coding Conventions

**Analysis Date:** 2026-03-16

## Naming Patterns

**Files:**
- Lowercase with underscores separating concepts: `incoming_body.go`, `outgoing_body.go`, `roundtripper.go`
- WIT binding files named with pattern: `wit_bindings.go`
- Single-letter package abbreviations for imports (e.g., `types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"`)

**Functions:**
- camelCase for all function names
- Public functions are capitalized: `Handle()`, `HandleFunc()`, `NewIncomingBodyReader()`, `NewOutgoingBodyWriter()`
- Private helper functions are lowercase: `parseHttpRequest()`, `mapMethod()`, `mapErrorCode()`, `newOutgoingRequest()`
- Factory functions use `New` prefix: `NewIncomingBodyReader()`, `NewOutgoingBodyWriter()`
- Mapper/converter functions use `map` prefix: `mapMethod()`, `mapErrorCode()`, `MapHttpHeader()`, `mapUrlScheme()`
- Parser functions use `parse` prefix: `parseHttpRequest()`, `parseFutureResponse()`, `parseIncomingResponse()`, `parseContentLength()`

**Variables:**
- camelCase for local and package-level variables
- Short variable names acceptable where scope is clear: `h` for headers, `p` for bytes slice, `w` for write stream, `r` for read stream, `ctx` for context
- Struct field names are capitalized: `StatusCode`, `Header`, `Body`, `Trailer`, `ContentLength`
- Abbreviations used in internal types: `ctx`, `err`, `res`, `opt` (for Option types), `innerResult`, `innerResponse`
- Receiver names are typically single letter: `t *Transport`, `r *incomingBody`, `w *outgoingBody`

**Types:**
- PascalCase for struct names: `Transport`, `incomingBody`, `outgoingBody`, `limitedBody`
- Unexported structs use lowercase: `incomingBody`, `outgoingBody`, `limitedBody`
- Use interfaces from standard library when applicable: `io.ReadCloser`, `io.WriteCloser`, `http.RoundTripper`
- Generated types prefixed with package alias: `types.OutgoingRequest`, `types.IncomingResponse`, `streams.InputStream`

## Code Style

**Formatting:**
- Standard Go formatting (`gofmt`-compliant)
- 2-space indentation (standard Go default)
- No explicit linting configuration files present, code follows Go conventions

**Line Formatting:**
- Imports organized in groups: standard library first, then external packages
- Long function signatures broken across multiple lines
- Return values from functions typically checked immediately with `if err != nil` pattern

**Linting:**
- No `.eslintrc`, `.golangci.yaml`, or similar files present
- Code appears to follow idiomatic Go conventions without strict linting enforcement

## Import Organization

**Order:**
1. Standard library imports (e.g., `"context"`, `"fmt"`, `"io"`, `"net/http"`)
2. External packages (e.g., `github.com/jamesstocktonj1/componentize-sdk/...`, `go.bytecodealliance.org/...`)

**Path Aliases:**
- `types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"` - WASI HTTP type bindings
- `witTypes "go.bytecodealliance.org/pkg/wit/types"` - WIT generic type utilities (Option, Some, None, etc.)
- `streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_streams"` - Stream handling
- `handler "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_outgoing_handler"` - HTTP handler
- Full module paths without aliases when only imported once

**Example from `request.go`:**
```go
import (
	"io"
	"net/http"
	"net/url"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	"github.com/jamesstocktonj1/componentize-sdk/internal/httptypes"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)
```

## Error Handling

**Patterns:**
- Immediate error checking: `if err != nil { return err }` or return wrapped error
- Named return values not used; explicit returns
- Error wrapping with context: `fmt.Errorf("failed to set header %s - %+v", key, res.Err())`
- Errors for WIT bindings checked via `.IsErr()` method on Result types
- Option/Result types from WIT bindings use `.IsNone()`, `.IsErr()`, `.Ok()`, `.Err()`, `.Some()` methods
- Early returns on error to avoid nesting
- No error recovery/retry patterns observed; errors propagate up

**Option/Result Checking Pattern:**
```go
optResponse := resp.Get()
if optResponse.IsNone() {
	return nil, errors.New("failed to fetch future response - response is empty")
}

innerResult := optResponse.Some()
if innerResult.IsErr() {
	return nil, errors.New("failed to unwrap future response")
}
innerResponse := innerResult.Ok()
```

**Stream Error Checking:**
```go
readRes := r.stream.Read(uint64(len(p)))
if readRes.IsErr() {
	if readRes.Err().Tag() == streams.StreamErrorClosed {
		r.trailerOnce.Do(r.parseTrailers)
		return 0, io.EOF
	}
	return 0, fmt.Errorf("failed to read from input stream - %+v", readRes.Err())
}
```

## Logging

**Framework:** None - no logging library present in codebase

**Patterns:**
- No logging calls observed in source files
- Code relies on error returns for failure signaling
- Generated WIT binding files may use their own logging (not in scope)

## Comments

**When to Comment:**
- Function-level doc comments on exported functions: `// MapHttpHeader converts an http.Header into a WASI Fields resource.`
- Inline comments explain non-obvious logic
- Limited commenting overall; code intent is typically clear from structure

**JSDoc/TSDoc:**
- Not applicable; Go uses simpler comment conventions
- Doc comments follow standard Go convention: start with function name in exported functions
- Example from `header.go`:
```go
// MapHttpHeader converts an http.Header into a WASI Fields resource.
func MapHttpHeader(h http.Header) (*types.Fields, error) {
```

**Multi-line Comments:**
```go
// NewIncomingBodyReader wraps a consumed IncomingBody as an io.ReadCloser.
// The returned http.Header map will be populated with trailers once the body
// has been fully read or closed. The context is used to cancel reads.
func NewIncomingBodyReader(ctx context.Context, body *types.IncomingBody) (io.ReadCloser, http.Header, error) {
```

## Function Design

**Size:**
- Most functions 20-60 lines; longest functions handle complex nested type conversion (e.g., `parseIncomingResponse()` at 85 lines)
- Helper functions kept small and focused (10-30 lines typically)

**Parameters:**
- Keep parameter count low (usually 1-3 parameters)
- Context as first parameter when present: `func NewIncomingBodyReader(ctx context.Context, ...)`
- Receiver for methods: single letter (`t`, `r`, `w`)
- No variadic parameters observed

**Return Values:**
- Two-value returns common: `(value, error)` pattern
- Multiple return values for complex operations: `(io.ReadCloser, http.Header, error)` in `NewIncomingBodyReader()`
- Early returns on error to avoid indentation

**Example Function Structure from `pollable.go`:**
```go
// Await blocks until the pollable is ready or the context is cancelled.
// Returns ctx.Err() if the context is cancelled before the pollable is ready.
// Each iteration yields to the Go scheduler then blocks on a host-level timer
// to allow the WASM host to advance async I/O.
func Await(ctx context.Context, p *poll.Pollable) error {
	for !p.Ready() {
		if err := ctx.Err(); err != nil {
			return err
		}
		runtime.Gosched()

		timer := monotonicclock.SubscribeDuration(monotonicclock.Duration(pollInterval))
		defer timer.Drop()
		timer.Block()
	}
	return nil
}
```

## Module Design

**Exports:**
- Public API functions and types are capitalized and documented
- Private helper functions/types are lowercase
- Consistent export naming: `Handle()`, `HandleFunc()`, `NewIncomingBodyReader()`, `NewOutgoingBodyWriter()`

**Barrel Files:**
- Not used in this codebase; each package directly exports its public interface

**Package Organization:**
- `net/wasihttp/` - Public HTTP handler API (Handle, HandleFunc, Transport)
- `internal/httptypes/` - Internal HTTP type converters (MapHttpHeader, NewIncomingBodyReader, NewOutgoingBodyWriter)
- `internal/pollable/` - Internal polling utilities (Await)
- `gen/wasi_*` - Auto-generated WIT bindings (not hand-written)

---

*Convention analysis: 2026-03-16*
