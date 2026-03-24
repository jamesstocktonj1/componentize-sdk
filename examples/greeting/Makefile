.PHONY: fetch
fetch:
	@echo "Fetching wit dependencies"
	@wash wit fetch

.PHONY: bindings
bindings: fetch
	@echo "Generating wit-bindgen bindings"
	@componentize-go --world component bindings \
		--pkg-name github.com/jamesstocktonj1/componentize-sdk/examples/greeting/gen \
		--output gen \
		--format

.PHONY: build
build: bindings
	@echo "Building component"
	@componentize-go --world component build --output build/greeting.wasm

.PHONY: clean
clean:
	@echo "Cleaning files"
	@rm -rf build wit/deps
