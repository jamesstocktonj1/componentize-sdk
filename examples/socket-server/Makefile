.PHONY: fetch
fetch:
	@echo "Fetching wit dependencies"
	@wash wit fetch

.PHONY: bindings
bindings: fetch
	@echo "Generating wit-bindgen bindings"
	@componentize-go --world server --wit-path ./wit bindings \
		--pkg-name github.com/jamesstocktonj1/componentize-sdk/examples/socket-server/gen \
		--output gen \
		--format

.PHONY: build
build: bindings
	@echo "Building component"
	@componentize-go --world server --wit-path ./wit build --output build/socket-server.wasm

.PHONY: clean
clean:
	@echo "Cleaning files"
	@rm -rf build gen/wasi* gen/wit* wit/deps
