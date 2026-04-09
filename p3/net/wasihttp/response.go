package wasihttp

import (
	"io"
	"net/http"
	"sync"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

func parseHttpResponse(req *http.Request, resp *httpTypes.Response) (*http.Response, error) {
	// Read status and headers before ResponseConsumeBody, which takes the handle.
	statusCode := int(resp.GetStatusCode())
	header := http.Header{}
	for _, v := range resp.GetHeaders().CopyAll() {
		header.Add(v.F0, string(v.F1))
	}

	body, trailer := newResponseBodyTrailer(resp)
	return &http.Response{
		StatusCode: statusCode,
		Header:     header,
		Body:       body,
		Trailer:    trailer,
		Request:    req,
	}, nil
}

func newResponseBodyTrailer(resp *httpTypes.Response) (io.ReadCloser, http.Header) {
	fut, read := httpTypes.MakeFutureResultUnitErrorCode()
	stream, trailersFut := httpTypes.ResponseConsumeBody(resp, read)

	trailerMap := http.Header{}
	return &responseBody{
		stream:      stream,
		trailersFut: trailersFut,
		fut:         fut,
		trailerMap:  trailerMap,
	}, trailerMap
}

type responseBody struct {
	stream      *witTypes.StreamReader[uint8]
	trailersFut *witTypes.FutureReader[witTypes.Result[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode]]
	fut         *witTypes.FutureWriter[witTypes.Result[witTypes.Unit, httpTypes.ErrorCode]]
	trailerMap  http.Header
	headerOnce  sync.Once
}

var _ io.ReadCloser = (*responseBody)(nil)

func (s *responseBody) Read(p []byte) (int, error) {
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

func (s *responseBody) readTrailers() {
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

func (s *responseBody) Close() error {
	// If EOF was never reached, signal completion to the host without reading trailers.
	s.headerOnce.Do(func() {
		s.fut.Write(witTypes.Ok[witTypes.Unit, httpTypes.ErrorCode](witTypes.Unit{}))
		s.trailersFut.Drop()
	})
	s.stream.Drop()
	return nil
}
