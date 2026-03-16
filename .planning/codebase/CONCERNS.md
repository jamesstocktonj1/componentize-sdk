# Codebase Concerns

**Analysis Date:** 2026-03-16

## Tech Debt

**Incomplete Timestamp Type Definition:**
- Issue: Timestamp type in blobstore is a bare `uint64` without proper semantics
- Files: `gen/wasi_blobstore_types/wit_bindings.go` (line 34-37)
- Impact: No way to represent nanosecond precision; unclear if value is seconds/milliseconds/nanoseconds; timestamp comparisons and serialization require manual handling
- Fix approach: Implement proper timestamp struct with explicit epoch representation (seconds + nanoseconds since Unix epoch) and helper functions for conversion; align with WASI standards as the upstream issue is resolved

**Missing Error Handling in Stream Flush:**
- Issue: `BlockingFlush()` call discards error return value during body closure
- Files: `internal/httptypes/outgoing_body.go` (line 62)
- Impact: Flush errors (stream closed, permissions denied, etc.) are silently dropped, making it impossible to distinguish between successful sends and failed ones
- Fix approach: Capture and propagate `BlockingFlush()` error; update `close()` signature to return error if not already doing so; ensure error is logged or bubbled to caller

**Hardcoded Default Authority:**
- Issue: Request handler defaults to "localhost" when authority is missing
- Files: `gen/export_wasi_http_incoming_handler/request.go` (line 20)
- Impact: Requests without Authority header will have incorrect Host/URL, breaking Host-based routing and virtual hosting; misleading error messages
- Fix approach: Use empty authority or raise error when authority cannot be determined; allow caller to provide default; document requirement

**Generated Code Contains Panics:**
- Issue: Generated wit-bindgen code contains `panic("unreachable")` in type switch statements
- Files: `gen/wasi_*/wit_bindings.go` (multiple files: tag mismatch, unreachable cases)
- Impact: If WASI interface evolves or has unknown variants, panics crash entire component; no graceful degradation
- Fix approach: Regenerate bindings with newer wit-bindgen version if available; monitor upstream WASI spec changes; consider wrapping generated code with recovery layer

## Missing Test Coverage

**No Unit or Integration Tests:**
- What's not tested: All core functionality in `net/wasihttp/`, `internal/httptypes/`, and `internal/pollable/`
- Files: `net/wasihttp/roundtripper.go`, `net/wasihttp/request.go`, `net/wasihttp/response.go`, `internal/httptypes/incoming_body.go`, `internal/httptypes/outgoing_body.go`, `internal/pollable/pollable.go`
- Risk: Regressions in critical path (HTTP round-trip, body streaming, polling) go undetected; changes to response parsing or request mapping may break production code silently
- Priority: High - This is a foundational utility library used by all components

**No Tests for Edge Cases:**
- What's not tested: Large body handling, header overflow, trailer parsing, context cancellation during polling, empty responses, error responses, connection timeouts
- Files: `gen/export_wasi_http_incoming_handler/response.go`, `internal/httptypes/incoming_body.go`, `internal/pollable/pollable.go`
- Risk: Production failures when requests exceed buffer sizes, trailers are present, or contexts are cancelled
- Priority: High

**No Transport Tests:**
- What's not tested: The `http.RoundTripper` interface implementation via WASI
- Files: `net/wasihttp/roundtripper.go`, `net/wasihttp/mapper.go`
- Risk: Integration with standard Go http client may break with no way to catch it
- Priority: High

## Fragile Areas

**Incoming Body Reader Context Handling:**
- Files: `internal/httptypes/incoming_body.go`
- Why fragile: Context is captured at reader creation time; if passed context is cancelled before reader construction, `Read()` will fail silently; no validation that context is suitable for streaming I/O
- Safe modification: Add context validation before creating reader; document that context must be valid for entire body lifetime; consider explicit error for cancelled contexts at Read time
- Test coverage: None - no tests for context cancellation during body reads

**Polling Loop in WASM:**
- Files: `internal/pollable/pollable.go`
- Why fragile: Tight polling with 10-microsecond sleep may not give WASM host enough time to process async I/O; if host scheduling is slow, loop spins inefficiently
- Safe modification: Make poll interval configurable; add exponential backoff; profile polling overhead; coordinate with host implementation for callback-based waits
- Test coverage: None

**Response Writer Header Synchronization:**
- Files: `gen/export_wasi_http_incoming_handler/response.go`
- Why fragile: Headers are flushed on first `Write()` or explicit `WriteHeader()`; if neither is called (empty response body), headers may not be sent; sync.Once ensures single flush but provides no visibility into flush errors
- Safe modification: Test empty response bodies explicitly; ensure `Close()` always flushes headers even if no body written; log flush errors; validate that trailers from request are properly passed
- Test coverage: None

**Content-Length Handling with Streaming:**
- Files: `gen/export_wasi_http_incoming_handler/request.go` (lines 41-56)
- Why fragile: Logic uses Content-Length OR Transfer-Encoding to determine if body exists; if both are absent and body is present, body is treated as non-existent; LimitReader silently truncates at content length without validation
- Safe modification: Add explicit tests for Content-Length mismatches; log warnings if declared length doesn't match actual; handle chunked encoding explicitly
- Test coverage: None

**Method Mapping for Custom Methods:**
- Files: `net/wasihttp/mapper.go` (line 32)
- Why fragile: Unknown HTTP methods mapped to `MakeMethodOther(m)` but reverse mapping may not exist; some custom methods may not round-trip correctly
- Safe modification: Test custom method round-trips; ensure mapMethod and reverse mapping are symmetric; log warnings for non-standard methods
- Test coverage: None

## Performance Bottlenecks

**Polling Efficiency in Async Operations:**
- Problem: `Await()` function in `internal/pollable/pollable.go` polls with fixed 10-microsecond intervals, potentially creating excessive scheduler pressure
- Files: `internal/pollable/pollable.go` (line 12)
- Cause: Busy-wait loop with fixed sleep doesn't account for actual WASM host event loop; may wake more frequently than necessary or less frequently than desired
- Improvement path: Profile polling behavior; implement adaptive backoff; coordinate with WASM host on event notification mechanisms; consider callback-based waits if host supports

**Blocking Write Operations:**
- Problem: `BlockingWriteAndFlush()` may block for entire request body, potentially holding resources
- Files: `internal/httptypes/outgoing_body.go` (line 42), `internal/httptypes/outgoing_body.go` (line 62)
- Cause: No streaming or buffering; waits for entire body to flush synchronously
- Improvement path: Implement buffering; allow partial flushes; add write timeout configuration

**Header Allocation in Loops:**
- Problem: Headers are set item-by-item in loops; each call may allocate
- Files: `net/wasihttp/request.go` (lines 26-30), `gen/export_wasi_http_incoming_handler/request.go` (lines 33-35)
- Cause: No pre-allocation; header conversion done inline
- Improvement path: Pre-allocate header capacity; batch header operations; benchmark header overhead

## Security Considerations

**Unchecked Content-Length Values:**
- Risk: LimitReader uses Content-Length header value without validation; malicious large values could cause memory exhaustion during streaming
- Files: `net/wasihttp/response.go` (line 63), `gen/export_wasi_http_incoming_handler/request.go` (line 52)
- Current mitigation: None explicit - relies on underlying io.LimitReader
- Recommendations: Validate Content-Length against maximum configured size; log oversized bodies; enforce per-component limits; reject negative Content-Length values (already done in `parseContentLength()`)

**No Request/Response Size Limits:**
- Risk: Malicious peers can send unbounded payloads causing memory or resource exhaustion
- Files: `internal/httptypes/incoming_body.go`, `internal/httptypes/outgoing_body.go`
- Current mitigation: None - library has no built-in limits
- Recommendations: Add configurable max body size; implement streaming limits; provide hooks for resource tracking

**Trailer Header Injection:**
- Risk: Trailer headers are parsed from stream without validation of header names/values
- Files: `internal/httptypes/incoming_body.go` (lines 80-110)
- Current mitigation: Only trailers explicitly supported by HTTP protocol should be present
- Recommendations: Whitelist allowed trailer headers; validate trailer values; log suspicious trailers; consider rejecting all trailers for strict mode

**Authority Default in Host Header:**
- Risk: Missing authority defaults to "localhost", breaking Host header validation and potential causing request smuggling if handler trusts Host header
- Files: `gen/export_wasi_http_incoming_handler/request.go` (line 20)
- Current mitigation: None
- Recommendations: Require explicit authority or fail; document that handlers must validate Host headers; consider adding strict mode that rejects requests without authority

## Scaling Limits

**Synchronous Request Handling:**
- Current capacity: One request at a time per component instance
- Limit: Multiple concurrent requests will queue in WASM runtime
- Scaling path: Request handling is fundamentally async via WASI but serialized in Go; if Go scheduling becomes bottleneck, implement request pooling or component cloning

**Header Field Limits:**
- Current capacity: HTTP header map can grow unbounded in `MapHttpHeader()`
- Limit: Very large header sets (e.g., >10,000 fields) may exhaust memory
- Scaling path: Implement header count limits; add streaming header processing; reject requests exceeding limit

**Stream Buffer Limits:**
- Current capacity: WASI streams have implementation-specific buffer sizes
- Limit: Large bodies may cause buffer overflows or memory pressure
- Scaling path: Profile WASM host stream implementation; configure buffer sizes; implement adaptive buffering

## Dependencies at Risk

**go.bytecodealliance.org/pkg (wasi bindings):**
- Risk: WASI is evolving; newer WASI versions may break compatibility; bindgen version mismatch could cause panics
- Impact: Component breaks if WASI host API changes; depends on external wit-bindgen tool maintenance
- Migration plan: Pin wit-bindgen version in Makefile; monitor WASI release notes; plan for version migrations with deprecation periods; test against multiple WASI versions in CI

**Generated Code Dependency:**
- Risk: `gen/` directory is machine-generated and not checked for correctness by human review
- Impact: Bugs in wit-bindgen output are invisible until runtime; panics in generated code are not covered by project tests
- Migration plan: Add script to validate generated code structure; run linters on generated output; wrap generated panics with recovery layer; version wit-bindgen strictly

## Known Issues

**Silent Flush Failures in Close:**
- Symptoms: Response body write succeeds but data not actually sent when flush fails during close
- Files: `internal/httptypes/outgoing_body.go` (line 62-63)
- Trigger: Stream becomes unavailable between last write and close (connection drop, WASM host shutdown)
- Workaround: Implement retry logic in handler; monitor WASM host stability; validate responses were actually transmitted

**Context Not Passed to Request Body Reading:**
- Symptoms: Request body reading in incoming handler uses `context.Background()` instead of request context
- Files: `gen/export_wasi_http_incoming_handler/request.go` (line 124)
- Trigger: Any handler that tries to cancel request body reads via context will fail silently
- Workaround: Use explicit timeouts; avoid relying on context for body reads; implement timeout wrapper around body reader

**Trailer Headers Lost on Error:**
- Symptoms: If response body completes but trailer parsing fails, trailers are silently dropped
- Files: `internal/httptypes/incoming_body.go` (lines 85-103)
- Trigger: Malformed trailer headers or stream errors during trailer retrieval
- Workaround: Don't rely on trailers for critical data; implement explicit error handling for trailer dependencies

---

*Concerns audit: 2026-03-16*
