package export_wasi_http_incoming_handler

import (
	"net/http"

	wittypes "github.com/bytecodealliance/wit-bindgen/wit_types"
	httptypes "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
)

var handler http.HandlerFunc

func SetHttpHandler(h http.HandlerFunc) {
	handler = h
}

func Handle(request *httptypes.IncomingRequest, responseOut *httptypes.ResponseOutparam) {
	req, err := newHttpRequest(request)
	if err != nil {
		Err := httptypes.MakeErrorCodeInternalError(wittypes.Some(err.Error()))
		result := wittypes.Err[*httptypes.OutgoingResponse](Err)
		httptypes.ResponseOutparamSet(responseOut, result)
		return
	}

	if req.Body != nil {
		defer req.Body.Close()
	}

	res := newHttpResponseWriter(responseOut)
	defer res.Close()

	handler(res, req)
}
