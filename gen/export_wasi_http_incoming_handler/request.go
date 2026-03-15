package export_wasi_http_incoming_handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	wasitypes "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	"github.com/jamesstocktonj1/componentize-sdk/internal/httptypes"
)

func newHttpRequest(request *wasitypes.IncomingRequest) (*http.Request, error) {
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

	// Parse headers before consuming the body.
	headers := request.Headers()
	httpHeaders := http.Header{}
	for _, vals := range headers.Entries() {
		httpHeaders.Set(vals.F0, string(vals.F1))
	}
	headers.Drop()

	var body io.ReadCloser = http.NoBody
	var trailers http.Header
	contentLength := parseContentLength(httpHeaders)
	if contentLength > 0 || httpHeaders.Get("Transfer-Encoding") != "" {
		rawBody, t, err := newRequestBodyTrailer(request)
		if err != nil {
			return nil, err
		}
		trailers = t
		// Wrap with a limit so that read-to-EOF calls (e.g. io.ReadAll) return
		// cleanly after contentLength bytes. The WASI stream for incoming request
		// bodies in wasmCloud never signals close, so without this limit any
		// blocking read past the body data hangs indefinitely.
		if contentLength > 0 {
			body = &limitedBody{Reader: io.LimitReader(rawBody, contentLength), Closer: rawBody}
		} else {
			body = rawBody
		}
	}

	url := fmt.Sprintf("http://%s%s", authority, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header = httpHeaders
	req.Trailer = trailers
	req.ContentLength = contentLength
	req.Host = authority
	req.URL.Host = authority
	req.RequestURI = path

	return req, nil
}

// limitedBody pairs an io.LimitReader with the underlying body's Close method.
type limitedBody struct {
	io.Reader
	io.Closer
}

// parseContentLength returns the Content-Length value, or 0 if absent/invalid.
func parseContentLength(h http.Header) int64 {
	cl := h.Get("Content-Length")
	if cl == "" || cl == "0" {
		return 0
	}
	n, err := strconv.ParseInt(cl, 10, 64)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func mapMethod(m wasitypes.Method) (string, error) {
	switch m.Tag() {
	case wasitypes.MethodGet:
		return http.MethodGet, nil
	case wasitypes.MethodHead:
		return http.MethodHead, nil
	case wasitypes.MethodPost:
		return http.MethodPost, nil
	case wasitypes.MethodPut:
		return http.MethodPut, nil
	case wasitypes.MethodDelete:
		return http.MethodDelete, nil
	case wasitypes.MethodConnect:
		return http.MethodConnect, nil
	case wasitypes.MethodOptions:
		return http.MethodOptions, nil
	case wasitypes.MethodTrace:
		return http.MethodTrace, nil
	case wasitypes.MethodPatch:
		return http.MethodPatch, nil
	case wasitypes.MethodOther:
		return m.Other(), nil
	default:
		return "", fmt.Errorf("unknown method type - %+v", m)
	}
}

func newRequestBodyTrailer(request *wasitypes.IncomingRequest) (io.ReadCloser, http.Header, error) {
	consumeRes := request.Consume()
	if consumeRes.IsErr() {
		return nil, nil, fmt.Errorf("failed to consume incoming request - %+v", consumeRes.Err())
	}
	return httptypes.NewIncomingBodyReader(context.Background(), consumeRes.Ok())
}
