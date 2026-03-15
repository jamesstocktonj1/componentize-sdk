package export_wasi_http_incoming_handler

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	wasitypes "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	"github.com/jamesstocktonj1/componentize-sdk/internal/httptypes"
	wittypes "go.bytecodealliance.org/pkg/wit/types"
)

func newHttpResponseWriter(response *wasitypes.ResponseOutparam) *responseHandler {
	return &responseHandler{
		outparam:    response,
		httpHeaders: http.Header{},
	}
}

type responseHandler struct {
	outparam *wasitypes.ResponseOutparam
	response *wasitypes.OutgoingResponse

	wasiHeaders *wasitypes.Fields
	httpHeaders http.Header

	writer io.WriteCloser

	headerOnce sync.Once
	headerErr  error

	statusCode int
}

var _ http.ResponseWriter = (*responseHandler)(nil)

func (r *responseHandler) Header() http.Header {
	return r.httpHeaders
}

func (r *responseHandler) Write(b []byte) (int, error) {
	r.headerOnce.Do(r.flush)
	if r.headerErr != nil {
		return 0, r.headerErr
	}
	return r.writer.Write(b)
}

func (r *responseHandler) WriteHeader(statusCode int) {
	r.headerOnce.Do(func() {
		r.statusCode = statusCode
		r.flush()
	})
}

func (r *responseHandler) Close() error {
	r.headerOnce.Do(r.flush)
	if r.writer == nil {
		return nil
	}
	return r.writer.Close()
}

func (r *responseHandler) flush() {
	var err error
	r.wasiHeaders, err = httptypes.MapHttpHeader(r.httpHeaders)
	if err != nil {
		r.headerErr = err
		return
	}

	r.response = wasitypes.MakeOutgoingResponse(r.wasiHeaders)
	r.response.SetStatusCode(uint16(r.statusCode))

	bodyRes := r.response.Body()
	if bodyRes.IsErr() {
		r.headerErr = fmt.Errorf("failed to open response body - %+v", bodyRes.Err())
		return
	}

	r.writer, err = httptypes.NewOutgoingBodyWriter(bodyRes.Ok(), r.httpHeaders)
	if err != nil {
		r.headerErr = err
		return
	}

	result := wittypes.Ok[*wasitypes.OutgoingResponse, wasitypes.ErrorCode](r.response)
	wasitypes.ResponseOutparamSet(r.outparam, result)
}
