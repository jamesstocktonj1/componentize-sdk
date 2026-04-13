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

	// req.Host may be empty on client requests; fall back to req.URL.Host
	// so the WASI host has a valid authority for TLS SNI.
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
