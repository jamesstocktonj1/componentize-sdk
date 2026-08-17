package wasihttp

import (
	"fmt"
	"io"
	"net/http"
	"time"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	internalhttp "github.com/jamesstocktonj1/componentize-sdk/p3/internal/wasihttp"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// parseHttpRequest builds a WASI request and returns a finish function that
// must be run concurrently with Send (via goroutine) to write the body and
// trailers into the stream after the runtime has opened it.
func parseHttpRequest(req *http.Request) (*httpTypes.Request, *witTypes.FutureReader[witTypes.Result[witTypes.Unit, httpTypes.ErrorCode]], func(), error) {
	f, err := internalhttp.MapHttpHeaders(req.Header)
	if err != nil {
		return nil, nil, nil, err
	}

	trailerWriter, trailerReader := httpTypes.MakeFutureResultOptionFieldsErrorCode()
	someBody := witTypes.None[*witTypes.StreamReader[uint8]]()

	var body io.WriteCloser
	if req.Body != nil {
		bodyWriter, bodyReader := httpTypes.MakeStreamU8()
		someBody = witTypes.Some(bodyReader)
		body = internalhttp.NewBodyWriter(bodyWriter, trailerWriter, req.Trailer)
	} else {
		body = internalhttp.NewBodyWriter(nil, trailerWriter, req.Trailer)
	}

	opts := requestOptionsFromContext(req)
	res, futureRead := httpTypes.RequestNew(f, someBody, trailerReader, opts)

	if res.SetMethod(internalhttp.MapMethodToWasi(req.Method)).IsErr() {
		return nil, nil, nil, fmt.Errorf("invalid request method %q", req.Method)
	}
	if res.SetScheme(mapUrlScheme(req.URL)).IsErr() {
		return nil, nil, nil, fmt.Errorf("invalid request scheme %q", req.URL.Scheme)
	}
	// req.Host may be empty on client requests; fall back to req.URL.Host
	// so the WASI host has a valid authority for TLS SNI.
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	if res.SetAuthority(witTypes.Some(host)).IsErr() {
		return nil, nil, nil, fmt.Errorf("invalid request authority %q", host)
	}
	if res.SetPathWithQuery(witTypes.Some(req.URL.RequestURI())).IsErr() {
		return nil, nil, nil, fmt.Errorf("invalid request path %q", req.URL.RequestURI())
	}

	finish := func() {
		if req.Body != nil {
			defer req.Body.Close()
			io.Copy(body, req.Body) //nolint:errcheck
		}
		body.Close() //nolint:errcheck
	}

	return res, futureRead, finish, nil
}

// requestOptionsFromContext derives WASI request options from the request's
// context deadline, if any, so the host enforces the same deadline the
// caller set via context.WithTimeout/WithDeadline instead of only relying on
// RoundTrip abandoning the request client-side once the deadline passes.
func requestOptionsFromContext(req *http.Request) witTypes.Option[*httpTypes.RequestOptions] {
	deadline, ok := req.Context().Deadline()
	if !ok {
		return witTypes.None[*httpTypes.RequestOptions]()
	}

	timeout := time.Until(deadline)
	if timeout < 0 {
		timeout = 0
	}
	duration := witTypes.Some(uint64(timeout))

	options := httpTypes.MakeRequestOptions()
	// Best-effort: a host that doesn't support these options still gets a
	// working request, since RoundTrip also cancels client-side via ctx.Done.
	options.SetConnectTimeout(duration)
	options.SetFirstByteTimeout(duration)
	options.SetBetweenBytesTimeout(duration)

	return witTypes.Some(options)
}
