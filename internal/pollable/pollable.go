package pollable

import (
	"context"
	"runtime"
	"time"

	monotonicclock "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_clocks_monotonic_clock"
	poll "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_poll"
)

const pollInterval = 10 * time.Microsecond

// Await blocks until the pollable is ready or the context is cancelled.
// Returns ctx.Err() if the context is cancelled before the pollable is ready.
// Each iteration yields to the Go scheduler then blocks on a host-level timer
// to allow the WASM host to advance async I/O.
func Await(ctx context.Context, p *poll.Pollable) error {
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
