// Package replay provides a mechanism for replaying recorded drift events
// from the event log, useful for diagnostics and catch-up after downtime.
package replay

import (
	"context"
	"time"

	"github.com/driftwatch/internal/eventlog"
	"github.com/driftwatch/internal/watcher"
)

// Handler is called for each replayed drift result.
type Handler func(result watcher.DriftResult)

// Options controls replay behaviour.
type Options struct {
	// Since filters events to those recorded at or after this time.
	// Zero value means replay all events.
	Since time.Time

	// Delay introduces an artificial pause between replayed events.
	// Zero means no delay.
	Delay time.Duration
}

// Replayer reads from an EventLog and re-delivers drift results to a handler.
type Replayer struct {
	log *eventlog.EventLog
	opts Options
}

// New returns a Replayer backed by the given EventLog.
func New(log *eventlog.EventLog, opts Options) *Replayer {
	return &Replayer{log: log, opts: opts}
}

// Run iterates over recorded events and calls h for each one that passes
// the Since filter. It respects context cancellation.
func (r *Replayer) Run(ctx context.Context, h Handler) error {
	events := r.log.All()
	for _, entry := range events {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !r.opts.Since.IsZero() && entry.RecordedAt.Before(r.opts.Since) {
			continue
		}

		h(entry.Result)

		if r.opts.Delay > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.opts.Delay):
			}
		}
	}
	return nil
}
