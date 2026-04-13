package wasihttp

import (
	"context"
	"errors"
	"io"
	"net/http"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	"github.com/jamesstocktonj1/componentize-sdk/internal/httptypes"
)

func parseFutureResponse(ctx context.Context, resp *types.FutureIncomingResponse) (*http.Response, error) {
	optResponse := resp.Get()
	if optResponse.IsNone() {
		return nil, errors.New("future incoming response not ready")
	}

	// The outer result signals whether the future has already been consumed;
	// only the inner result carries the wasi:http error-code.
	innerResult := optResponse.Some()
	if innerResult.IsErr() {
		return nil, ErrFutureResponseConsumed
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

	body, trailer, err := newResponseBody(ctx, resp)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: int(resp.Status()),
		Status:     http.StatusText(int(resp.Status())),
		Header:     header,
		Body:       body,
		Trailer:    trailer,
	}, nil
}

func newResponseBody(ctx context.Context, resp *types.IncomingResponse) (io.ReadCloser, http.Header, error) {
	bodyRes := resp.Consume()
	if bodyRes.IsErr() {
		return nil, nil, ErrIncomingResponseConsumed
	}
	rawBody, trailer, err := httptypes.NewIncomingBodyReader(ctx, bodyRes.Ok())
	if err != nil {
		return nil, nil, err
	}
	return rawBody, trailer, nil
}
