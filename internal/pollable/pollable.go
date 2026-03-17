package pollable

import (
	"context"
	"runtime"
	"time"

	monotonicclock "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_clocks_monotonic_clock"
)

const pollInterval = 10 * time.Microsecond

// ReadyChecker reports whether an I/O resource is ready.
// *poll.Pollable satisfies this interface.
type ReadyChecker interface {
	Ready() bool
}

// blockTimer is an internal interface for host timer resources used in Await.
type blockTimer interface {
	Block()
	Drop()
}

// timerFactory creates a blockTimer for a given duration.
type timerFactory func(d time.Duration) blockTimer

var defaultTimerFactory timerFactory = func(d time.Duration) blockTimer {
	return monotonicclock.SubscribeDuration(monotonicclock.Duration(d))
}

// Await blocks until the pollable is ready or the context is cancelled.
// Returns ctx.Err() if the context is cancelled before the pollable is ready.
// Each iteration yields to the Go scheduler then blocks on a host-level timer
// to allow the WASM host to advance async I/O.
func Await(ctx context.Context, p ReadyChecker) error {
	return awaitWithFactory(ctx, p, defaultTimerFactory)
}

func awaitWithFactory(ctx context.Context, p ReadyChecker, factory timerFactory) error {
	for !p.Ready() {
		if err := ctx.Err(); err != nil {
			return err
		}
		runtime.Gosched()

		timer := factory(pollInterval)
		defer timer.Drop()
		timer.Block()
	}
	return nil
}
