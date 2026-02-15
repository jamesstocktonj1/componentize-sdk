package export_wasi_http_incoming_handler

import (
	"net/http"

	httptypes "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
)

var handler http.HandlerFunc

func SetHttpHandler(h http.HandlerFunc) {
	handler = h
}

func Handle(request *httptypes.IncomingRequest, responseOut *httptypes.ResponseOutparam) {
	// TODO: add handler implementation
}
