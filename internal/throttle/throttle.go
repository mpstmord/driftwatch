// Package throttle provides rate-limiting for drift alerts to prevent
// alert storms when many files change simultaneously.
package throttle

import (
	"sync"
	"time"
)

// Throttle tracks per-key alert times and suppresses repeated alerts
// within a configurable cooldown window.
type Throttle struct {
	mu       sync.Mutex
	cooldown time.Duration
	lastSent map[string]time.Time
	now      func() time.Time
}

// New creates a Throttle with the given cooldown duration.
// Alerts for the same key will be suppressed until the cooldown elapses.
func New(cooldown time.Duration) *Throttle {
	return &Throttle{
		cooldown: cooldown,
		lastSent: make(map[string]time.Time),
		now:      time.Now,
	}
}

// Allow returns true if an alert for the given key should be sent.
// It updates the last-sent timestamp when it returns true.
func (t *Throttle) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	if last, ok := t.lastSent[key]; ok {
		if now.Sub(last) < t.cooldown {
			return false
		}
	}
	t.lastSent[key] = now
	return true
}

// Reset clears the last-sent record for the given key, allowing the
// next alert through regardless of the cooldown.
func (t *Throttle) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.lastSent, key)
}

// ResetAll clears all throttle state.
func (t *Throttle) ResetAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastSent = make(map[string]time.Time)
}
