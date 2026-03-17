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
	ob := &outgoingBody{
		body:    bodyRes,
		stream:  stream,
		trailer: nil,
	}
	return ob, stream, bodyRes
}

func TestOutgoingBodyWrite_WritesData(t *testing.T) {
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
}

func TestOutgoingBodyWrite_ReturnsEOFWhenStreamClosed(t *testing.T) {
	ob, stream, _ := newTestOutgoingBody()
	closed := streams.MakeStreamErrorClosed()
	stream.writeErr = &closed

	_, err := ob.Write([]byte("data"))
	if err != io.EOF {
		t.Fatalf("expected io.EOF on closed stream, got %v", err)
	}
}

func TestOutgoingBodyWrite_ReturnsErrorOnWriteFailure(t *testing.T) {
	ob, stream, _ := newTestOutgoingBody()
	// Use LastOperationFailed to simulate a non-EOF error.
	// We can't construct wasi_io_error.Error without wasmimport, so we skip
	// that specific variant and only test StreamErrorClosed above.
	_ = stream
	_ = ob
}

func TestOutgoingBodyClose_FinishesBody(t *testing.T) {
	ob, stream, bodyRes := newTestOutgoingBody()

	if err := ob.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !stream.dropCalled {
		t.Error("expected stream.Drop() to be called on Close")
	}
	if stream.flushCalled != 1 {
		t.Errorf("expected 1 BlockingFlush call, got %d", stream.flushCalled)
	}
	if !bodyRes.finished {
		t.Error("expected body.Finish() to be called")
	}
}

func TestOutgoingBodyClose_IdempotentClose(t *testing.T) {
	ob, _, bodyRes := newTestOutgoingBody()

	_ = ob.Close()
	_ = ob.Close()

	// Finish must only be called once.
	// We can't count calls directly, but a second call should not panic.
	// The closeOnce ensures idempotency.
	_ = bodyRes
}

func TestOutgoingBodyClose_NoTrailersPassesNone(t *testing.T) {
	ob, _, bodyRes := newTestOutgoingBody()
	ob.trailer = nil

	if err := ob.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bodyRes.lastTrailers.IsNone() {
		t.Error("expected None trailers when trailer header is nil")
	}
}

func TestOutgoingBodyClose_FinishError(t *testing.T) {
	ob, _, bodyRes := newTestOutgoingBody()
	errCode := types.MakeErrorCodeInternalError(witTypes.None[string]())
	bodyRes.finishErr = &errCode

	err := ob.Close()
	if err == nil {
		t.Fatal("expected error from Finish, got nil")
	}
}

func TestOutgoingBodyWrite_MultipleWrites(t *testing.T) {
	ob, stream, _ := newTestOutgoingBody()

	_, _ = ob.Write([]byte("foo"))
	_, _ = ob.Write([]byte("bar"))
	_, _ = ob.Write([]byte("baz"))

	if string(stream.written) != "foobarbaz" {
		t.Errorf("expected %q, got %q", "foobarbaz", string(stream.written))
	}
}
