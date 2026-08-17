package wasihttp

import (
	"net/http"

	"github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_client"
)

type Transport struct{}

var _ http.RoundTripper = (*Transport)(nil)

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

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

	type sendResult struct {
		resp *wasi_http_client.Response
		err  error
	}

	// Send blocks the calling goroutine until a response (or error) is
	// available, so run it on its own goroutine and race it against the
	// request's context. This lets RoundTrip return as soon as the context
	// is cancelled or its deadline expires, instead of only being able to
	// return once the underlying WASI call completes. The generated
	// bindings offer no way to abort an in-flight send, so on cancellation
	// the goroutine keeps running in the background and its result is
	// discarded.
	done := make(chan sendResult, 1)
	go func() {
		sendRes := wasi_http_client.Send(request)
		if sendRes.IsErr() {
			done <- sendResult{err: mapErrorCode(sendRes.Err())}
			return
		}
		done <- sendResult{resp: sendRes.Ok()}
	}()

	select {
	case res := <-done:
		if res.err != nil {
			return nil, res.err
		}
		return parseHttpResponse(req, res.resp)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
