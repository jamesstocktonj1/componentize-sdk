package httptypes

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_streams"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// outputStream abstracts *streams.OutputStream for writing outgoing body data.
type outputStream interface {
	BlockingWriteAndFlush(contents []uint8) witTypes.Result[witTypes.Unit, streams.StreamError]
	BlockingFlush() witTypes.Result[witTypes.Unit, streams.StreamError]
	Drop()
}

// outgoingBodyResource abstracts *types.OutgoingBody.
// FinishWithTrailers corresponds to types.OutgoingBodyFinish, converting the
// trailer http.Header to WASI format internally.
type outgoingBodyResource interface {
	FinishWithTrailers(trailer http.Header) error
}

type outgoingBody struct {
	body    outgoingBodyResource
	stream  outputStream
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
	return w.body.FinishWithTrailers(w.trailer)
}
