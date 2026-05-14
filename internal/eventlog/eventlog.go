// Package eventlog provides an in-memory ring buffer of recent drift events
// that can be queried for observability and debugging purposes.
package eventlog

import (
	"sync"
	"time"

	"github.com/driftwatch/internal/watcher"
)

// Entry represents a single logged drift event.
type Entry struct {
	Path      string
	OldHash   string
	NewHash   string
	DetectedAt time.Time
}

// EventLog is a bounded ring buffer of drift event entries.
type EventLog struct {
	mu      sync.RWMutex
	entries []Entry
	maxLen  int
}

// New creates an EventLog that retains at most maxLen entries.
// If maxLen is <= 0, it defaults to 100.
func New(maxLen int) *EventLog {
	if maxLen <= 0 {
		maxLen = 100
	}
	return &EventLog{
		entries: make([]Entry, 0, maxLen),
		maxLen:  maxLen,
	}
}

// Record appends a drift result to the log, evicting the oldest entry
// when the buffer is full.
func (el *EventLog) Record(r watcher.DriftResult) {
	if !r.Drifted {
		return
	}
	el.mu.Lock()
	defer el.mu.Unlock()

	entry := Entry{
		Path:       r.Path,
		OldHash:    r.OldHash,
		NewHash:    r.NewHash,
		DetectedAt: time.Now(),
	}

	if len(el.entries) >= el.maxLen {
		el.entries = el.entries[1:]
	}
	el.entries = append(el.entries, entry)
}

// All returns a snapshot copy of all entries currently in the log,
// ordered from oldest to newest.
func (el *EventLog) All() []Entry {
	el.mu.RLock()
	defer el.mu.RUnlock()

	snap := make([]Entry, len(el.entries))
	copy(snap, el.entries)
	return snap
}

// Len returns the number of entries currently held.
func (el *EventLog) Len() int {
	el.mu.RLock()
	defer el.mu.RUnlock()
	return len(el.entries)
}

// Clear removes all entries from the log.
func (el *EventLog) Clear() {
	el.mu.Lock()
	defer el.mu.Unlock()
	el.entries = el.entries[:0]
}
