package httptypes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_streams"
	"github.com/jamesstocktonj1/componentize-sdk/internal/pollable"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// subscription is returned by subscribing to a stream and can report readiness.
type subscription interface {
	Ready() bool
	Drop()
}

// inputStream abstracts *streams.InputStream for reading incoming body data.
type inputStream interface {
	Subscribe() subscription
	Read(length uint64) witTypes.Result[[]uint8, streams.StreamError]
	Drop()
}

// incomingBodyResource abstracts *types.IncomingBody for use by incomingBody.
type incomingBodyResource interface {
	// FinishAndGetTrailers consumes the body, populates trailer with any HTTP
	// trailers, and releases all associated WASM resources.
	FinishAndGetTrailers(trailer http.Header)
	// Drop releases the body resource without consuming it.
	Drop()
}

type incomingBody struct {
	ctx    context.Context
	body   incomingBodyResource
	stream inputStream

	trailer     http.Header
	trailerOnce sync.Once
}

var _ io.ReadCloser = (*incomingBody)(nil)

func (r *incomingBody) Read(p []byte) (int, error) {
	waitable := r.stream.Subscribe()
	defer waitable.Drop()

	if err := pollable.Await(r.ctx, waitable); err != nil {
		return 0, err
	}

	readRes := r.stream.Read(uint64(len(p)))
	if readRes.IsErr() {
		if readRes.Err().Tag() == streams.StreamErrorClosed {
			r.trailerOnce.Do(r.parseTrailers)
			return 0, io.EOF
		}
		return 0, fmt.Errorf("failed to read from input stream - %+v", readRes.Err())
	}

	data := readRes.Ok()
	copy(p, data)
	return len(data), nil
}

func (r *incomingBody) Close() error {
	r.trailerOnce.Do(r.parseTrailers)

	if r.stream != nil {
		r.stream.Drop()
	}

	if r.body != nil {
		r.body.Drop()
		r.body = nil
	}
	return nil
}

func (r *incomingBody) parseTrailers() {
	r.body.FinishAndGetTrailers(r.trailer)
	r.body = nil
}
