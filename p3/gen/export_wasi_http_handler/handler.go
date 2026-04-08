package export_wasi_http_handler

import (
	"net/http"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

var handler http.HandlerFunc

func SetHttpHandler(h http.HandlerFunc) {
	handler = h
}

func Handle(request *httpTypes.Request) witTypes.Result[*httpTypes.Response, httpTypes.ErrorCode] {
	req, err := newHttpRequest(request)
	if err != nil {
		Err := httpTypes.MakeErrorCodeInternalError(witTypes.Some(err.Error()))
		return witTypes.Err[*httpTypes.Response](Err)
	}

	if req.Body != nil {
		defer req.Body.Close()
	}

	res := newHttpResponseWriter()

	go func() {
		handler(res, req)
		res.Close()
	}()

	return res.awaitResponse()
}
