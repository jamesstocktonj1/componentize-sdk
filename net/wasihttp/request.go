package wasihttp

import (
	"io"
	"net/http"
	"net/url"

	witTypes "go.bytecodealliance.org/pkg/wit/types"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	"github.com/jamesstocktonj1/componentize-sdk/internal/httptypes"
)

func parseHTTPRequest(req *http.Request) *types.OutgoingRequest {
	resp := newOutgoingRequest(req.Header)

	resp.SetAuthority(witTypes.Some(req.Host))
	resp.SetMethod(mapMethod(req.Method))
	resp.SetPathWithQuery(witTypes.Some(req.URL.RequestURI()))
	resp.SetScheme(mapURLScheme(req.URL))

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

func mapURLScheme(u *url.URL) witTypes.Option[types.Scheme] {
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
		if _, copyErr := io.Copy(writer, req.Body); copyErr != nil {
			return copyErr
		}
	}

	return writer.Close()
}
