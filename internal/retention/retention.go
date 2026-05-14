// Package retention manages automatic expiry of drift records older than a
// configurable age. It is safe for concurrent use.
package retention

import (
	"sync"
	"time"
)

// Entry holds a single retained drift record.
type Entry struct {
	Path      string
	RecordedAt time.Time
	Checksum  string
}

// Policy defines how long drift records are kept.
type Policy struct {
	MaxAge time.Duration
}

// Store holds drift entries and evicts those that exceed the policy age.
type Store struct {
	mu      sync.Mutex
	entries []Entry
	policy  Policy
	now     func() time.Time
}

// New creates a Store with the given retention policy.
func New(p Policy) *Store {
	return &Store{
		policy: p,
		now:    time.Now,
	}
}

// Add appends a new drift entry to the store.
func (s *Store) Add(path, checksum string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, Entry{
		Path:       path,
		Checksum:   checksum,
		RecordedAt: s.now(),
	})
}

// Evict removes all entries older than the policy MaxAge and returns the
// number of records removed.
func (s *Store) Evict() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := s.now().Add(-s.policy.MaxAge)
	kept := s.entries[:0]
	removed := 0
	for _, e := range s.entries {
		if e.RecordedAt.After(cutoff) {
			kept = append(kept, e)
		} else {
			removed++
		}
	}
	s.entries = kept
	return removed
}

// All returns a snapshot of current entries (post-eviction is not implied).
func (s *Store) All() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Entry, len(s.entries))
	copy(out, s.entries)
	return out
}

// Len returns the current number of stored entries.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}
