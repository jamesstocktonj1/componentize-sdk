.PHONY: fetch
fetch:
	@echo "Fetching wit dependencies"
	@wkg wit fetch

.PHONY: build
build: fetch
	@echo "Building component"
	@componentize-go --world hello build --output build/hello.wasm

.PHONY: clean
clean:
	@echo "Cleaning files"
	@rm -rf build wit/deps