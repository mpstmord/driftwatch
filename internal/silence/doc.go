// Package silence provides a per-path alert silence registry for driftwatch.
//
// Silences allow operators to temporarily suppress drift notifications for
// known-changing paths (e.g. during a planned deployment) without disabling
// monitoring globally.
//
// Usage:
//
//	sr := silence.New()
//	sr.Add("/etc/myapp/config.yaml", 30*time.Minute)
//
//	// later, in the alert pipeline:
//	if sr.IsSilenced(path) {
//		return // skip notification
//	}
//
// Silences expire automatically; no background goroutine is required.
package silence
