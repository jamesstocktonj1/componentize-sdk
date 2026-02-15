module github.com/jamesstocktonj1/componentize-sdk/examples/hello

go 1.25.5

require github.com/jamesstocktonj1/componentize-sdk v0.0.0-00010101000000-000000000000

require github.com/bytecodealliance/wit-bindgen v0.53.1 // indirect

replace github.com/jamesstocktonj1/componentize-sdk => ../../

replace github.com/bytecodealliance/wit-bindgen => github.com/bytecodealliance/wit-bindgen/crates/go/src/package v0.51.0
