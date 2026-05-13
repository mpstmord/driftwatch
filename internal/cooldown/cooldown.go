// Package cooldown provides per-key cooldown tracking to prevent
// repeated alerts from firing too frequently for the same resource.
package cooldown

import (
	"sync"
	"time"
)

// Cooldown tracks the last alert time per key and enforces a minimum
// duration between successive alerts for the same key.
type Cooldown struct {
	mu       sync.Mutex
	lastSeen map[string]time.Time
	duration time.Duration
	now      func() time.Time
}

// New creates a Cooldown with the given minimum duration between alerts.
// Keys that have not been seen before are always allowed through.
func New(d time.Duration) *Cooldown {
	return &Cooldown{
		lastSeen: make(map[string]time.Time),
		duration: d,
		now:      time.Now,
	}
}

// Ready reports whether the given key is ready to fire again.
// It returns true (and records the time) if the key has never fired
// or if the cooldown period has elapsed since the last fire.
func (c *Cooldown) Ready(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	last, seen := c.lastSeen[key]
	if !seen || now.Sub(last) >= c.duration {
		c.lastSeen[key] = now
		return true
	}
	return false
}

// Reset clears the cooldown state for the given key so that the next
// call to Ready will always return true.
func (c *Cooldown) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.lastSeen, key)
}

// Remaining returns the time left in the cooldown for the given key.
// It returns zero if the key is already ready to fire.
func (c *Cooldown) Remaining(key string) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

	last, seen := c.lastSeen[key]
	if !seen {
		return 0
	}
	elapsed := c.now().Sub(last)
	if elapsed >= c.duration {
		return 0
	}
	return c.duration - elapsed
}
