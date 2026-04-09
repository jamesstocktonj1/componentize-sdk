package wasihttp

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// parseHttpRequest builds a WASI request and returns a finish function that
// must be run concurrently with Send (via goroutine) to write the body and
// trailers into the stream after the runtime has opened it.
func parseHttpRequest(req *http.Request) (*httpTypes.Request, *witTypes.FutureReader[witTypes.Result[witTypes.Unit, httpTypes.ErrorCode]], func(), error) {
	f, err := MapHttpHeaders(req.Header)
	if err != nil {
		return nil, nil, nil, err
	}

	trailerWriter, trailerReader := httpTypes.MakeFutureResultOptionFieldsErrorCode()
	someBody := witTypes.None[*witTypes.StreamReader[uint8]]()

	body := &requestBody{
		trailerWriter: trailerWriter,
		httpTrailers:  req.Trailer,
	}
	if req.Body != nil {
		bodyWriter, bodyReader := httpTypes.MakeStreamU8()
		someBody = witTypes.Some(bodyReader)
		body.stream = bodyWriter
	}

	opts := witTypes.None[*httpTypes.RequestOptions]()
	res, futureRead := httpTypes.RequestNew(f, someBody, trailerReader, opts)

	res.SetMethod(mapMethod(req.Method))
	res.SetScheme(mapUrlScheme(req.URL))
	res.SetAuthority(witTypes.Some(req.Host))
	res.SetPathWithQuery(witTypes.Some(req.URL.RequestURI()))

	finish := func() {
		if req.Body != nil {
			defer req.Body.Close()
			io.Copy(body, req.Body) //nolint:errcheck
		}
		body.Close() //nolint:errcheck
	}

	return res, futureRead, finish, nil
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

type requestBody struct {
	stream        *witTypes.StreamWriter[uint8]
	trailerWriter *witTypes.FutureWriter[witTypes.Result[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode]]

	httpTrailers http.Header

	trailerOnce sync.Once
	trailerErr  error
}

var _ io.WriteCloser = (*requestBody)(nil)

func (r *requestBody) Write(b []byte) (int, error) {
	return int(r.stream.WriteAll(b)), nil
}

func (r *requestBody) Close() error {
	r.trailerOnce.Do(func() {
		if r.stream != nil {
			r.stream.Drop()
		}

		maybeTrailers := witTypes.None[*httpTypes.Fields]()
		if len(r.httpTrailers) > 0 {
			wasiHeaders, err := MapHttpHeaders(r.httpTrailers)
			if err != nil {
				r.trailerErr = err
				return
			}
			maybeTrailers = witTypes.Some(wasiHeaders)
		}
		r.trailerWriter.Write(witTypes.Ok[witTypes.Option[*httpTypes.Fields], httpTypes.ErrorCode](maybeTrailers))
	})
	return r.trailerErr
}
