// Package pipeline wires together the drift-detection stages into a
// single reusable processing chain: filter → dedupe → cooldown → notify.
package pipeline

import (
	"context"
	"log"

	"github.com/example/driftwatch/internal/cooldown"
	"github.com/example/driftwatch/internal/dedupe"
	"github.com/example/driftwatch/internal/filter"
	"github.com/example/driftwatch/internal/notifier"
	"github.com/example/driftwatch/internal/watcher"
)

// Pipeline processes drift results through a series of gates before
// forwarding surviving events to the notifier.
type Pipeline struct {
	filter   *filter.Filter
	dedupe   *dedupe.Dedupe
	cooldown *cooldown.Cooldown
	notifier *notifier.Notifier
	log      *log.Logger
}

// Config holds constructor options for Pipeline.
type Config struct {
	Filter   *filter.Filter
	Dedupe   *dedupe.Dedupe
	Cooldown *cooldown.Cooldown
	Notifier *notifier.Notifier
	Logger   *log.Logger
}

// New creates a Pipeline from the provided Config.
// All fields are required; New panics if any is nil.
func New(cfg Config) *Pipeline {
	if cfg.Filter == nil || cfg.Dedupe == nil || cfg.Cooldown == nil || cfg.Notifier == nil {
		panic("pipeline: all Config fields must be non-nil")
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	return &Pipeline{
		filter:   cfg.Filter,
		dedupe:   cfg.Dedupe,
		cooldown: cfg.Cooldown,
		notifier: cfg.Notifier,
		log:      cfg.Logger,
	}
}

// Process evaluates a batch of DriftResults and forwards events that
// pass every gate to the notifier. It respects ctx cancellation.
func (p *Pipeline) Process(ctx context.Context, results []watcher.DriftResult) {
	for _, r := range results {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !r.Drifted {
			continue
		}

		if !p.filter.Allow(r.Path) {
			p.log.Printf("pipeline: filtered %s", r.Path)
			continue
		}

		if p.dedupe.IsDuplicate(r.Path, r.CurrentSum) {
			p.log.Printf("pipeline: dedupe suppressed %s", r.Path)
			continue
		}

		if !p.cooldown.Ready(r.Path) {
			p.log.Printf("pipeline: cooldown suppressed %s", r.Path)
			continue
		}

		p.notifier.Notify(notifier.Event{
			Path:    r.Path,
			Message: "drift detected",
		})
	}
}
