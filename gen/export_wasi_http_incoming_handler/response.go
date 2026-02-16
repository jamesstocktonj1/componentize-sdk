package export_wasi_http_incoming_handler

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	wittypes "github.com/bytecodealliance/wit-bindgen/wit_types"
	httptypes "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_streams"
)

func newHttpResponseWriter(response *httptypes.ResponseOutparam) *responseHandler {
	return &responseHandler{
		outparam:    response,
		httpHeaders: http.Header{},
	}
}

type responseHandler struct {
	outparam *httptypes.ResponseOutparam
	response *httptypes.OutgoingResponse

	wasiHeaders *httptypes.Fields
	httpHeaders http.Header

	body   *httptypes.OutgoingBody
	stream *streams.OutputStream

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

	writeRes := r.stream.Write(b)
	if writeRes.IsErr() {
		if writeRes.Err().Tag() == streams.StreamErrorClosed {
			return 0, io.EOF
		}
		return 0, fmt.Errorf("failed to write to response body - %+v", writeRes.Err())
	}
	r.stream.BlockingFlush()
	return len(b), nil
}

func (r *responseHandler) WriteHeader(statusCode int) {
	r.headerOnce.Do(func() {
		r.statusCode = statusCode
		r.flush()
	})
}

func (r *responseHandler) Close() error {
	r.headerOnce.Do(r.flush)
	if r.stream == nil {
		return nil
	}

	r.stream.BlockingFlush()
	r.stream.Drop()
	r.stream = nil

	trailers := wittypes.None[*httptypes.Fields]()
	if len(r.httpHeaders) > 0 {
		wasiTrailers, err := mapHttpHeader(r.httpHeaders)
		if err != nil {
			return err
		}
		trailers = wittypes.Some(wasiTrailers)
	}

	res := httptypes.OutgoingBodyFinish(r.body, trailers)
	if res.IsErr() {
		return fmt.Errorf("failed to set trailers - %+v", res.Err())
	}
	return nil
}

func (r *responseHandler) flushHeader() (err error) {
	r.wasiHeaders, err = mapHttpHeader(r.httpHeaders)
	return err
}

func (r *responseHandler) flush() {
	if err := r.flushHeader(); err != nil {
		r.headerErr = err
		return
	}

	r.response = httptypes.MakeOutgoingResponse(r.wasiHeaders)
	r.response.SetStatusCode(uint16(r.statusCode))

	bodyRes := r.response.Body()
	if bodyRes.IsErr() {
		r.headerErr = fmt.Errorf("failed to open response body - %+v", bodyRes.Err())
		return
	}
	r.body = bodyRes.Ok()

	writeRes := r.body.Write()
	if writeRes.IsErr() {
		r.headerErr = fmt.Errorf("failed to open response body stream %+v", writeRes.Err())
		return
	}
	r.stream = writeRes.Ok()

	result := wittypes.Ok[*httptypes.OutgoingResponse, httptypes.ErrorCode](r.response)
	httptypes.ResponseOutparamSet(r.outparam, result)
}

func mapHttpHeader(h http.Header) (*httptypes.Fields, error) {
	output := httptypes.MakeFields()
	for key, vals := range h {
		values := [][]uint8{}
		for _, val := range vals {
			values = append(values, []byte(val))
		}
		if res := output.Set(key, values); res.IsErr() {
			return nil, fmt.Errorf("failed to set header %s - %+v", key, res.Err())
		}
	}
	return output, nil
}
