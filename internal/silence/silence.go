// Package silence provides a registry for temporarily silencing drift
// alerts on specific paths. A silenced path will not produce notifications
// until the silence expires or is explicitly lifted.
package silence

import (
	"sync"
	"time"
)

// entry holds the expiry time for a single silenced path.
type entry struct {
	until time.Time
}

// Silence tracks per-path silences with optional expiry.
type Silence struct {
	mu      sync.RWMutex
	entries map[string]entry
	now     func() time.Time
}

// New returns a new Silence registry.
func New() *Silence {
	return &Silence{
		entries: make(map[string]entry),
		now:     time.Now,
	}
}

// Add silences the given path for the specified duration.
// Calling Add on an already-silenced path replaces the previous expiry.
func (s *Silence) Add(path string, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[path] = entry{until: s.now().Add(d)}
}

// Lift removes any active silence for the given path immediately.
func (s *Silence) Lift(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, path)
}

// IsSilenced reports whether the path currently has an active silence.
// Expired silences are lazily evicted on read.
func (s *Silence) IsSilenced(path string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[path]
	if !ok {
		return false
	}
	if s.now().After(e.until) {
		delete(s.entries, path)
		return false
	}
	return true
}

// Active returns a snapshot of all currently silenced paths and their
// remaining duration.
func (s *Silence) Active() map[string]time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	out := make(map[string]time.Duration, len(s.entries))
	for path, e := range s.entries {
		if now.After(e.until) {
			delete(s.entries, path)
			continue
		}
		out[path] = e.until.Sub(now)
	}
	return out
}
