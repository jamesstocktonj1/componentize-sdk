package httptypes

import (
	"fmt"
	"net/http"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// FieldsSetter is the subset of *types.Fields used by MapHttpHeader.
type FieldsSetter interface {
	Set(name string, value [][]uint8) witTypes.Result[witTypes.Unit, types.HeaderError]
}

// mapHttpHeaderTo populates the given FieldsSetter from an http.Header.
// This is the testable core of MapHttpHeader.
func mapHttpHeaderTo(h http.Header, output FieldsSetter) error {
	for key, vals := range h {
		values := make([][]uint8, len(vals))
		for i, val := range vals {
			values[i] = []byte(val)
		}
		if res := output.Set(key, values); res.IsErr() {
			return fmt.Errorf("failed to set header %s - %+v", key, res.Err())
		}
	}
	return nil
}

// MapHttpHeader converts an http.Header into a WASI Fields resource.
func MapHttpHeader(h http.Header) (*types.Fields, error) {
	output := types.MakeFields()
	if err := mapHttpHeaderTo(h, output); err != nil {
		return nil, err
	}
	return output, nil
}
