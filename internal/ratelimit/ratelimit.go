// Package ratelimit provides a token-bucket style rate limiter for
// controlling how frequently drift alerts are emitted per watched path.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter enforces a maximum number of events per path within a rolling window.
type Limiter struct {
	mu      sync.Mutex
	window  time.Duration
	max     int
	buckets map[string][]time.Time
}

// New creates a Limiter that allows at most maxEvents within the given window
// for any single key (e.g. a file path).
func New(window time.Duration, maxEvents int) *Limiter {
	return &Limiter{
		window:  window,
		max:     maxEvents,
		buckets: make(map[string][]time.Time),
	}
}

// Allow reports whether an event for the given key should be allowed through.
// It records the event timestamp and prunes entries outside the current window.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	times := l.buckets[key]
	pruned := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}

	if len(pruned) >= l.max {
		l.buckets[key] = pruned
		return false
	}

	l.buckets[key] = append(pruned, now)
	return true
}

// Reset clears all recorded events for the given key.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// Remaining returns how many more events are allowed for key within the
// current window without advancing state.
func (l *Limiter) Remaining(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	count := 0
	for _, t := range l.buckets[key] {
		if t.After(cutoff) {
			count++
		}
	}

	remaining := l.max - count
	if remaining < 0 {
		return 0
	}
	return remaining
}
