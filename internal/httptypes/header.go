package httptypes

import (
	"fmt"
	"net/http"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
)

// MapHttpHeader converts an http.Header into a WASI Fields resource.
func MapHttpHeader(h http.Header) (*types.Fields, error) {
	output := types.MakeFields()
	for key, vals := range h {
		values := make([][]uint8, len(vals))
		for i, val := range vals {
			values[i] = []byte(val)
		}
		if res := output.Set(key, values); res.IsErr() {
			return nil, fmt.Errorf("failed to set header %s - %+v", key, res.Err())
		}
	}
	return output, nil
}
