// Package dedupe provides event deduplication for drift alerts,
// suppressing repeated notifications for the same file path within
// a configurable time window.
package dedupe

import (
	"sync"
	"time"
)

// entry holds the last seen hash and the time it was recorded.
type entry struct {
	hash      string
	recordedAt time.Time
}

// Deduper tracks previously seen drift events and filters duplicates.
type Deduper struct {
	mu      sync.Mutex
	entries map[string]entry
	window  time.Duration
	now     func() time.Time
}

// New creates a Deduper that suppresses repeated events for the same
// (path, hash) pair within the given window duration.
func New(window time.Duration) *Deduper {
	return &Deduper{
		entries: make(map[string]entry),
		window:  window,
		now:     time.Now,
	}
}

// IsDuplicate reports whether a drift event for path with the given
// content hash has already been seen within the deduplication window.
// If it is not a duplicate the event is recorded and false is returned.
func (d *Deduper) IsDuplicate(path, hash string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	if e, ok := d.entries[path]; ok {
		if e.hash == hash && now.Sub(e.recordedAt) < d.window {
			return true
		}
	}

	d.entries[path] = entry{hash: hash, recordedAt: now}
	return false
}

// Forget removes the stored entry for path, allowing the next event
// for that path to pass through regardless of the window.
func (d *Deduper) Forget(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.entries, path)
}

// Reset clears all stored entries.
func (d *Deduper) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.entries = make(map[string]entry)
}
