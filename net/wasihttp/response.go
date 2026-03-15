package wasihttp

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	"github.com/jamesstocktonj1/componentize-sdk/internal/httptypes"
)

func parseFutureResponse(ctx context.Context, resp *types.FutureIncomingResponse) (*http.Response, error) {
	optResponse := resp.Get()
	if optResponse.IsNone() {
		return nil, errors.New("failed to fetch future response - response is empty")
	}

	innerResult := optResponse.Some()
	if innerResult.IsErr() {
		return nil, errors.New("failed to unwrap future response")
	}
	innerResponse := innerResult.Ok()

	if innerResponse.IsErr() {
		return nil, mapErrorCode(innerResponse.Err())
	}
	return parseIncomingResponse(ctx, innerResponse.Ok())
}

func parseIncomingResponse(ctx context.Context, resp *types.IncomingResponse) (*http.Response, error) {
	header := http.Header{}
	for _, v := range resp.Headers().Entries() {
		header.Add(v.F0, string(v.F1))
	}

	contentLength := parseContentLength(header)
	body, trailer, err := newResponseBody(ctx, resp, contentLength)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode:    int(resp.Status()),
		Status:        http.StatusText(int(resp.Status())),
		Header:        header,
		Body:          body,
		Trailer:       trailer,
		ContentLength: contentLength,
	}, nil
}

func newResponseBody(ctx context.Context, resp *types.IncomingResponse, contentLength int64) (io.ReadCloser, http.Header, error) {
	bodyRes := resp.Consume()
	if bodyRes.IsErr() {
		return nil, nil, errors.New("failed to consume incoming response")
	}
	rawBody, trailer, err := httptypes.NewIncomingBodyReader(ctx, bodyRes.Ok())
	if err != nil {
		return nil, nil, err
	}
	if contentLength > 0 {
		return &limitedBody{Reader: io.LimitReader(rawBody, contentLength), Closer: rawBody}, trailer, nil
	}
	return rawBody, trailer, nil
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
