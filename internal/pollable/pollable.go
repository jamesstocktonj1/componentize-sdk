package pollable

import (
	"context"
	"runtime"
	"time"

	monotonicclock "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_clocks_monotonic_clock"
)

const pollInterval = 10 * time.Microsecond

// Readable is the minimal interface satisfied by all WASI pollable types.
type Readable interface {
	Ready() bool
	Drop()
}

// AwaitContext blocks until the pollable is ready or the context is cancelled.
// Returns ctx.Err() if the context is cancelled before the pollable is ready.
// Each iteration yields to the Go scheduler then blocks on a host-level timer
// to allow the WASM host to advance async I/O.
func AwaitContext(ctx context.Context, p Readable) error {
	for !p.Ready() {
		if err := ctx.Err(); err != nil {
			return err
		}
		runtime.Gosched()

		timer := monotonicclock.SubscribeDuration(monotonicclock.Duration(pollInterval))
		defer timer.Drop()
		timer.Block()
	}
	return nil
}

// Await blocks until the pollable is ready, yielding to the Go scheduler
// on each iteration to allow other goroutines to run.
func Await(p Readable) error {
	return AwaitContext(context.Background(), p)
}

// AwaitAndDrop waits until the pollable is ready using the goroutine-friendly
// Await, then drops it.
func AwaitAndDrop(p Readable) {
	defer p.Drop()
	_ = Await(p)
}
