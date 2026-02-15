package wasihttp

import (
	"net/http"

	"github.com/jamesstocktonj1/componentize-sdk/gen/export_wasi_http_incoming_handler"
	_ "github.com/jamesstocktonj1/componentize-sdk/gen/wit_exports"
)

func Handle(h http.Handler) {
	export_wasi_http_incoming_handler.SetHttpHandler(h.ServeHTTP)
}

func HandleFunc(h http.HandlerFunc) {
	export_wasi_http_incoming_handler.SetHttpHandler(h)
}
