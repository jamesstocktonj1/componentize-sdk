.PHONY: generate
generate:
	componentize-go --world sdk bindings \
		--pkg-name github.com/jamesstocktonj1/componentize-sdk/gen \
		--output gen \
		--format