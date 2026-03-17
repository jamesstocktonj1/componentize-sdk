package httptypes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
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

// futureTrailersHandle abstracts *types.FutureTrailers.
type futureTrailersHandle interface {
	Get() witTypes.Option[witTypes.Result[witTypes.Result[witTypes.Option[*types.Fields], types.ErrorCode], witTypes.Unit]]
	Drop()
}

// incomingBodyResource abstracts *types.IncomingBody for use by incomingBody.
type incomingBodyResource interface {
	// Finish consumes the body and returns a future for trailers.
	// Corresponds to types.IncomingBodyFinish.
	Finish() futureTrailersHandle
	Drop()
}

// wasiInputStream wraps *streams.InputStream to implement inputStream.
type wasiInputStream struct {
	s *streams.InputStream
}

func (w *wasiInputStream) Subscribe() subscription {
	return w.s.Subscribe()
}

func (w *wasiInputStream) Read(length uint64) witTypes.Result[[]uint8, streams.StreamError] {
	return w.s.Read(length)
}

func (w *wasiInputStream) Drop() {
	w.s.Drop()
}

// wasiIncomingBody wraps *types.IncomingBody to implement incomingBodyResource.
type wasiIncomingBody struct {
	body *types.IncomingBody
}

func (w *wasiIncomingBody) Finish() futureTrailersHandle {
	return types.IncomingBodyFinish(w.body)
}

func (w *wasiIncomingBody) Drop() {
	w.body.Drop()
}

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
		body:    &wasiIncomingBody{body: body},
		stream:  &wasiInputStream{s: streamRes.Ok()},
		trailer: trailer,
	}, trailer, nil
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
	futureTrailers := r.body.Finish()
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
