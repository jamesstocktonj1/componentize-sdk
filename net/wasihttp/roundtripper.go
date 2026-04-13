package wasihttp

import (
	"net/http"

	handler "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_outgoing_handler"
	"github.com/jamesstocktonj1/componentize-sdk/internal/pollable"
)

type Transport struct{}

var _ http.RoundTripper = (*Transport)(nil)

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// parse request
	outRequest, err := parseHttpRequest(req)
	if err != nil {
		return nil, err
	}

	// open outgoing request body
	bodyRes := outRequest.Body()
	if bodyRes.IsErr() {
		return nil, ErrOutgoingBodyTaken
	}
	body := bodyRes.Ok()

	// send request
	futureRes := handler.Handle(outRequest, mapRequestOptions())
	if futureRes.IsErr() {
		return nil, mapErrorCode(futureRes.Err())
	}
	futureResp := futureRes.Ok()

	// write request body and trailers
	if err := finishRequestBody(req, body); err != nil {
		return nil, err
	}

	// wait for response
	waitable := futureResp.Subscribe()
	defer waitable.Drop()
	if err := pollable.AwaitContext(req.Context(), waitable); err != nil {
		return nil, err
	}

	// parse response
	resp, err := parseFutureResponse(req.Context(), futureResp)
	if err != nil {
		return nil, err
	}
	resp.Request = req
	return resp, nil
}
