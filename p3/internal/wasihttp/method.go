package wasihttp

import (
	"fmt"
	"net/http"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
)

// MapMethodToWasi converts a Go HTTP method string into a WASI Method.
func MapMethodToWasi(m string) httpTypes.Method {
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

// MapMethodFromWasi converts a WASI Method into a Go HTTP method string.
func MapMethodFromWasi(m httpTypes.Method) (string, error) {
	switch m.Tag() {
	case httpTypes.MethodGet:
		return http.MethodGet, nil
	case httpTypes.MethodHead:
		return http.MethodHead, nil
	case httpTypes.MethodPost:
		return http.MethodPost, nil
	case httpTypes.MethodPut:
		return http.MethodPut, nil
	case httpTypes.MethodDelete:
		return http.MethodDelete, nil
	case httpTypes.MethodConnect:
		return http.MethodConnect, nil
	case httpTypes.MethodOptions:
		return http.MethodOptions, nil
	case httpTypes.MethodTrace:
		return http.MethodTrace, nil
	case httpTypes.MethodPatch:
		return http.MethodPatch, nil
	case httpTypes.MethodOther:
		return m.Other(), nil
	default:
		return "", fmt.Errorf("unknown method type - %+v", m)
	}
}
