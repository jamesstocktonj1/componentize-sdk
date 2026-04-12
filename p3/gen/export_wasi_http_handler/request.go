package export_wasi_http_handler

import (
	"fmt"
	"io"
	"net/http"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	internalhttp "github.com/jamesstocktonj1/componentize-sdk/p3/internal/wasihttp"
)

func newHttpRequest(request *httpTypes.Request) (*http.Request, error) {
	method, err := internalhttp.MapMethodFromWasi(request.GetMethod())
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
		httpHeaders.Add(vals.F0, string(vals.F1))
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

func newRequestBodyTrailer(request *httpTypes.Request) (io.ReadCloser, http.Header) {
	fut, read := httpTypes.MakeFutureResultUnitErrorCode()
	stream, trailersFut := httpTypes.RequestConsumeBody(request, read)

	trailerMap := http.Header{}
	return internalhttp.NewBodyReader(stream, trailersFut, fut, trailerMap), trailerMap
}
