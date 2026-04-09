package wasihttp

import (
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

// NewBodyReader wraps a WASI stream + trailer futures as an io.ReadCloser.
// The provided trailerMap will be populated with trailers once the body is
// fully read. The caller must share the same trailerMap reference with the
// http.Response.Trailer field so trailers are visible after the body is read.
func NewBodyReader(
	stream *witTypes.StreamReader[uint8],
	trailersFut *witTypes.FutureReader[witTypes.Result[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode]],
	fut *witTypes.FutureWriter[witTypes.Result[witTypes.Unit, httpTypes.ErrorCode]],
	trailerMap http.Header,
) io.ReadCloser {
	return &bodyReader{
		stream:      stream,
		trailersFut: trailersFut,
		fut:         fut,
		trailerMap:  trailerMap,
	}
}

type bodyReader struct {
	stream      *witTypes.StreamReader[uint8]
	trailersFut *witTypes.FutureReader[witTypes.Result[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode]]
	fut         *witTypes.FutureWriter[witTypes.Result[witTypes.Unit, httpTypes.ErrorCode]]
	trailerMap  http.Header
	headerOnce  sync.Once
}

var _ io.ReadCloser = (*bodyReader)(nil)

func (s *bodyReader) Read(p []byte) (int, error) {
	n := int(s.stream.Read(p))

	if s.stream.WriterDropped() {
		s.headerOnce.Do(s.readTrailers)
		if n > 0 {
			return n, io.EOF
		}
		return 0, io.EOF
	}

	return n, nil
}

func (s *bodyReader) readTrailers() {
	s.fut.Write(witTypes.Ok[witTypes.Unit, httpTypes.ErrorCode](witTypes.Unit{}))
	result := s.trailersFut.Read()
	if result.IsOk() {
		opt := result.Ok()
		if opt.IsSome() {
			fields := opt.Some()
			for _, kv := range fields.CopyAll() {
				s.trailerMap.Add(kv.F0, string(kv.F1))
			}
			fields.Drop()
		}
	}
}

func (s *bodyReader) Close() error {
	// If EOF was never reached, signal completion to the host without reading trailers.
	s.headerOnce.Do(func() {
		s.fut.Write(witTypes.Ok[witTypes.Unit, httpTypes.ErrorCode](witTypes.Unit{}))
		s.trailersFut.Drop()
	})
	s.stream.Drop()
	return nil
}
