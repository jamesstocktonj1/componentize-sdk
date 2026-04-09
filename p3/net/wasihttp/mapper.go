package wasihttp

import (
	"fmt"
	"net/url"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

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
