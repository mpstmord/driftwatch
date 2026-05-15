// Package acknowledge provides a mechanism for operators to acknowledge
// known drift events, suppressing repeated alerts for a specific path
// until the acknowledgement expires or is lifted.
package acknowledge

import (
	"sync"
	"time"
)

// entry holds acknowledgement state for a single path.
type entry struct {
	AckedAt  time.Time
	ExpiresAt time.Time
	Reason   string
}

// Store tracks acknowledged drift paths.
type Store struct {
	mu      sync.RWMutex
	entries map[string]entry
	now     func() time.Time
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		entries: make(map[string]entry),
		now:     time.Now,
	}
}

// Acknowledge marks path as acknowledged for the given duration with an
// optional human-readable reason. Calling Acknowledge on an already-
// acknowledged path resets the expiry.
func (s *Store) Acknowledge(path, reason string, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := s.now()
	s.entries[path] = entry{
		AckedAt:   n,
		ExpiresAt: n.Add(d),
		Reason:    reason,
	}
}

// IsAcknowledged reports whether path is currently acknowledged.
func (s *Store) IsAcknowledged(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[path]
	if !ok {
		return false
	}
	return s.now().Before(e.ExpiresAt)
}

// Lift removes an acknowledgement for path before it expires.
func (s *Store) Lift(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, path)
}

// Reason returns the reason string for an active acknowledgement, and
// whether one exists.
func (s *Store) Reason(path string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[path]
	if !ok || !s.now().Before(e.ExpiresAt) {
		return "", false
	}
	return e.Reason, true
}
