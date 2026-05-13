// Package fingerprint tracks per-path change fingerprints to detect
// whether a file has changed since the last check cycle.
package fingerprint

import (
	"sync"
	"time"
)

// Record holds the last-seen checksum and the time it was recorded.
type Record struct {
	Checksum  string
	RecordedAt time.Time
}

// Tracker stores the most recent fingerprint for each watched path.
type Tracker struct {
	mu      sync.RWMutex
	records map[string]Record
	now     func() time.Time
}

// New returns an initialised Tracker.
func New() *Tracker {
	return &Tracker{
		records: make(map[string]Record),
		now:     time.Now,
	}
}

// Update stores a new checksum for path. It returns true if the checksum
// differs from the previously stored value (i.e. the file changed).
func (t *Tracker) Update(path, checksum string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	prev, exists := t.records[path]
	t.records[path] = Record{
		Checksum:   checksum,
		RecordedAt: t.now(),
	}

	if !exists {
		return false // first observation is never a change
	}
	return prev.Checksum != checksum
}

// Get returns the stored Record for path and whether one exists.
func (t *Tracker) Get(path string) (Record, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	r, ok := t.records[path]
	return r, ok
}

// Delete removes the fingerprint for path.
func (t *Tracker) Delete(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.records, path)
}

// Paths returns all tracked paths.
func (t *Tracker) Paths() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	paths := make([]string, 0, len(t.records))
	for p := range t.records {
		paths = append(paths, p)
	}
	return paths
}
