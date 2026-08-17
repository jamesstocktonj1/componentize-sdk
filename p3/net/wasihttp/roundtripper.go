package wasihttp

import (
	"net/http"
	"time"

	"github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_client"
	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// Transport is an http.RoundTripper backed by wasi:http/outgoing-handler.
//
// The timeout fields below are sent to the WASI host as request options, so
// the host itself can enforce them (e.g. abort a stalled connection) rather
// than relying solely on client-side cancellation via the request's
// context. A zero value leaves the corresponding timeout unset, so the host
// applies its own default. These are independent of, and not derived from,
// the request's context deadline.
type Transport struct {
	// ConnectTimeout bounds how long the host may spend establishing the
	// underlying connection.
	ConnectTimeout time.Duration
	// FirstByteTimeout bounds how long the host may wait for the first
	// byte of the response after the request has been sent.
	FirstByteTimeout time.Duration
	// BetweenBytesTimeout bounds how long the host may wait between
	// successive chunks of the response body.
	BetweenBytesTimeout time.Duration
}

var _ http.RoundTripper = (*Transport)(nil)

// requestOptions builds WASI request options from the Transport's configured
// timeouts. It returns None if none are set, leaving the host's defaults in
// place.
func (t *Transport) requestOptions() witTypes.Option[*httpTypes.RequestOptions] {
	if t.ConnectTimeout <= 0 && t.FirstByteTimeout <= 0 && t.BetweenBytesTimeout <= 0 {
		return witTypes.None[*httpTypes.RequestOptions]()
	}

	options := httpTypes.MakeRequestOptions()
	if t.ConnectTimeout > 0 {
		options.SetConnectTimeout(witTypes.Some(uint64(t.ConnectTimeout)))
	}
	if t.FirstByteTimeout > 0 {
		options.SetFirstByteTimeout(witTypes.Some(uint64(t.FirstByteTimeout)))
	}
	if t.BetweenBytesTimeout > 0 {
		options.SetBetweenBytesTimeout(witTypes.Some(uint64(t.BetweenBytesTimeout)))
	}
	return witTypes.Some(options)
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// parse request
	request, futureRead, finish, err := parseHttpRequest(req, t.requestOptions())
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
