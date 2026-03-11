package wasihttp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

func parseHttpRequest(req *http.Request) *types.OutgoingRequest {
	resp := newOutgoingRequest(req.Header)

	resp.SetAuthority(witTypes.Some(req.Host))
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
	// parse trailers
	optTrailer := witTypes.None[*types.Fields]()
	if len(req.Trailer) > 0 {
		trailer := types.MakeFields()
		for k, v := range req.Header {
			for _, vs := range v {
				trailer.Append(k, []uint8(vs))
			}
		}
		optTrailer = witTypes.Some(trailer)
	}

	// parse body
	if req.Body != nil {
		defer req.Body.Close()

		streamRes := body.Write()
		if streamRes.IsErr() {
			return errors.New("failed to fetch outgoing request body stream")
		}
		stream := streamRes.Ok()

		buf := make([]byte, 4096)
		for {
			n, err := req.Body.Read(buf)
			if err != nil && err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			writeRes := stream.BlockingWriteAndFlush(buf[:n])
			if writeRes.IsErr() {
				return fmt.Errorf("failed to write request body - %+v", writeRes.Err())
			}
		}
	}

	// finish outgoing body
	finishRes := types.OutgoingBodyFinish(body, optTrailer)
	if finishRes.IsErr() {
		return mapErrorCode(finishRes.Err())
	}
	return nil
}
