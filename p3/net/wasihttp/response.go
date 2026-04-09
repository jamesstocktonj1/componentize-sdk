package wasihttp

import (
	"io"
	"net/http"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	internalhttp "github.com/jamesstocktonj1/componentize-sdk/p3/internal/wasihttp"
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
	return internalhttp.NewBodyReader(stream, trailersFut, fut, trailerMap), trailerMap
}
