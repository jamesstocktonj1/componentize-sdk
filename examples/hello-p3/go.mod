module github.com/jamesstocktonj1/componentize-sdk/examples/hello-p3

go 1.25.5

replace github.com/jamesstocktonj1/componentize-sdk/p3 => ../../p3

replace github.com/jamesstocktonj1/componentize-sdk/p3/exports/http => ../../p3/exports/http

require github.com/jamesstocktonj1/componentize-sdk/p3/exports/http v0.0.0-00010101000000-000000000000

require (
	github.com/jamesstocktonj1/componentize-sdk/p3 v0.0.0-00010101000000-000000000000 // indirect
	go.bytecodealliance.org/pkg v0.2.3 // indirect
)
