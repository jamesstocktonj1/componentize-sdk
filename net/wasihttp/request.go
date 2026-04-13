package wasihttp

import (
	"io"
	"net/http"
	"net/url"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	"github.com/jamesstocktonj1/componentize-sdk/internal/httptypes"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

func parseHttpRequest(req *http.Request) *types.OutgoingRequest {
	resp := newOutgoingRequest(req.Header)

	// req.Host may be empty on client requests; per net/http, the transport
	// falls back to req.URL.Host in that case. The authority is also used by
	// the WASI host for TLS SNI, so an empty value causes TLS handshake
	// failures (certificate/hostname mismatch) against HTTPS endpoints.
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	resp.SetAuthority(witTypes.Some(host))
	resp.SetMethod(mapMethod(req.Method))
	resp.SetPathWithQuery(witTypes.Some(req.URL.RequestURI()))
	resp.SetScheme(mapUrlScheme(req.URL))

	return resp
}

func newOutgoingRequest(h http.Header) *types.OutgoingRequest {
	outHeaders := types.MakeFields()
	for k, v := range h {
		// The Host header is carried by SetAuthority on the outgoing request.
		// Forwarding it as a field can conflict with the authority used for
		// TLS SNI and is rejected by some WASI HTTP implementations.
		if http.CanonicalHeaderKey(k) == "Host" {
			continue
		}
		for _, vs := range v {
			outHeaders.Append(k, []uint8(vs))
		}
	}
	return types.MakeOutgoingRequest(outHeaders)
}

func mapUrlScheme(u *url.URL) witTypes.Option[types.Scheme] {
	switch u.Scheme {
	case "http":
		return witTypes.Some(types.MakeSchemeHttp())
	case "https":
		return witTypes.Some(types.MakeSchemeHttps())
	default:
		return witTypes.Some(types.MakeSchemeOther(u.Scheme))
	}
}

func finishRequestBody(req *http.Request, body *types.OutgoingBody) error {
	writer, err := httptypes.NewOutgoingBodyWriter(body, req.Trailer)
	if err != nil {
		return err
	}

	if req.Body != nil {
		defer req.Body.Close()
		if _, err := io.Copy(writer, req.Body); err != nil {
			return err
		}
	}

	return writer.Close()
}
