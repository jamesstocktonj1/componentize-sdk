package wasihttp

import (
	"net/http"

	"github.com/jamesstocktonj1/componentize-sdk/p3/gen/export_wasi_http_handler"
	_ "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wit_exports"
)

func Handle(h http.Handler) {
	export_wasi_http_handler.SetHttpHandler(h.ServeHTTP)
}

func HandleFunc(h http.HandlerFunc) {
	export_wasi_http_handler.SetHttpHandler(h)
}
