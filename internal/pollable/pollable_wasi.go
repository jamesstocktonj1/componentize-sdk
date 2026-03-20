//go:build wasip1

package pollable

import (
	"time"

	monotonicclock "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_clocks_monotonic_clock"
)

func init() {
	defaultTimerFactory = func(d time.Duration) blockTimer {
		return monotonicclock.SubscribeDuration(monotonicclock.Duration(d))
	}
}
