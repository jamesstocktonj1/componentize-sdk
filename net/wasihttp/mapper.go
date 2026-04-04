package wasihttp

import (
	"fmt"
	"net/http"

	witTypes "go.bytecodealliance.org/pkg/wit/types"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
)

func mapMethod(m string) types.Method {
	switch m {
	case http.MethodGet:
		return types.MakeMethodGet()
	case http.MethodHead:
		return types.MakeMethodHead()
	case http.MethodPost:
		return types.MakeMethodPost()
	case http.MethodPut:
		return types.MakeMethodPut()
	case http.MethodDelete:
		return types.MakeMethodDelete()
	case http.MethodConnect:
		return types.MakeMethodConnect()
	case http.MethodOptions:
		return types.MakeMethodOptions()
	case http.MethodTrace:
		return types.MakeMethodTrace()
	case http.MethodPatch:
		return types.MakeMethodPatch()
	default:
		return types.MakeMethodOther(m)
	}
}

func mapRequestOptions() witTypes.Option[*types.RequestOptions] {
	return witTypes.None[*types.RequestOptions]()
}

func mapErrorCode(e types.ErrorCode) error {
	return fmt.Errorf("http error - %+v", e)
}
