// Package debounce prevents repeated alerts for the same path within a
// configurable quiet window. Unlike throttle (which uses a cooldown per key),
// debounce resets the timer on every new event, delaying delivery until the
// stream of events goes quiet.
package debounce

import (
	"sync"
	"time"
)

// Debouncer delays action callbacks until no new events arrive for the
// configured wait duration.
type Debouncer struct {
	wait  time.Duration
	mu    sync.Mutex
	timers map[string]*time.Timer
	now   func() time.Time // injectable for tests
}

// New creates a Debouncer with the given quiet-window duration.
func New(wait time.Duration) *Debouncer {
	return &Debouncer{
		wait:   wait,
		timers: make(map[string]*time.Timer),
		now:    time.Now,
	}
}

// Trigger schedules fn to be called after the quiet window elapses for key.
// If Trigger is called again for the same key before the window expires, the
// previous scheduled call is cancelled and the timer resets.
func (d *Debouncer) Trigger(key string, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if t, ok := d.timers[key]; ok {
		t.Stop()
	}

	d.timers[key] = time.AfterFunc(d.wait, func() {
		d.mu.Lock()
		delete(d.timers, key)
		d.mu.Unlock()
		fn()
	})
}

// Cancel stops any pending timer for key without invoking the callback.
func (d *Debouncer) Cancel(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if t, ok := d.timers[key]; ok {
		t.Stop()
		delete(d.timers, key)
	}
}

// Pending returns true if there is an outstanding timer for key.
func (d *Debouncer) Pending(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.timers[key]
	return ok
}
