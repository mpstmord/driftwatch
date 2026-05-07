// Package throttle implements per-key rate limiting for drift alert
// notifications in driftwatch.
//
// When a large number of files drift simultaneously — for example after
// a bulk deployment — the throttle prevents the alerting subsystem from
// flooding downstream sinks with duplicate messages.
//
// Usage:
//
//	th := throttle.New(5 * time.Minute)
//	if th.Allow(path) {
//		alert.Send(path)
//	}
//
// Each unique key (typically a file path) is tracked independently.
// The cooldown window is shared across all keys but evaluated per-key.
package throttle
