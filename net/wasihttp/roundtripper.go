package wasihttp

import (
	"errors"
	"net/http"

	handler "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_outgoing_handler"
)

type Transport struct{}

var _ http.RoundTripper = (*Transport)(nil)

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// parse request
	outRequest := parseHttpRequest(req)

	// open outgoing request body
	bodyRes := outRequest.Body()
	if bodyRes.IsErr() {
		return nil, errors.New("failed to fetch response body")
	}
	body := bodyRes.Ok()

	// send request
	futureRes := handler.Handle(outRequest, mapRequestOptions())
	if futureRes.IsErr() {
		return nil, mapErrorCode(futureRes.Err())
	}
	futureResp := futureRes.Ok()

	// write request body and trailers
	err := finishRequestBody(req, body)
	if err != nil {
		return nil, err
	}

	// wait for response (+ should handle context cancel)
	futureResp.Subscribe().Block()

	// parse response
	resp, err := parseFutureResponse(futureResp)
	if err != nil {
		return nil, err
	}
	resp.Request = req
	return resp, nil
}
