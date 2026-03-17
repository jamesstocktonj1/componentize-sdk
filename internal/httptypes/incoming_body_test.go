package httptypes

import (
	"context"
	"io"
	"testing"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_streams"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// alwaysReadySubscription implements subscription and is immediately ready.
type alwaysReadySubscription struct{}

func (s *alwaysReadySubscription) Ready() bool { return true }
func (s *alwaysReadySubscription) Drop()       {}

// neverReadySubscription implements subscription and is never ready.
type neverReadySubscription struct{}

func (s *neverReadySubscription) Ready() bool { return false }
func (s *neverReadySubscription) Drop()       {}

// mockInputStream implements inputStream for testing.
type mockInputStream struct {
	data          []byte
	readPos       int
	dropCalled    bool
	subscribeFunc func() subscription
}

func (m *mockInputStream) Subscribe() subscription {
	if m.subscribeFunc != nil {
		return m.subscribeFunc()
	}
	return &alwaysReadySubscription{}
}

func (m *mockInputStream) Read(length uint64) witTypes.Result[[]uint8, streams.StreamError] {
	if m.readPos >= len(m.data) {
		return witTypes.Err[[]uint8, streams.StreamError](streams.MakeStreamErrorClosed())
	}
	end := m.readPos + int(length)
	if end > len(m.data) {
		end = len(m.data)
	}
	chunk := make([]byte, end-m.readPos)
	copy(chunk, m.data[m.readPos:end])
	m.readPos = end
	return witTypes.Ok[[]uint8, streams.StreamError](chunk)
}

func (m *mockInputStream) Drop() { m.dropCalled = true }

// mockFutureTrailers implements futureTrailersHandle returning no trailers.
type mockFutureTrailers struct {
	dropCalled bool
}

func (m *mockFutureTrailers) Get() witTypes.Option[witTypes.Result[witTypes.Result[witTypes.Option[*types.Fields], types.ErrorCode], witTypes.Unit]] {
	return witTypes.None[witTypes.Result[witTypes.Result[witTypes.Option[*types.Fields], types.ErrorCode], witTypes.Unit]]()
}

func (m *mockFutureTrailers) Drop() { m.dropCalled = true }

// mockIncomingBodyResource implements incomingBodyResource for testing.
type mockIncomingBodyResource struct {
	future     *mockFutureTrailers
	dropCalled bool
}

func (m *mockIncomingBodyResource) Finish() futureTrailersHandle { return m.future }
func (m *mockIncomingBodyResource) Drop()                        { m.dropCalled = true }

func newTestIncomingBody(ctx context.Context, data []byte) (*incomingBody, *mockInputStream, *mockIncomingBodyResource) {
	stream := &mockInputStream{data: data}
	future := &mockFutureTrailers{}
	bodyRes := &mockIncomingBodyResource{future: future}

	ib := &incomingBody{
		ctx:     ctx,
		body:    bodyRes,
		stream:  stream,
		trailer: make(map[string][]string),
	}
	return ib, stream, bodyRes
}

func TestIncomingBodyRead_ReadsData(t *testing.T) {
	ib, _, _ := newTestIncomingBody(context.Background(), []byte("hello world"))

	buf := make([]byte, 64)
	n, err := ib.Read(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 11 {
		t.Errorf("expected 11 bytes read, got %d", n)
	}
	if string(buf[:n]) != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", string(buf[:n]))
	}
}

func TestIncomingBodyRead_ReturnsEOFOnEmptyStream(t *testing.T) {
	ib, _, _ := newTestIncomingBody(context.Background(), []byte{})

	buf := make([]byte, 64)
	_, err := ib.Read(buf)
	if err != io.EOF {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

func TestIncomingBodyRead_FullReadThenEOF(t *testing.T) {
	ib, _, _ := newTestIncomingBody(context.Background(), []byte("abc"))

	buf := make([]byte, 3)
	n, err := ib.Read(buf)
	if err != nil || n != 3 {
		t.Fatalf("first read: n=%d err=%v", n, err)
	}
	if string(buf[:n]) != "abc" {
		t.Errorf("expected %q, got %q", "abc", string(buf[:n]))
	}

	_, err = ib.Read(buf)
	if err != io.EOF {
		t.Fatalf("expected io.EOF on second read, got %v", err)
	}
}

func TestIncomingBodyRead_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before any read

	ib, _, _ := newTestIncomingBody(ctx, []byte("hello"))
	// Override Subscribe to return a never-ready subscription so Await checks ctx.
	ib.stream.(*mockInputStream).subscribeFunc = func() subscription {
		return &neverReadySubscription{}
	}

	buf := make([]byte, 64)
	_, err := ib.Read(buf)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestIncomingBodyClose_DropsStream(t *testing.T) {
	ib, stream, _ := newTestIncomingBody(context.Background(), []byte("data"))

	if err := ib.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stream.dropCalled {
		t.Error("expected stream.Drop() to be called on Close")
	}
}

func TestIncomingBodyClose_CallsFinishOnce(t *testing.T) {
	ib, _, bodyRes := newTestIncomingBody(context.Background(), []byte("data"))

	// Closing twice must not panic (trailerOnce ensures Finish is called once).
	_ = ib.Close()
	_ = ib.Close()

	if !bodyRes.future.dropCalled {
		t.Error("expected future trailers to be dropped")
	}
}

func TestIncomingBodyRead_EOFTriggersTrailerParsing(t *testing.T) {
	ib, _, bodyRes := newTestIncomingBody(context.Background(), []byte{})

	buf := make([]byte, 64)
	_, err := ib.Read(buf)
	if err != io.EOF {
		t.Fatalf("expected io.EOF, got %v", err)
	}

	// EOF on read should have called Finish (via trailerOnce).
	if !bodyRes.future.dropCalled {
		t.Error("expected future trailers to be dropped after EOF read")
	}
}
