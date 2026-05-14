// Package rollup aggregates multiple drift events within a time window
// and emits a single summary instead of one alert per file.
package rollup

import (
	"context"
	"sync"
	"time"

	"github.com/example/driftwatch/internal/watcher"
)

// Handler is called with the accumulated drift results after each window.
type Handler func(results []watcher.DriftResult)

// Rollup buffers DriftResults and flushes them on a fixed window.
type Rollup struct {
	mu      sync.Mutex
	buf     []watcher.DriftResult
	window  time.Duration
	handler Handler
}

// New creates a Rollup that flushes buffered results every window duration.
func New(window time.Duration, handler Handler) *Rollup {
	return &Rollup{
		window:  window,
		handler: handler,
	}
}

// Add buffers a DriftResult for the next flush.
func (r *Rollup) Add(result watcher.DriftResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf = append(r.buf, result)
}

// Run starts the flush loop. It blocks until ctx is cancelled.
func (r *Rollup) Run(ctx context.Context) {
	ticker := time.NewTicker(r.window)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.flush()
		case <-ctx.Done():
			r.flush()
			return
		}
	}
}

// flush drains the buffer and calls the handler if there is anything to report.
func (r *Rollup) flush() {
	r.mu.Lock()
	batch := r.buf
	r.buf = nil
	r.mu.Unlock()
	if len(batch) > 0 {
		r.handler(batch)
	}
}

// Len returns the number of buffered results (useful for testing).
func (r *Rollup) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.buf)
}
