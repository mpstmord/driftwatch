// Package escalation tracks repeated drift events for a given path and
// escalates the alert severity once a configurable threshold is crossed.
package escalation

import (
	"sync"
	"time"
)

// Level represents an alert severity level.
type Level int

const (
	LevelInfo  Level = iota // first occurrence
	LevelWarn               // threshold approaching
	LevelCrit               // threshold exceeded
)

func (l Level) String() string {
	switch l {
	case LevelWarn:
		return "WARN"
	case LevelCrit:
		return "CRIT"
	default:
		return "INFO"
	}
}

// entry holds per-path hit counts and the time the window opened.
type entry struct {
	count     int
	windowAt  time.Time
}

// Tracker counts drift events per path within a sliding window and returns
// the appropriate Level for each new event.
type Tracker struct {
	mu        sync.Mutex
	entries   map[string]*entry
	warnAt    int           // hits before WARN
	critAt    int           // hits before CRIT
	window    time.Duration // window over which hits are counted
	now       func() time.Time
}

// New returns a Tracker that emits LevelWarn after warnAt hits and
// LevelCrit after critAt hits within window. critAt must be > warnAt.
func New(warnAt, critAt int, window time.Duration) *Tracker {
	return &Tracker{
		entries: make(map[string]*entry),
		warnAt:  warnAt,
		critAt:  critAt,
		window:  window,
		now:     time.Now,
	}
}

// Record registers a drift event for path and returns the current Level.
func (t *Tracker) Record(path string) Level {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	e, ok := t.entries[path]
	if !ok || now.Sub(e.windowAt) > t.window {
		e = &entry{windowAt: now}
		t.entries[path] = e
	}
	e.count++

	switch {
	case e.count >= t.critAt:
		return LevelCrit
	case e.count >= t.warnAt:
		return LevelWarn
	default:
		return LevelInfo
	}
}

// Reset clears the hit counter for path, useful when drift is resolved.
func (t *Tracker) Reset(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, path)
}
