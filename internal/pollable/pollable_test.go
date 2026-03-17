package pollable

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockPollable implements ReadyChecker and returns true after readyAfter calls.
type mockPollable struct {
	readyAfter int
	callCount  int
}

func (m *mockPollable) Ready() bool {
	m.callCount++
	return m.callCount > m.readyAfter
}

// mockTimer implements blockTimer and records calls.
type mockTimer struct {
	blockCalled int
	dropCalled  int
}

func (m *mockTimer) Block() { m.blockCalled++ }
func (m *mockTimer) Drop()  { m.dropCalled++ }

func TestAwaitWithFactory(t *testing.T) {
	t.Run("immediately ready", func(t *testing.T) {
		p := &mockPollable{readyAfter: 0}
		timer := &mockTimer{}
		factory := func(_ time.Duration) blockTimer { return timer }

		if err := awaitWithFactory(context.Background(), p, factory); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if timer.blockCalled != 0 {
			t.Errorf("expected no Block() calls, got %d", timer.blockCalled)
		}
	})

	t.Run("ready after polling", func(t *testing.T) {
		// readyAfter: 2 means Ready() returns false twice, then true
		p := &mockPollable{readyAfter: 2}
		timer := &mockTimer{}
		factory := func(_ time.Duration) blockTimer { return timer }

		if err := awaitWithFactory(context.Background(), p, factory); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if timer.blockCalled != 2 {
			t.Errorf("expected 2 Block() calls, got %d", timer.blockCalled)
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		p := &mockPollable{readyAfter: 100}
		err := awaitWithFactory(ctx, p, func(_ time.Duration) blockTimer { return &mockTimer{} })
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
		defer cancel()

		p := &mockPollable{readyAfter: 100}
		err := awaitWithFactory(ctx, p, func(_ time.Duration) blockTimer { return &mockTimer{} })
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected context.DeadlineExceeded, got %v", err)
		}
	})

	t.Run("passes poll interval to timer", func(t *testing.T) {
		var receivedDuration time.Duration
		p := &mockPollable{readyAfter: 1}
		factory := func(d time.Duration) blockTimer {
			receivedDuration = d
			return &mockTimer{}
		}

		if err := awaitWithFactory(context.Background(), p, factory); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedDuration != pollInterval {
			t.Errorf("expected pollInterval %v, got %v", pollInterval, receivedDuration)
		}
	})
}
