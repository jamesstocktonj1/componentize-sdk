package pollable

import (
	"context"
	"runtime"

	poll "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_io_poll"
)

// Await blocks until the pollable is ready or the context is cancelled.
// Returns ctx.Err() if the context is cancelled before the pollable is ready.
func Await(ctx context.Context, p *poll.Pollable) error {
	for !p.Ready() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			runtime.Gosched()
		}
	}
	return nil
}
