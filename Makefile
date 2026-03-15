.PHONY: clean
clean:
	@echo "Removing generated files"
	@rm -rf wit/deps
	@rm -rf gen/wasi*
	@rm -rf gen/wit*

.PHONY: fetch
fetch:
	@echo "Fetching wit dependencies"
	@wash wit fetch

.PHONY: generate
generate: fetch
	@echo "Generating wit-bindgen bindings"
	@componentize-go --world sdk bindings \
		--pkg-name github.com/jamesstocktonj1/componentize-sdk/gen \
		--output gen \
		--format