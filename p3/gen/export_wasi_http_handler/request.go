package export_wasi_http_handler

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

func newHttpRequest(request *httpTypes.Request) (*http.Request, error) {
	method, err := mapMethod(request.GetMethod())
	if err != nil {
		return nil, err
	}

	authority := "localhost"
	if request.GetAuthority().IsSome() {
		authority = request.GetAuthority().Some()
	}

	path := "/"
	if request.GetPathWithQuery().IsSome() {
		path = request.GetPathWithQuery().Some()
	}

	// Parse headers before consuming the body.
	headers := request.GetHeaders()
	httpHeaders := http.Header{}
	for _, vals := range headers.CopyAll() {
		httpHeaders.Set(vals.F0, string(vals.F1))
	}
	headers.Drop()

	body, trailers := newRequestBodyTrailer(request)

	url := fmt.Sprintf("http://%s%s", authority, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header = httpHeaders
	req.Trailer = trailers
	req.Host = authority
	req.URL.Host = authority
	req.RequestURI = path

	return req, nil
}

func mapMethod(m httpTypes.Method) (string, error) {
	switch m.Tag() {
	case httpTypes.MethodGet:
		return http.MethodGet, nil
	case httpTypes.MethodHead:
		return http.MethodHead, nil
	case httpTypes.MethodPost:
		return http.MethodPost, nil
	case httpTypes.MethodPut:
		return http.MethodPut, nil
	case httpTypes.MethodDelete:
		return http.MethodDelete, nil
	case httpTypes.MethodConnect:
		return http.MethodConnect, nil
	case httpTypes.MethodOptions:
		return http.MethodOptions, nil
	case httpTypes.MethodTrace:
		return http.MethodTrace, nil
	case httpTypes.MethodPatch:
		return http.MethodPatch, nil
	case httpTypes.MethodOther:
		return m.Other(), nil
	default:
		return "", fmt.Errorf("unknown method type - %+v", m)
	}
}

func newRequestBodyTrailer(request *httpTypes.Request) (io.ReadCloser, http.Header) {
	fut, read := httpTypes.MakeFutureResultUnitErrorCode()

	stream, trailersFut := httpTypes.RequestConsumeBody(request, read)

	trailerMap := http.Header{}
	return &streamWrapper{
		stream:      stream,
		trailersFut: trailersFut,
		fut:         fut,
		trailerMap:  trailerMap,
	}, trailerMap
}

type streamWrapper struct {
	stream      *witTypes.StreamReader[uint8]
	trailersFut *witTypes.FutureReader[witTypes.Result[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode]]
	fut         *witTypes.FutureWriter[witTypes.Result[witTypes.Unit, httpTypes.ErrorCode]]
	trailerMap  http.Header
	headerOnce  sync.Once
}

var _ io.ReadCloser = (*streamWrapper)(nil)

func (s *streamWrapper) Read(p []byte) (int, error) {
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

func (s *streamWrapper) readTrailers() {
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

func (s *streamWrapper) Close() error {
	// If EOF was never reached, signal completion to the host without reading trailers.
	s.headerOnce.Do(func() {
		s.fut.Write(witTypes.Ok[witTypes.Unit, httpTypes.ErrorCode](witTypes.Unit{}))
		s.trailersFut.Drop()
	})
	s.stream.Drop()
	return nil
}
