package export_wasi_http_incoming_handler

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"

	httptypes "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_streams"
)

func newHttpRequest(request *httptypes.IncomingRequest) (*http.Request, error) {
	method, err := mapMethod(request.Method())
	if err != nil {
		return nil, err
	}

	authority := "localhost"
	if request.Authority().IsSome() {
		authority = request.Authority().Some()
	}

	path := "/"
	if request.PathWithQuery().IsSome() {
		path = request.PathWithQuery().Some()
	}

	body, trailers, err := newRequestBodyTrailer(request)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%s%s", authority, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Trailer = trailers

	headers := request.Headers()
	for _, vals := range headers.Entries() {
		req.Header.Set(vals.F0, string(vals.F1))
	}
	headers.Drop()

	req.Host = authority
	req.URL.Host = authority
	req.RequestURI = path

	return req, nil
}

func mapMethod(m httptypes.Method) (string, error) {
	switch m.Tag() {
	case httptypes.MethodGet:
		return http.MethodGet, nil
	case httptypes.MethodHead:
		return http.MethodHead, nil
	case httptypes.MethodPost:
		return http.MethodPost, nil
	case httptypes.MethodPut:
		return http.MethodPut, nil
	case httptypes.MethodDelete:
		return http.MethodDelete, nil
	case httptypes.MethodConnect:
		return http.MethodConnect, nil
	case httptypes.MethodOptions:
		return http.MethodOptions, nil
	case httptypes.MethodTrace:
		return http.MethodTrace, nil
	case httptypes.MethodPatch:
		return http.MethodPatch, nil
	case httptypes.MethodOther:
		return m.Other(), nil
	default:
		return "", fmt.Errorf("unknown method type - %+v", m)
	}
}

func newRequestBodyTrailer(request *httptypes.IncomingRequest) (io.ReadCloser, http.Header, error) {
	consumeRes := request.Consume()
	if consumeRes.IsErr() {
		return nil, nil, fmt.Errorf("failed to consume incoming request - %+v", consumeRes.Err())
	}
	body := consumeRes.Ok()

	streamRes := body.Stream()
	if streamRes.IsErr() {
		return nil, nil, fmt.Errorf("failed to open request body stream - %+v", streamRes.Err())
	}
	stream := streamRes.Ok()

	trailers := http.Header{}

	return &requestBody{
		body:     body,
		stream:   stream,
		trailers: trailers,
	}, trailers, nil
}

type requestBody struct {
	body   *httptypes.IncomingBody
	stream *streams.InputStream

	trailers    http.Header
	trailerOnce sync.Once
}

var _ io.ReadCloser = (*requestBody)(nil)

func (r *requestBody) Read(p []byte) (n int, err error) {
	pollable := r.stream.Subscribe()
	for !pollable.Ready() {
		runtime.Gosched()
	}
	pollable.Drop()

	readRes := r.stream.Read(uint64(len(p)))
	if readRes.IsErr() {
		if readRes.Err().Tag() == streams.StreamErrorClosed {
			r.trailerOnce.Do(r.parseTrailers)
			return 0, io.EOF
		}
		return 0, fmt.Errorf("failed to read from input stream - %+v", readRes.Err())
	}

	data := readRes.Ok()
	copy(p, data)
	return len(data), nil
}

func (r *requestBody) parseTrailers() {
	r.stream.Drop()
	r.stream = nil

	futureTrailers := httptypes.IncomingBodyFinish(r.body)
	defer futureTrailers.Drop()

	trailersRes := futureTrailers.Get()
	if trailersRes.IsNone() || trailersRes.Some().IsErr() || trailersRes.Some().Ok().IsErr() || trailersRes.Some().Ok().Ok().IsNone() {
		return
	}

	wasiTrailers := trailersRes.Some().Ok().Ok().Some()
	for _, vals := range wasiTrailers.Entries() {
		r.trailers.Set(vals.F0, string(vals.F1))
	}
	wasiTrailers.Drop()
}

func (r *requestBody) Close() error {
	r.trailerOnce.Do(r.parseTrailers)

	if r.stream != nil {
		r.stream.Drop()
	}

	if r.body != nil {
		r.body.Drop()
		r.body = nil
	}
	return nil
}
