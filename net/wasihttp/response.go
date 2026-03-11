package wasihttp

import (
	"errors"
	"io"
	"net/http"
	"runtime"
	"sync"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_streams"
)

func parseFutureResponse(resp *types.FutureIncomingResponse) (*http.Response, error) {
	optResponse := resp.Get()
	if optResponse.IsNone() {
		return nil, errors.New("failed to fetch future response - response is empty")
	}

	innerResult := optResponse.Some()
	if innerResult.IsErr() {
		return nil, errors.New("failed to unwrap future response")
	}
	innerResponse := innerResult.Ok()

	if innerResponse.IsErr() {
		return nil, mapErrorCode(innerResponse.Err())
	}
	return parseIncomingResponse(innerResponse.Ok())
}

func parseIncomingResponse(resp *types.IncomingResponse) (*http.Response, error) {
	header := http.Header{}
	for _, v := range resp.Headers().Entries() {
		header.Add(v.F0, string(v.F1))
	}

	body, trailer, err := newResponseBody(resp)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: int(resp.Status()),
		Status:     http.StatusText(int(resp.Status())),
		Header:     header,
		Body:       body,
		Trailer:    trailer,
	}, nil
}

func newResponseBody(resp *types.IncomingResponse) (io.ReadCloser, http.Header, error) {
	bodyRes := resp.Consume()
	if bodyRes.IsErr() {
		return nil, nil, errors.New("failed to consume incoming response")
	}
	body := bodyRes.Ok()

	streamRes := body.Stream()
	if streamRes.IsErr() {
		return nil, nil, errors.New("failed to consume incoming response body")
	}
	stream := streamRes.Ok()

	trailer := http.Header{}
	return &responseBody{
		consumer: resp,
		body:     body,
		stream:   stream,
		trailer:  trailer,
	}, trailer, nil
}

type responseBody struct {
	consumer *types.IncomingResponse
	body     *types.IncomingBody
	stream   *streams.InputStream

	trailerLock sync.Mutex
	trailer     http.Header
	trailerOnce sync.Once
}

var _ io.ReadCloser = (*responseBody)(nil)

func (r *responseBody) Close() error {
	r.trailerOnce.Do(r.parseTrailers)

	if r.stream != nil {
		r.stream.Drop()
	}

	if r.body != nil {
		r.body.Drop()
		r.body = nil
	}
	return nil
}

func (r *responseBody) Read(p []byte) (n int, err error) {
	pollable := r.stream.Subscribe()
	defer pollable.Drop()
	for !pollable.Ready() {
		runtime.Gosched()
		// TODO: handle context herr
	}

	readRes := r.stream.Read(uint64(len(p)))
	if readRes.IsErr() {
		if readRes.Err().Tag() == streams.StreamErrorClosed {
			r.trailerOnce.Do(r.parseTrailers)
			return 0, io.EOF
		}
		return 0, errors.New(readRes.Err().LastOperationFailed().ToDebugString())
	}
	data := readRes.Ok()

	copy(p, data)
	return len(data), nil
}

func (r *responseBody) parseTrailers() {
	r.trailerLock.Lock()
	defer r.trailerLock.Unlock()

	futureTrailer := types.IncomingBodyFinish(r.body)
	defer futureTrailer.Drop()

	r.body = nil

	trailerRes := futureTrailer.Get()
	if trailerRes.IsNone() {
		return
	} else if trailerRes.Some().IsErr() {
		return
	} else if trailerRes.Some().Ok().IsErr() {
		return
	} else if trailerRes.Some().Ok().Ok().IsNone() {
		return
	}

	trailer := trailerRes.Some().Ok().Ok().Some()
	defer trailer.Drop()

	for _, kv := range trailer.Entries() {
		r.trailer.Add(kv.F0, string(kv.F1))
	}
}
