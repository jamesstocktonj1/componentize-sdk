.PHONY: build
build:
	@echo "Building component"
	@componentize-go build --output build/socket-server.wasm

.PHONY: clean
clean:
	@echo "Cleaning files"
	@rm -rf build gen/wasi* gen/wit* wit/deps
