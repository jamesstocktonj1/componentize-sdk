package export_wasi_http_incoming_handler

import (
	"fmt"
	"io"
	"net/http"

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
	return httptypes.NewIncomingBodyReader(consumeRes.Ok())
}
