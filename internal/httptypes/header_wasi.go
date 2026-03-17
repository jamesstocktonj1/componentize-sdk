//go:build wasip1

package httptypes

import (
	"net/http"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
)

// MapHttpHeader converts an http.Header into a WASI Fields resource.
func MapHttpHeader(h http.Header) (*types.Fields, error) {
	output := types.MakeFields()
	if err := mapHttpHeaderTo(h, output); err != nil {
		return nil, err
	}
	return output, nil
}
