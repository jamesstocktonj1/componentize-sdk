.PHONY: clean
clean:
	@echo "Removing generated files"
	@rm -rf gen/wasi* gen/wit*

.PHONY: fetch
fetch:
	@echo "Fetching wit dependencies"
	@wkg wit fetch

.PHONY: generate
generate: fetch
	@echo "Generating sdk bindings"
	@componentize-go --world sdk bindings \
		--pkg-name github.com/jamesstocktonj1/componentize-sdk/p3/gen \
		--output gen \
		--format