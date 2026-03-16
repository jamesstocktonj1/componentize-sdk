# Testing Patterns

**Analysis Date:** 2026-03-16

## Test Framework

**Status:** Not detected

**Current state:**
- No `*_test.go` files found in the codebase
- No test framework dependencies in `go.mod`
- No `testing` package imports in hand-written code
- Generated WIT bindings contain no test files

**Testing approach:**
- Testing via integration/examples rather than unit tests
- Example applications in `examples/` directory serve as functional validators
- CI pipeline (`.github/workflows/ci.yaml`) handles build verification

## Test File Organization

**Location:**
- Not applicable - no test files present

**Naming:**
- Not established - no test files to define pattern

**Structure:**
- Not applicable

## Run Commands

```bash
# No dedicated test commands found in Makefile
# Build is verified via:
make clean
make generate
# Examples build as part of CI/CD pipeline
```

## Testing Approach & Coverage

**Integration Testing:**
- Located in `examples/` directory: `examples/hello/`, `examples/client/`, `examples/blobby/`
- Examples demonstrate:
  - Server-side HTTP handling (`hello/main.go` - serves `/hello`, `/echo`, `/greet` routes)
  - Client-side HTTP requests (`client/main.go` - makes outgoing HTTP requests)
  - Blobstore functionality (`blobby/main.go` - currently not examined but available)

**Example: Hello Server (`examples/hello/main.go`):**
```go
func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", hello)
	mux.HandleFunc("/echo", echo)
	mux.HandleFunc("/greet", greeting)

	wasihttp.Handle(mux)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func echo(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func greeting(w http.ResponseWriter, r *http.Request) {
	request := struct {
		Name string `json:"name"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Hello, %s!", request.Name)
}
```

**Example: Client (`examples/client/main.go`):**
```go
var client = http.Client{Transport: &wasihttp.Transport{}}

func init() {
	wasihttp.HandleFunc(handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/hi", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	response := struct {
		Greeting string `json:"greeting"`
		Name     string `json:"name"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Success %d - %s: %s", resp.StatusCode, response.Name, response.Greeting)
}
```

## Manual Verification Points

**What is tested by examples:**
- Server HTTP handler registration (`wasihttp.Handle()`, `wasihttp.HandleFunc()`)
- Standard library HTTP handler patterns (JSON decoding, error handling)
- Client-side HTTP requests via custom transport (`wasihttp.Transport`)
- Request/response headers, body reading/writing
- JSON encoding/decoding

**What should be tested but may not have explicit tests:**
- Error recovery in streams
- Trailer header handling
- Content-Length parsing edge cases
- Concurrent request handling
- Resource cleanup (`.Drop()` calls)

## CI/CD Testing

**Pipeline:** `.github/workflows/ci.yaml`

**Current verification:**
- Code builds successfully (implicitly via go build in examples)
- Examples compile without errors
- No automated test suite execution

## Testing Recommendations (Not Currently Implemented)

**Unit Test Locations:**
- `net/wasihttp/*_test.go` - for Handler, Transport, request/response parsing
- `internal/httptypes/*_test.go` - for header mapping, body reading/writing
- `internal/pollable/*_test.go` - for polling logic

**Key functionality lacking explicit tests:**
- `mapMethod()` - HTTP method mapping completeness
- `parseContentLength()` - edge cases (missing, invalid, negative values)
- `limitedBody` - proper behavior with limited content
- `incomingBody.parseTrailers()` - trailer extraction
- `outgoingBody.Close()` - safe idempotency
- Error cases in stream operations

## Coverage Requirements

**Currently:** Not enforced (no test framework present)

**Observed Code Patterns Needing Testing:**
- All error paths in `net/wasihttp/response.go` - 26 error checks across Option/Result unwrapping
- Stream error handling in `internal/httptypes/incoming_body.go` - 7 different error conditions
- WIT binding integration across all modules - 40+ `.IsErr()`, `.IsNone()` checks

---

*Testing analysis: 2026-03-16*
