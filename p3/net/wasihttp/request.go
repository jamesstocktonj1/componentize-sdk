package wasihttp

import (
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
	// The Host header is carried by SetAuthority on the outgoing request.
	// Forwarding it as a field can conflict with the authority used for TLS
	// SNI and is rejected by some WASI HTTP implementations.
	f.Delete("Host") //nolint:errcheck

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

	res.SetMethod(internalhttp.MapMethodToWasi(req.Method))
	res.SetScheme(mapUrlScheme(req.URL))
	// req.Host may be empty on client requests; per net/http, the transport
	// falls back to req.URL.Host in that case. The authority is also used by
	// the WASI host for TLS SNI, so an empty value causes TLS handshake
	// failures (certificate/hostname mismatch) against HTTPS endpoints.
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	res.SetAuthority(witTypes.Some(host))
	res.SetPathWithQuery(witTypes.Some(req.URL.RequestURI()))

	finish := func() {
		if req.Body != nil {
			defer req.Body.Close()
			io.Copy(body, req.Body) //nolint:errcheck
		}
		body.Close() //nolint:errcheck
	}

	return res, futureRead, finish, nil
}
