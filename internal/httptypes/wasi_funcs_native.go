//go:build !wasip1

package httptypes

import (
	"context"
	"io"
	"net/http"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
)

// MapHttpHeader is only functional on wasip1. On native builds it panics.
func MapHttpHeader(_ http.Header) (*types.Fields, error) {
	panic("MapHttpHeader is only available on wasip1")
}

// NewIncomingBodyReader is only functional on wasip1. On native builds it panics.
func NewIncomingBodyReader(_ context.Context, _ *types.IncomingBody) (io.ReadCloser, http.Header, error) {
	panic("NewIncomingBodyReader is only available on wasip1")
}

// NewOutgoingBodyWriter is only functional on wasip1. On native builds it panics.
func NewOutgoingBodyWriter(_ *types.OutgoingBody, _ http.Header) (io.WriteCloser, error) {
	panic("NewOutgoingBodyWriter is only available on wasip1")
}
