package httptypes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_0_2_0_streams"
	"github.com/jamesstocktonj1/componentize-sdk/internal/pollable"
)

// NewIncomingBodyReader wraps a consumed IncomingBody as an io.ReadCloser.
// The returned http.Header map will be populated with trailers once the body
// has been fully read or closed. The context is used to cancel reads.
func NewIncomingBodyReader(ctx context.Context, body *types.IncomingBody) (io.ReadCloser, http.Header, error) {
	streamRes := body.Stream()
	if streamRes.IsErr() {
		return nil, nil, fmt.Errorf("failed to open incoming body stream - %+v", streamRes.Err())
	}

	trailer := http.Header{}
	return &incomingBody{
		ctx:     ctx,
		body:    body,
		stream:  streamRes.Ok(),
		trailer: trailer,
	}, trailer, nil
}

type incomingBody struct {
	ctx    context.Context
	body   *types.IncomingBody
	stream *streams.InputStream

	trailer     http.Header
	trailerOnce sync.Once
}

var _ io.ReadCloser = (*incomingBody)(nil)

func (r *incomingBody) Read(p []byte) (int, error) {
	waitable := r.stream.Subscribe()
	defer waitable.Drop()

	if err := pollable.AwaitContext(r.ctx, waitable); err != nil {
		return 0, err
	}

	readRes := r.stream.Read(uint64(len(p)))
	if readRes.IsErr() {
		if readRes.Err().Tag() == streams.StreamErrorClosed {
			return 0, io.EOF
		}
		return 0, fmt.Errorf("failed to read from input stream - %+v", readRes.Err())
	}

	data := readRes.Ok()
	copy(p, data)
	return len(data), nil
}

func (r *incomingBody) Close() error {
	if r.stream != nil {
		r.stream.Drop()
		r.stream = nil
	}
	r.trailerOnce.Do(r.parseTrailers)

	if r.body != nil {
		r.body.Drop()
		r.body = nil
	}
	return nil
}

func (r *incomingBody) parseTrailers() {
	futureTrailers := types.IncomingBodyFinish(r.body)
	defer futureTrailers.Drop()
	r.body = nil

	trailerRes := futureTrailers.Get()
	if trailerRes.IsNone() {
		return
	}

	outerResult := trailerRes.Some()
	if outerResult.IsErr() {
		return
	}

	innerResult := outerResult.Ok()
	if innerResult.IsErr() {
		return
	}

	optTrailer := innerResult.Ok()
	if optTrailer.IsNone() {
		return
	}

	wasiTrailers := optTrailer.Some()
	defer wasiTrailers.Drop()
	for _, kv := range wasiTrailers.Entries() {
		r.trailer.Add(kv.F0, string(kv.F1))
	}
}
