// Package dedupe implements drift-event deduplication for driftwatch.
//
// When a file drifts, the watcher may emit the same event on every
// polling cycle until the file is remediated. Dedupe prevents the
// alert pipeline from being flooded by identical (path, hash) pairs
// within a configurable time window.
//
// # How it works
//
// Each call to IsDuplicate records the (path, hash) pair along with
// the current timestamp. Subsequent calls with the same pair are
// considered duplicates until the configured TTL has elapsed, at
// which point the entry expires and the event is treated as new.
//
// # Usage
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
//
// # Concurrency
//
// All exported methods are safe for concurrent use by multiple
// goroutines. The internal state is protected by a sync.Mutex.
package dedupe
