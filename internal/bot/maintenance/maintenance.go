package maintenance

import "sync/atomic"

var enabled atomic.Bool

func Enabled() bool { return enabled.Load() }

func Set(v bool) { enabled.Store(v) }

func Toggle() bool {
	for {
		current := enabled.Load()
		if enabled.CompareAndSwap(current, !current) {
			return !current
		}
	}
}
