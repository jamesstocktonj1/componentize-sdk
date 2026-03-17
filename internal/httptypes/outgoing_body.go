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

// outputStream abstracts *streams.OutputStream for writing outgoing body data.
type outputStream interface {
	BlockingWriteAndFlush(contents []uint8) witTypes.Result[witTypes.Unit, streams.StreamError]
	BlockingFlush() witTypes.Result[witTypes.Unit, streams.StreamError]
	Drop()
}

// outgoingBodyResource abstracts *types.OutgoingBody.
// Finish corresponds to types.OutgoingBodyFinish.
type outgoingBodyResource interface {
	Finish(trailers witTypes.Option[*types.Fields]) witTypes.Result[witTypes.Unit, types.ErrorCode]
}

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

func (w *wasiOutgoingBody) Finish(trailers witTypes.Option[*types.Fields]) witTypes.Result[witTypes.Unit, types.ErrorCode] {
	return types.OutgoingBodyFinish(w.body, trailers)
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

	optTrailer := witTypes.None[*types.Fields]()
	if len(w.trailer) > 0 {
		wasiTrailer, err := MapHttpHeader(w.trailer)
		if err != nil {
			return err
		}
		optTrailer = witTypes.Some(wasiTrailer)
	}

	finishRes := w.body.Finish(optTrailer)
	if finishRes.IsErr() {
		return fmt.Errorf("failed to finish outgoing body - %+v", finishRes.Err())
	}
	return nil
}
