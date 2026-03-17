package httptypes

import (
	"io"
	"testing"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_streams"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// mockOutputStream implements outputStream for testing.
type mockOutputStream struct {
	written     []byte
	dropCalled  bool
	flushCalled int
	writeErr    *streams.StreamError
}

func (m *mockOutputStream) BlockingWriteAndFlush(contents []uint8) witTypes.Result[witTypes.Unit, streams.StreamError] {
	if m.writeErr != nil {
		return witTypes.Err[witTypes.Unit, streams.StreamError](*m.writeErr)
	}
	m.written = append(m.written, contents...)
	return witTypes.Ok[witTypes.Unit, streams.StreamError](witTypes.Unit{})
}

func (m *mockOutputStream) BlockingFlush() witTypes.Result[witTypes.Unit, streams.StreamError] {
	m.flushCalled++
	return witTypes.Ok[witTypes.Unit, streams.StreamError](witTypes.Unit{})
}

func (m *mockOutputStream) Drop() { m.dropCalled = true }

// mockOutgoingBodyResource implements outgoingBodyResource for testing.
type mockOutgoingBodyResource struct {
	finished     bool
	lastTrailers witTypes.Option[*types.Fields]
	finishErr    *types.ErrorCode
}

func (m *mockOutgoingBodyResource) Finish(trailers witTypes.Option[*types.Fields]) witTypes.Result[witTypes.Unit, types.ErrorCode] {
	m.finished = true
	m.lastTrailers = trailers
	if m.finishErr != nil {
		return witTypes.Err[witTypes.Unit, types.ErrorCode](*m.finishErr)
	}
	return witTypes.Ok[witTypes.Unit, types.ErrorCode](witTypes.Unit{})
}

func newTestOutgoingBody() (*outgoingBody, *mockOutputStream, *mockOutgoingBodyResource) {
	stream := &mockOutputStream{}
	bodyRes := &mockOutgoingBodyResource{}
	return &outgoingBody{body: bodyRes, stream: stream}, stream, bodyRes
}

func TestOutgoingBodyWrite(t *testing.T) {
	t.Run("writes data", func(t *testing.T) {
		ob, stream, _ := newTestOutgoingBody()

		n, err := ob.Write([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != 5 {
			t.Errorf("expected 5 bytes written, got %d", n)
		}
		if string(stream.written) != "hello" {
			t.Errorf("expected %q in stream, got %q", "hello", string(stream.written))
		}
	})

	t.Run("multiple writes accumulate", func(t *testing.T) {
		ob, stream, _ := newTestOutgoingBody()

		for _, s := range []string{"foo", "bar", "baz"} {
			if _, err := ob.Write([]byte(s)); err != nil {
				t.Fatalf("unexpected error writing %q: %v", s, err)
			}
		}
		if string(stream.written) != "foobarbaz" {
			t.Errorf("expected %q, got %q", "foobarbaz", string(stream.written))
		}
	})

	t.Run("returns EOF when stream closed", func(t *testing.T) {
		ob, stream, _ := newTestOutgoingBody()
		closed := streams.MakeStreamErrorClosed()
		stream.writeErr = &closed

		_, err := ob.Write([]byte("data"))
		if err != io.EOF {
			t.Fatalf("expected io.EOF, got %v", err)
		}
	})
}

func TestOutgoingBodyClose(t *testing.T) {
	t.Run("flushes drops stream and finishes body", func(t *testing.T) {
		ob, stream, bodyRes := newTestOutgoingBody()

		if err := ob.Close(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stream.flushCalled != 1 {
			t.Errorf("expected 1 BlockingFlush call, got %d", stream.flushCalled)
		}
		if !stream.dropCalled {
			t.Error("expected stream.Drop() to be called")
		}
		if !bodyRes.finished {
			t.Error("expected body.Finish() to be called")
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		ob, _, _ := newTestOutgoingBody()

		// Both calls must succeed without panic; closeOnce ensures Finish runs once.
		if err := ob.Close(); err != nil {
			t.Fatalf("first Close: %v", err)
		}
		if err := ob.Close(); err != nil {
			t.Fatalf("second Close: %v", err)
		}
	})

	t.Run("nil trailers passes None to Finish", func(t *testing.T) {
		ob, _, bodyRes := newTestOutgoingBody()
		ob.trailer = nil

		if err := ob.Close(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bodyRes.lastTrailers.IsNone() {
			t.Error("expected None trailers when trailer header is nil")
		}
	})

	t.Run("finish error is propagated", func(t *testing.T) {
		ob, _, bodyRes := newTestOutgoingBody()
		errCode := types.MakeErrorCodeInternalError(witTypes.None[string]())
		bodyRes.finishErr = &errCode

		if err := ob.Close(); err == nil {
			t.Fatal("expected error from Finish, got nil")
		}
	})
}
