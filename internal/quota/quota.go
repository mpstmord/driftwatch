// Package quota enforces per-path alert quotas over a rolling time window.
// Once a path exceeds its allowed alert count within the window it is
// suppressed until the window rolls over.
package quota

import (
	"sync"
	"time"
)

// entry tracks the number of alerts issued for a single path within the
// current window.
type entry struct {
	count     int
	windowEnd time.Time
}

// Quota enforces a maximum number of alerts per path per window duration.
type Quota struct {
	mu      sync.Mutex
	entries map[string]*entry
	max     int
	window  time.Duration
	now     func() time.Time
}

// New returns a Quota that allows at most max alerts per path within window.
func New(max int, window time.Duration) *Quota {
	return &Quota{
		entries: make(map[string]*entry),
		max:     max,
		window:  window,
		now:     time.Now,
	}
}

// Allow reports whether an alert for path is permitted under the current quota.
// It increments the counter when the call is allowed.
func (q *Quota) Allow(path string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.now()
	e, ok := q.entries[path]
	if !ok || now.After(e.windowEnd) {
		q.entries[path] = &entry{count: 1, windowEnd: now.Add(q.window)}
		return true
	}
	if e.count >= q.max {
		return false
	}
	e.count++
	return true
}

// Remaining returns how many alerts are still permitted for path in the
// current window. A negative value indicates the window has already expired
// and will be reset on the next Allow call.
func (q *Quota) Remaining(path string) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	e, ok := q.entries[path]
	if !ok || q.now().After(e.windowEnd) {
		return q.max
	}
	rem := q.max - e.count
	if rem < 0 {
		return 0
	}
	return rem
}

// Reset clears the quota state for path, allowing it to be treated as new.
func (q *Quota) Reset(path string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.entries, path)
}
