.PHONY: build
build:
	@echo "Building component"
	@componentize-go --world wasi:sockets/imports@0.3.0 --world componentize-sdk-p3 build --output build/socket-server-p3.wasm

.PHONY: clean
clean:
	@echo "Cleaning files"
	@rm -rf build gen/wasi* gen/wit* wit/deps
