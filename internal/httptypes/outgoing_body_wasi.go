//go:build wasip1

package httptypes

import (
	"fmt"
	"io"
	"net/http"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_streams"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// wasiOutputStream wraps *streams.OutputStream to implement outputStream.
type wasiOutputStream struct {
	s *streams.OutputStream
}

func (w *wasiOutputStream) BlockingWriteAndFlush(contents []uint8) witTypes.Result[witTypes.Unit, streams.StreamError] {
	return w.s.BlockingWriteAndFlush(contents)
}

func (w *wasiOutputStream) BlockingFlush() witTypes.Result[witTypes.Unit, streams.StreamError] {
	return w.s.BlockingFlush()
}

func (w *wasiOutputStream) Drop() {
	w.s.Drop()
}

// wasiOutgoingBody wraps *types.OutgoingBody to implement outgoingBodyResource.
type wasiOutgoingBody struct {
	body *types.OutgoingBody
}

func (w *wasiOutgoingBody) FinishWithTrailers(trailer http.Header) error {
	optTrailer := witTypes.None[*types.Fields]()
	if len(trailer) > 0 {
		wasiTrailer, err := MapHttpHeader(trailer)
		if err != nil {
			return err
		}
		optTrailer = witTypes.Some(wasiTrailer)
	}
	finishRes := types.OutgoingBodyFinish(w.body, optTrailer)
	if finishRes.IsErr() {
		return fmt.Errorf("failed to finish outgoing body - %+v", finishRes.Err())
	}
	return nil
}

// NewOutgoingBodyWriter wraps an OutgoingBody as an outgoingBody.
// The trailer map is read at Close time, so callers may populate it after
// construction. Passing nil means no trailers will be sent.
func NewOutgoingBodyWriter(body *types.OutgoingBody, trailer http.Header) (io.WriteCloser, error) {
	streamRes := body.Write()
	if streamRes.IsErr() {
		return nil, fmt.Errorf("failed to open outgoing body stream - %+v", streamRes.Err())
	}

	return &outgoingBody{
		body:    &wasiOutgoingBody{body: body},
		stream:  &wasiOutputStream{s: streamRes.Ok()},
		trailer: trailer,
	}, nil
}
