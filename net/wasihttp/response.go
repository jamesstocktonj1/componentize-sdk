package wasihttp

import (
	"errors"
	"io"
	"net/http"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	"github.com/jamesstocktonj1/componentize-sdk/internal/httptypes"
)

func parseFutureResponse(resp *types.FutureIncomingResponse) (*http.Response, error) {
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
	return parseIncomingResponse(innerResponse.Ok())
}

func parseIncomingResponse(resp *types.IncomingResponse) (*http.Response, error) {
	header := http.Header{}
	for _, v := range resp.Headers().Entries() {
		header.Add(v.F0, string(v.F1))
	}

	body, trailer, err := newResponseBody(resp)
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

func newResponseBody(resp *types.IncomingResponse) (io.ReadCloser, http.Header, error) {
	bodyRes := resp.Consume()
	if bodyRes.IsErr() {
		return nil, nil, errors.New("failed to consume incoming response")
	}
	return httptypes.NewIncomingBodyReader(bodyRes.Ok())
}
