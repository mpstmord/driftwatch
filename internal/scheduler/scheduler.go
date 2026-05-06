package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/example/driftwatch/internal/alert"
	"github.com/example/driftwatch/internal/watcher"
)

// Scheduler periodically runs the drift watcher and dispatches alerts.
type Scheduler struct {
	interval time.Duration
	watcher  *watcher.Watcher
	handler  *alert.DriftHandler
}

// New creates a Scheduler that ticks at the given interval.
func New(interval time.Duration, w *watcher.Watcher, h *alert.DriftHandler) *Scheduler {
	return &Scheduler{
		interval: interval,
		watcher:  w,
		handler:  h,
	}
}

// Run starts the scheduling loop and blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Printf("scheduler: starting drift checks every %s", s.interval)

	// Run an immediate check before waiting for the first tick.
	s.runOnce()

	for {
		select {
		case <-ticker.C:
			s.runOnce()
		case <-ctx.Done():
			log.Println("scheduler: shutting down")
			return ctx.Err()
		}
	}
}

func (s *Scheduler) runOnce() {
	results := s.watcher.Check()
	for _, r := range results {
		if err := s.handler.Handle(r); err != nil {
			log.Printf("scheduler: alert error for %s: %v", r.Path, err)
		}
	}
}
