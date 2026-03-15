package httptypes

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_streams"
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
	writeRes := w.stream.BlockingWriteAndFlush(p)
	if writeRes.IsErr() {
		if writeRes.Err().Tag() == streams.StreamErrorClosed {
			return 0, io.EOF
		}
		return 0, fmt.Errorf("failed to write to outgoing body stream - %+v", writeRes.Err())
	}
	return len(p), nil
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
	w.stream.BlockingFlush()
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
