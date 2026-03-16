# Technology Stack

**Analysis Date:** 2026-03-16

## Languages

**Primary:**
- Go 1.25.5 - WebAssembly component development using componentize-go

## Runtime

**Environment:**
- WebAssembly Component Model (WASI)
- WebAssembly System Interface v0.2.0

**Package Manager:**
- Go Modules
- Lockfile: `go.mod`, `go.sum` present

## Frameworks

**Core:**
- `componentize-go` (canary) - Compiles Go code to WASM components with Component Model ABI
- `wit-bindgen` - Generates Go bindings from WIT interface definitions
- `go.bytecodealliance.org/pkg` v0.2.1 - Bytecode Alliance package utilities for WASM components

**Build/Dev:**
- `wash` (v2.0.0-rc.8) - wasmCloud utility for fetching WIT dependencies and component management
- `make` - Build orchestration (Makefile at `/home/james/Programming/Wasm/componentize-sdk/Makefile`)

**Testing:**
- `go test` - Standard Go testing framework (run via `go test ./... -coverprofile=coverage.out`)

## Key Dependencies

**Critical:**
- `go.bytecodealliance.org/pkg` v0.2.1 - Runtime utilities for Bytecode Alliance WASM components
  - Provides bindings for WASI interfaces (HTTP, I/O, clocks, etc.)
  - Used by all generated bindings in `gen/` directory

## Configuration

**Build:**
- `Makefile` - Build targets: `clean`, `fetch`, `generate`
  - `fetch` - Downloads WIT dependencies via `wash wit fetch`
  - `generate` - Invokes `componentize-go` to generate bindings from WIT definitions

**WIT Dependencies:**
- Stored in `wkg.lock` (automatically generated package lock file)
- WIT sources fetched to `wit/deps/` directory
- Generated Go bindings output to `gen/` directory with package prefix `github.com/jamesstocktonj1/componentize-sdk/gen`

## Platform Requirements

**Development:**
- Go 1.25.5 or later (configured via GOTOOLCHAIN env var in CI)
- `componentize-go` - Must be installed from bytecodealliance/componentize-go canary branch
- `wash` - wasmCloud CLI tool v2.0.0-rc.8 or later

**Compilation Target:**
- WebAssembly Component Model binary format (.wasm)
- Targets WASI 0.2.0 (or 0.2.0-draft for experimental interfaces like blobstore)

**Runtime:**
- WebAssembly runtime with WASI support (e.g., wasmCloud host, Wasmtime with WASI support)

---

*Stack analysis: 2026-03-16*
