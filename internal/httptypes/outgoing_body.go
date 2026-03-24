package httptypes

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_0_2_0_streams"
	"github.com/jamesstocktonj1/componentize-sdk/internal/pollable"
	"github.com/jamesstocktonj1/componentize-sdk/internal/stream"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// NewOutgoingBodyWriter wraps an OutgoingBody as an outgoingBody.
// The trailer map is read at Close time, so callers may populate it after
// construction. Passing nil means no trailers will be sent.
func NewOutgoingBodyWriter(body *types.OutgoingBody, trailer http.Header) (io.WriteCloser, error) {
	streamRes := body.Write()
	if streamRes.IsErr() {
		return nil, fmt.Errorf("failed to open outgoing body stream - %+v", streamRes.Err())
	}

	return &outgoingBody{
		body:    body,
		stream:  streamRes.Ok(),
		trailer: trailer,
	}, nil
}

type outgoingBody struct {
	body    *types.OutgoingBody
	stream  *streams.OutputStream
	trailer http.Header

	closeOnce sync.Once
	closeErr  error
}

var _ io.WriteCloser = (*outgoingBody)(nil)

func (w *outgoingBody) Write(p []byte) (int, error) {
	return stream.WriteStream(w.stream, p)
}

// Close flushes, drops the stream, and finishes the body with any trailers
// passed at construction. Safe to call multiple times.
func (w *outgoingBody) Close() error {
	w.closeOnce.Do(func() {
		w.closeErr = w.close()
	})
	return w.closeErr
}

func (w *outgoingBody) close() error {
	w.stream.Flush()
	pollable.AwaitAndDrop(w.stream.Subscribe())
	w.stream.Drop()
	w.stream = nil

	optTrailer := witTypes.None[*types.Fields]()
	if len(w.trailer) > 0 {
		wasiTrailer, err := MapHttpHeader(w.trailer)
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
