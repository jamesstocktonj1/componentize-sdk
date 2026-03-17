package httptypes

import (
	"context"
	"io"
	"net/http"
	"testing"

	streams "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_streams"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// mockSubscription implements subscription. Set ready to control readiness.
type mockSubscription struct {
	ready bool
}

func (s *mockSubscription) Ready() bool { return s.ready }
func (s *mockSubscription) Drop()       {}

// mockInputStream implements inputStream for testing.
type mockInputStream struct {
	data       []byte
	readPos    int
	dropCalled bool
	sub        mockSubscription
}

func (m *mockInputStream) Subscribe() subscription {
	return &m.sub
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

// mockIncomingBodyResource implements incomingBodyResource for testing.
type mockIncomingBodyResource struct {
	finishCalled bool
	dropCalled   bool
}

func (m *mockIncomingBodyResource) FinishAndGetTrailers(_ http.Header) { m.finishCalled = true }
func (m *mockIncomingBodyResource) Drop()                              { m.dropCalled = true }

func newTestIncomingBody(ctx context.Context, data []byte) (*incomingBody, *mockInputStream, *mockIncomingBodyResource) {
	stream := &mockInputStream{
		data: data,
		sub:  mockSubscription{ready: true},
	}
	bodyRes := &mockIncomingBodyResource{}

	ib := &incomingBody{
		ctx:     ctx,
		body:    bodyRes,
		stream:  stream,
		trailer: make(map[string][]string),
	}
	return ib, stream, bodyRes
}

func TestIncomingBodyRead(t *testing.T) {
	t.Run("reads data", func(t *testing.T) {
		ib, _, _ := newTestIncomingBody(context.Background(), []byte("hello world"))

		buf := make([]byte, 64)
		n, err := ib.Read(buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != 11 {
			t.Errorf("expected 11 bytes, got %d", n)
		}
		if string(buf[:n]) != "hello world" {
			t.Errorf("expected %q, got %q", "hello world", string(buf[:n]))
		}
	})

	t.Run("returns EOF on empty stream", func(t *testing.T) {
		ib, _, _ := newTestIncomingBody(context.Background(), []byte{})

		_, err := ib.Read(make([]byte, 64))
		if err != io.EOF {
			t.Fatalf("expected io.EOF, got %v", err)
		}
	})

	t.Run("full read then EOF", func(t *testing.T) {
		ib, _, _ := newTestIncomingBody(context.Background(), []byte("abc"))

		buf := make([]byte, 3)
		n, err := ib.Read(buf)
		if err != nil || n != 3 || string(buf[:n]) != "abc" {
			t.Fatalf("first read: n=%d err=%v data=%q", n, err, string(buf[:n]))
		}

		_, err = ib.Read(buf)
		if err != io.EOF {
			t.Fatalf("expected io.EOF on second read, got %v", err)
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		ib, stream, _ := newTestIncomingBody(ctx, []byte("hello"))
		stream.sub.ready = false // force Await to check ctx before proceeding

		_, err := ib.Read(make([]byte, 64))
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("EOF triggers trailer parsing", func(t *testing.T) {
		ib, _, bodyRes := newTestIncomingBody(context.Background(), []byte{})

		_, err := ib.Read(make([]byte, 64))
		if err != io.EOF {
			t.Fatalf("expected io.EOF, got %v", err)
		}
		if !bodyRes.finishCalled {
			t.Error("expected FinishAndGetTrailers to be called after EOF")
		}
	})
}

func TestIncomingBodyClose(t *testing.T) {
	t.Run("drops stream", func(t *testing.T) {
		ib, stream, _ := newTestIncomingBody(context.Background(), []byte("data"))

		if err := ib.Close(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !stream.dropCalled {
			t.Error("expected stream.Drop() to be called")
		}
	})

	t.Run("calls finish once", func(t *testing.T) {
		ib, _, bodyRes := newTestIncomingBody(context.Background(), []byte("data"))

		// Closing twice must not panic; trailerOnce ensures Finish runs once.
		_ = ib.Close()
		_ = ib.Close()

		if !bodyRes.finishCalled {
			t.Error("expected FinishAndGetTrailers to be called")
		}
	})
}
