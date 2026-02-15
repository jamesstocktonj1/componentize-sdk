.PHONY: clean
clean:
	@echo "Removing generated files"
	@rm -rf gen/wasi*
	@rm -rf gen/wit*

.PHONY: generate
generate:
	@echo "Generating wit-bindgen bindings"
	@componentize-go --world sdk bindings \
		--pkg-name github.com/jamesstocktonj1/componentize-sdk/gen \
		--output gen \
		--format