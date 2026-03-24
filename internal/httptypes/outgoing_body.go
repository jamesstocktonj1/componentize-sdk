package httptypes

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_0_2_0_streams"
	"github.com/jamesstocktonj1/componentize-sdk/internal/pollable"
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
	written := 0
	for written < len(p) {
		checkRes := w.stream.CheckWrite()
		if checkRes.IsErr() {
			if checkRes.Err().Tag() == streams.StreamErrorClosed {
				return written, io.EOF
			}
			return written, fmt.Errorf("failed to check write capacity - %+v", checkRes.Err())
		}
		capacity := checkRes.Ok()
		if capacity == 0 {
			pollable.AwaitAndDrop(w.stream.Subscribe())
			continue
		}
		chunk := p[written:]
		if uint64(len(chunk)) > capacity {
			chunk = chunk[:capacity]
		}
		writeRes := w.stream.Write(chunk)
		if writeRes.IsErr() {
			if writeRes.Err().Tag() == streams.StreamErrorClosed {
				return written, io.EOF
			}
			return written, fmt.Errorf("failed to write to outgoing body stream - %+v", writeRes.Err())
		}
		written += len(chunk)
	}

	// Flush and wait for the flush to complete.
	if flushRes := w.stream.Flush(); flushRes.IsErr() {
		if flushRes.Err().Tag() == streams.StreamErrorClosed {
			return written, io.EOF
		}
		return written, fmt.Errorf("failed to flush outgoing body stream - %+v", flushRes.Err())
	}
	pollable.AwaitAndDrop(w.stream.Subscribe())
	return written, nil
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
