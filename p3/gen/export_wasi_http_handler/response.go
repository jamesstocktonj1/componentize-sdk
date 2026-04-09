package export_wasi_http_handler

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

func newHttpResponseWriter() *responseHandler {
	return &responseHandler{
		responseChan: make(chan witTypes.Result[*httpTypes.Response, httpTypes.ErrorCode]),
		httpHeaders:  make(http.Header),
		statusCode:   http.StatusOK,
	}
}

type responseHandler struct {
	responseChan  chan witTypes.Result[*httpTypes.Response, httpTypes.ErrorCode]
	responseBody  *witTypes.StreamWriter[uint8]
	trailerWriter *witTypes.FutureWriter[witTypes.Result[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode]]

	httpHeaders http.Header

	headerOnce sync.Once
	statusCode int
}

var _ http.ResponseWriter = (*responseHandler)(nil)

func (r *responseHandler) Header() http.Header {
	return r.httpHeaders
}

func (r *responseHandler) Write(b []byte) (int, error) {
	r.headerOnce.Do(r.flush)
	if r.responseBody == nil {
		return 0, errors.New("response body stream is nil")
	}
	return int(r.responseBody.WriteAll(b)), nil
}

func (r *responseHandler) WriteHeader(statusCode int) {
	r.headerOnce.Do(func() {
		r.statusCode = statusCode
		r.flush()
	})
}

func (r *responseHandler) Close() error {
	r.headerOnce.Do(r.flush)
	if r.responseBody != nil {
		r.responseBody.Drop()
	}

	maybeTrailers := witTypes.None[*httpTypes.Fields]()
	if len(r.httpHeaders) > 0 {
		wasiHeaders, err := MapHttpHeaders(r.httpHeaders)
		if err != nil {
			return err
		}
		maybeTrailers = witTypes.Some(wasiHeaders)
	}
	r.trailerWriter.Write(witTypes.Ok[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode](maybeTrailers))

	return nil
}

func (r *responseHandler) flush() {
	wasiHeaders, err := MapHttpHeaders(r.httpHeaders)
	if err != nil {
		// TODO: handle error
		return
	}

	bodyWriter, bodyReader := httpTypes.MakeStreamU8()
	r.responseBody = bodyWriter

	trailerWriter, trailerFuture := httpTypes.MakeFutureResultOptionFieldsErrorCode()
	r.trailerWriter = trailerWriter
	r.httpHeaders = make(http.Header)

	// create response
	resp, send := httpTypes.ResponseNew(
		wasiHeaders,
		witTypes.Some(bodyReader),
		trailerFuture,
	)
	resp.SetStatusCode(uint16(r.statusCode))
	defer send.Drop()

	// send response payload
	r.responseChan <- witTypes.Ok[*httpTypes.Response, httpTypes.ErrorCode](resp)
}

// awaitResponse returns a single
func (r *responseHandler) awaitResponse() witTypes.Result[*httpTypes.Response, httpTypes.ErrorCode] {
	return <-r.responseChan
}

func MapHttpHeaders(h http.Header) (*httpTypes.Fields, error) {
	output := httpTypes.MakeFields()
	for key, vals := range h {
		values := make([][]uint8, len(vals))
		for i, val := range vals {
			values[i] = []byte(val)
		}
		if res := output.Set(key, values); res.IsErr() {
			return nil, fmt.Errorf("failed to set header %s - %+v", key, res.Err())
		}
	}
	return output, nil
}
