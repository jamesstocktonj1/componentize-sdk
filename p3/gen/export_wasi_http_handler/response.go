package export_wasi_http_handler

import (
	"errors"
	"io"
	"net/http"
	"sync"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	internalhttp "github.com/jamesstocktonj1/componentize-sdk/p3/internal/wasihttp"
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
	responseChan chan witTypes.Result[*httpTypes.Response, httpTypes.ErrorCode]
	bodyWriter   io.WriteCloser

	httpHeaders http.Header

	flushOnce sync.Once
	statusCode int
}

var _ http.ResponseWriter = (*responseHandler)(nil)

func (r *responseHandler) Header() http.Header {
	return r.httpHeaders
}

func (r *responseHandler) Write(b []byte) (int, error) {
	r.flushOnce.Do(r.flush)
	if r.bodyWriter == nil {
		return 0, errors.New("response body stream is nil")
	}
	return r.bodyWriter.Write(b)
}

func (r *responseHandler) WriteHeader(statusCode int) {
	r.flushOnce.Do(func() {
		r.statusCode = statusCode
		r.flush()
	})
}

func (r *responseHandler) Close() error {
	r.flushOnce.Do(r.flush)
	if r.bodyWriter != nil {
		return r.bodyWriter.Close()
	}
	return nil
}

func (r *responseHandler) flush() {
	wasiHeaders, err := internalhttp.MapHttpHeaders(r.httpHeaders)
	if err != nil {
		errCode := httpTypes.MakeErrorCodeInternalError(witTypes.Some(err.Error()))
		r.responseChan <- witTypes.Err[*httpTypes.Response](errCode)
		return
	}

	bodyWriter, bodyReader := httpTypes.MakeStreamU8()
	trailerWriter, trailerFuture := httpTypes.MakeFutureResultOptionFieldsErrorCode()

	// Reset httpHeaders so callers can populate trailers between flush and Close.
	r.httpHeaders = make(http.Header)
	r.bodyWriter = internalhttp.NewBodyWriter(bodyWriter, trailerWriter, r.httpHeaders)

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

// awaitResponse returns a single result from the response channel.
func (r *responseHandler) awaitResponse() witTypes.Result[*httpTypes.Response, httpTypes.ErrorCode] {
	return <-r.responseChan
}
