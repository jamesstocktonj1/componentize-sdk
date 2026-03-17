//go:build wasip1

package httptypes

import (
	"context"
	"fmt"
	"io"
	"net/http"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_streams"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

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

func (w *wasiIncomingBody) FinishAndGetTrailers(trailer http.Header) {
	futureTrailers := types.IncomingBodyFinish(w.body)
	defer futureTrailers.Drop()

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
		trailer.Add(kv.F0, string(kv.F1))
	}
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
