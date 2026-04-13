package wasihttp

import (
	"fmt"
	"io"
	"net/http"

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

	opts := witTypes.None[*httpTypes.RequestOptions]()
	res, futureRead := httpTypes.RequestNew(f, someBody, trailerReader, opts)

	if r := res.SetMethod(internalhttp.MapMethodToWasi(req.Method)); r.IsErr() {
		return nil, nil, nil, fmt.Errorf("invalid request method %q", req.Method)
	}
	if r := res.SetScheme(mapUrlScheme(req.URL)); r.IsErr() {
		return nil, nil, nil, fmt.Errorf("invalid request scheme %q", req.URL.Scheme)
	}
	// req.Host may be empty on client requests; fall back to req.URL.Host
	// so the WASI host has a valid authority for TLS SNI.
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	if r := res.SetAuthority(witTypes.Some(host)); r.IsErr() {
		return nil, nil, nil, fmt.Errorf("invalid request authority %q", host)
	}
	if r := res.SetPathWithQuery(witTypes.Some(req.URL.RequestURI())); r.IsErr() {
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
