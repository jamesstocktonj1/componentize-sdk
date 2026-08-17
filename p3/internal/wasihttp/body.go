package wasihttp

import (
	"context"
	"io"
	"net/http"
	"sync"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// NewBodyWriter wraps a WASI stream writer + trailer future as an io.WriteCloser.
// trailerMap is read at Close time, so callers may populate it after construction.
// Passing nil or an empty map means no trailers will be sent.
func NewBodyWriter(
	stream *witTypes.StreamWriter[uint8],
	trailerWriter *witTypes.FutureWriter[witTypes.Result[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode]],
	trailerMap http.Header,
) io.WriteCloser {
	return &bodyWriter{
		stream:        stream,
		trailerWriter: trailerWriter,
		trailerMap:    trailerMap,
	}
}

type bodyWriter struct {
	stream        *witTypes.StreamWriter[uint8]
	trailerWriter *witTypes.FutureWriter[witTypes.Result[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode]]
	trailerMap    http.Header
	closeOnce     sync.Once
	closeErr      error
}

var _ io.WriteCloser = (*bodyWriter)(nil)

func (w *bodyWriter) Write(b []byte) (int, error) {
	return int(w.stream.WriteAll(b)), nil
}

func (w *bodyWriter) Close() error {
	w.closeOnce.Do(func() {
		if w.stream != nil {
			w.stream.Drop()
		}

		maybeTrailers := witTypes.None[*httpTypes.Fields]()
		if len(w.trailerMap) > 0 {
			fields, err := MapHttpHeaders(w.trailerMap)
			if err != nil {
				w.closeErr = err
				return
			}
			maybeTrailers = witTypes.Some(fields)
		}
		w.trailerWriter.Write(witTypes.Ok[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode](maybeTrailers))
	})
	return w.closeErr
}

// ErrorCodeMapper maps a wasi:http error-code into a Go error. The internal
// body reader uses this to surface trailer-read failures without depending on
// the public net/wasihttp package (which would create an import cycle).
type ErrorCodeMapper func(httpTypes.ErrorCode) error

// NewBodyReader wraps a WASI stream + trailer futures as an io.ReadCloser.
// The provided trailerMap will be populated with trailers once the body is
// fully read. The caller must share the same trailerMap reference with the
// http.Response.Trailer field so trailers are visible after the body is read.
// mapErr is used to translate an error-code from the trailer future into a
// Go error surfaced via Read/Close; if nil, trailer errors are discarded.
//
// ctx is the context the body was produced under (typically the originating
// request's context). Read races the underlying (blocking) stream read
// against ctx.Done, so a slow/stalled response body stops delivering data to
// the caller as soon as the context is cancelled or its deadline expires,
// the same way a cancelled net/http response body does.
func NewBodyReader(
	ctx context.Context,
	stream *witTypes.StreamReader[uint8],
	trailersFut *witTypes.FutureReader[witTypes.Result[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode]],
	fut *witTypes.FutureWriter[witTypes.Result[witTypes.Unit, httpTypes.ErrorCode]],
	trailerMap http.Header,
	mapErr ErrorCodeMapper,
) io.ReadCloser {
	if ctx == nil {
		ctx = context.Background()
	}
	return &bodyReader{
		ctx:         ctx,
		stream:      stream,
		trailersFut: trailersFut,
		fut:         fut,
		trailerMap:  trailerMap,
		mapErr:      mapErr,
	}
}

type bodyReader struct {
	ctx         context.Context
	stream      *witTypes.StreamReader[uint8]
	trailersFut *witTypes.FutureReader[witTypes.Result[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode]]
	fut         *witTypes.FutureWriter[witTypes.Result[witTypes.Unit, httpTypes.ErrorCode]]
	trailerMap  http.Header
	mapErr      ErrorCodeMapper
	headerOnce  sync.Once
	trailerErr  error
	cancelErr   error
}

var _ io.ReadCloser = (*bodyReader)(nil)

func (s *bodyReader) Read(p []byte) (int, error) {
	if s.cancelErr != nil {
		return 0, s.cancelErr
	}

	done := s.ctx.Done()
	if done == nil {
		return s.read(p)
	}

	type result struct {
		n   int
		err error
	}
	ch := make(chan result, 1)
	go func() {
		n, err := s.read(p)
		ch <- result{n, err}
	}()

	select {
	case res := <-ch:
		return res.n, res.err
	case <-done:
		// The read above is still blocked in the background - the WASI
		// bindings offer no way to abort an in-flight stream read - and it
		// now owns the stream's handle, so further Reads must not touch the
		// stream and instead keep returning this error.
		s.cancelErr = s.ctx.Err()
		return 0, s.cancelErr
	}
}

// read performs the actual (blocking) stream read. It must never run
// concurrently with itself on the same bodyReader.
func (s *bodyReader) read(p []byte) (int, error) {
	n := int(s.stream.Read(p))

	if s.stream.WriterDropped() {
		s.headerOnce.Do(s.readTrailers)
		// Prefer surfacing the runtime error over io.EOF so the caller
		// doesn't mistake a TLS/transport failure for a clean end of body.
		err := s.trailerErr
		if err == nil {
			err = io.EOF
		}
		if n > 0 {
			return n, err
		}
		return 0, err
	}

	return n, nil
}

func (s *bodyReader) readTrailers() {
	s.fut.Write(witTypes.Ok[witTypes.Unit, httpTypes.ErrorCode](witTypes.Unit{}))
	result := s.trailersFut.Read()
	if result.IsErr() {
		if s.mapErr != nil {
			s.trailerErr = s.mapErr(result.Err())
		}
		return
	}
	opt := result.Ok()
	if opt.IsSome() {
		fields := opt.Some()
		for _, kv := range fields.CopyAll() {
			s.trailerMap.Add(kv.F0, string(kv.F1))
		}
		fields.Drop()
	}
}

func (s *bodyReader) Close() error {
	// If EOF was never reached, signal completion to the host without reading trailers.
	s.headerOnce.Do(func() {
		s.fut.Write(witTypes.Ok[witTypes.Unit, httpTypes.ErrorCode](witTypes.Unit{}))
		s.trailersFut.Drop()
	})
	s.stream.Drop()
	return s.trailerErr
}
