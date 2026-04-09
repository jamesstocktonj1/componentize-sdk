package wasihttp

import (
	"fmt"
	"net/http"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
)

// MapHttpHeaders converts an http.Header into a WASI Fields resource.
func MapHttpHeaders(h http.Header) (*httpTypes.Fields, error) {
	output := httpTypes.MakeFields()
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
