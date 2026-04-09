package wasihttp

import (
	"fmt"
	"net/http"
	"net/url"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

func mapMethod(m string) httpTypes.Method {
	switch m {
	case http.MethodGet:
		return httpTypes.MakeMethodGet()
	case http.MethodHead:
		return httpTypes.MakeMethodHead()
	case http.MethodPost:
		return httpTypes.MakeMethodPost()
	case http.MethodPut:
		return httpTypes.MakeMethodPut()
	case http.MethodDelete:
		return httpTypes.MakeMethodDelete()
	case http.MethodConnect:
		return httpTypes.MakeMethodConnect()
	case http.MethodOptions:
		return httpTypes.MakeMethodOptions()
	case http.MethodTrace:
		return httpTypes.MakeMethodTrace()
	case http.MethodPatch:
		return httpTypes.MakeMethodPatch()
	default:
		return httpTypes.MakeMethodOther(m)
	}
}

func mapUrlScheme(u *url.URL) witTypes.Option[httpTypes.Scheme] {
	switch u.Scheme {
	case "http":
		return witTypes.Some(httpTypes.MakeSchemeHttp())
	case "https":
		return witTypes.Some(httpTypes.MakeSchemeHttps())
	default:
		return witTypes.Some(httpTypes.MakeSchemeOther(u.Scheme))
	}
}

func mapErrorCode(e httpTypes.ErrorCode) error {
	return fmt.Errorf("http error - %+v", e)
}
