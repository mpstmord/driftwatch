// Package digest provides utilities for tracking and comparing file digest
// histories, enabling driftwatch to detect repeated or transient changes.
package digest

import (
	"fmt"
	"sync"
	"time"
)

// Entry records a digest value captured at a specific point in time.
type Entry struct {
	Hash      string
	CapturedAt time.Time
}

// History maintains a bounded, ordered log of digest entries per file path.
type History struct {
	mu      sync.Mutex
	records map[string][]Entry
	maxLen  int
}

// New creates a History that retains at most maxLen entries per path.
// If maxLen is less than 1 it defaults to 10.
func New(maxLen int) *History {
	if maxLen < 1 {
		maxLen = 10
	}
	return &History{
		records: make(map[string][]Entry),
		maxLen:  maxLen,
	}
}

// Record appends a new digest entry for the given path, evicting the oldest
// entry when the history for that path exceeds maxLen.
func (h *History) Record(path, hash string, at time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()

	entries := append(h.records[path], Entry{Hash: hash, CapturedAt: at})
	if len(entries) > h.maxLen {
		entries = entries[len(entries)-h.maxLen:]
	}
	h.records[path] = entries
}

// Latest returns the most recently recorded Entry for path and true, or an
// empty Entry and false if no history exists for that path.
func (h *History) Latest(path string) (Entry, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	entries, ok := h.records[path]
	if !ok || len(entries) == 0 {
		return Entry{}, false
	}
	return entries[len(entries)-1], true
}

// All returns a copy of all recorded entries for path in chronological order.
func (h *History) All(path string) []Entry {
	h.mu.Lock()
	defer h.mu.Unlock()

	src := h.records[path]
	out := make([]Entry, len(src))
	copy(out, src)
	return out
}

// Changed reports whether the most recent hash for path differs from prev.
// It returns false when no history exists (nothing recorded yet).
func (h *History) Changed(path, prev string) (bool, error) {
	latest, ok := h.Latest(path)
	if !ok {
		return false, fmt.Errorf("digest: no history for path %q", path)
	}
	return latest.Hash != prev, nil
}
