.PHONY: build
build:
	@echo "Building component"
	@componentize-go build --output build/client.wasm

.PHONY: clean
clean:
	@echo "Cleaning files"
	@rm -rf build wit/deps

.PHONY: run
run: build
	@echo "Starting wasmtime"
	@wasmtime serve --addr 0.0.0.0:8000 -Scli build/client.wasm