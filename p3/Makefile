.PHONY: clean
clean:
	@echo "Removing generated files"
	@rm -rf \
		gen/wasi* gen/wit* \
		exports/cli/gen/wit* \
		exports/http/gen/wit*

.PHONY: fetch
fetch:
	@echo "Fetching wit dependencies"
	@wkg wit fetch

# The exports passes share the import bindings in p3/gen (--pkg-name) and only
# keep their own wit_exports shims (--export-pkg-name). Each pass generates a
# full tree, so we generate into a temp dir and copy just wit_exports out.
.PHONY: generate
generate: fetch
	@echo "Generating sdk imports"
	@componentize-go --world sdk-imports bindings \
		--pkg-name github.com/jamesstocktonj1/componentize-sdk/p3/gen \
		--output gen \
		--format

	@echo "Generating cli exports"
	@rm -rf .gen-tmp
	@componentize-go --world export-cli bindings \
		--pkg-name github.com/jamesstocktonj1/componentize-sdk/p3/gen \
		--export-pkg-name github.com/jamesstocktonj1/componentize-sdk/p3/exports/cli/gen \
		--output .gen-tmp \
		--format
	@rm -rf exports/cli/gen/wit_exports
	@cp -r .gen-tmp/wit_exports exports/cli/gen/wit_exports
	@rm -rf .gen-tmp

	@echo "Generating http exports"
	@componentize-go --world export-http bindings \
		--pkg-name github.com/jamesstocktonj1/componentize-sdk/p3/gen \
		--export-pkg-name github.com/jamesstocktonj1/componentize-sdk/p3/exports/http/gen \
		--output .gen-tmp \
		--format
	@rm -rf exports/http/gen/wit_exports
	@cp -r .gen-tmp/wit_exports exports/http/gen/wit_exports
	@rm -rf .gen-tmp
