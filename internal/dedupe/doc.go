// Package dedupe implements drift-event deduplication for driftwatch.
//
// When a file drifts, the watcher may emit the same event on every
// polling cycle until the file is remediated. Dedupe prevents the
// alert pipeline from being flooded by identical (path, hash) pairs
// within a configurable time window.
//
// Usage:
//
//	dd := dedupe.New(10 * time.Minute)
//
//	// Inside the polling loop:
//	if !dd.IsDuplicate(result.Path, result.CurrentHash) {
//		// forward the drift event to the notifier
//	}
//
// Forget and Reset allow explicit invalidation when a file is
// known to have changed state (e.g. after a baseline update).
package dedupe
