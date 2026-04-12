package wasihttp

import (
	"net/http"

	"github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_client"
)

type Transport struct{}

var _ http.RoundTripper = (*Transport)(nil)

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// parse request
	request, futureRead, finish, err := parseHttpRequest(req)
	if err != nil {
		return nil, err
	}

	// We don't need the body-consumed notification; drop it to avoid blocking.
	futureRead.Drop()

	// Write the body concurrently: the goroutine streams request body data
	// into the WASI stream while Send blocks waiting for the response.
	go finish()

	// send request
	sendRes := wasi_http_client.Send(request)
	if sendRes.IsErr() {
		return nil, mapErrorCode(sendRes.Err())
	}

	return parseHttpResponse(req, sendRes.Ok())
}
