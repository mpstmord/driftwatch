// Package runner ties together the watcher, differ, baseline, and alert
// subsystems into a single top-level orchestration loop.
package runner

import (
	"context"
	"log"

	"github.com/example/driftwatch/internal/alert"
	"github.com/example/driftwatch/internal/baseline"
	"github.com/example/driftwatch/internal/config"
	"github.com/example/driftwatch/internal/differ"
	"github.com/example/driftwatch/internal/reporter"
	"github.com/example/driftwatch/internal/snapshot"
)

// Runner orchestrates a single drift-check cycle.
type Runner struct {
	cfg      *config.Config
	baseline *baseline.Baseline
	alert    *alert.Alerter
	reporter *reporter.Reporter
}

// New creates a Runner from the supplied configuration.
func New(cfg *config.Config) (*Runner, error) {
	b, err := baseline.Load(cfg.BaselinePath)
	if err != nil {
		log.Printf("runner: no existing baseline, starting fresh: %v", err)
		b = baseline.New()
	}

	a := alert.New(nil) // defaults to stdout
	r := reporter.New(cfg.Format)

	return &Runner{
		cfg:      cfg,
		baseline: b,
		alert:    a,
		reporter: r,
	}, nil
}

// Run performs one full snapshot + diff cycle and emits alerts for any drift
// detected. It respects ctx cancellation.
func (r *Runner) Run(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	snap, err := snapshot.New(r.cfg.WatchPaths)
	if err != nil {
		return err
	}

	results := differ.Diff(r.baseline, snap)

	if err := r.reporter.Report(results); err != nil {
		log.Printf("runner: reporter error: %v", err)
	}

	for _, res := range results {
		if res.Drifted {
			if err := r.alert.Drift(res.Path, res.Expected, res.Actual); err != nil {
				log.Printf("runner: alert error for %s: %v", res.Path, err)
			}
		}
	}

	return nil
}

// UpdateBaseline snapshots the current state and persists it as the new baseline.
func (r *Runner) UpdateBaseline() error {
	snap, err := snapshot.New(r.cfg.WatchPaths)
	if err != nil {
		return err
	}
	for path, rec := range snap.Records {
		r.baseline.Set(path, rec)
	}
	return baseline.Save(r.baseline, r.cfg.BaselinePath)
}
