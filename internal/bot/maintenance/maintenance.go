// Package maintenance holds the global maintenance switch.
//
// The state is in-memory only (atomic bool): it resets on restart, which is
// intentional — a restart ends the maintenance window.
//
// The switch is read by the router dispatch layer (which blocks non-owner
// interactions while enabled), by the presence scheduler (DND override) and by
// any scheduler that must skip its tick during maintenance.
package maintenance

import "sync/atomic"

var enabled atomic.Bool

func Enabled() bool { return enabled.Load() }

func Set(v bool) { enabled.Store(v) }

// Toggle flips maintenance mode and returns the new value.
func Toggle() bool {
	for {
		current := enabled.Load()
		if enabled.CompareAndSwap(current, !current) {
			return !current
		}
	}
}
