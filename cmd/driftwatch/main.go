// Package main is the entry point for the driftwatch daemon.
// It wires together configuration loading, scheduling, diffing,
// alerting, auditing, and reporting into a single runnable process.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/driftwatch/internal/audit"
	"github.com/yourorg/driftwatch/internal/config"
	"github.com/yourorg/driftwatch/internal/filter"
	"github.com/yourorg/driftwatch/internal/notifier"
	"github.com/yourorg/driftwatch/internal/reporter"
	"github.com/yourorg/driftwatch/internal/runner"
	"github.com/yourorg/driftwatch/internal/throttle"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "driftwatch: fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		cfgPath    = flag.String("config", "", "path to driftwatch config file (YAML)")
		logFormat  = flag.String("format", "text", "report format: text or json")
		auditPath  = flag.String("audit-log", "", "path to persistent audit log (optional)")
		cooldown   = flag.Duration("cooldown", 5*time.Minute, "minimum time between repeated alerts for the same file")
		printVer   = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()

	if *printVer {
		fmt.Println("driftwatch v0.1.0")
		return nil
	}

	// Load configuration — fall back to watch paths supplied as positional args.
	watchPaths := flag.Args()
	cfg, err := config.LoadOrDefault(*cfgPath, watchPaths)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Build shared sub-systems.
	notify := notifier.New(os.Stderr)

	rpt, err := reporter.New(*logFormat, os.Stdout)
	if err != nil {
		return fmt.Errorf("creating reporter: %w", err)
	}

	throt := throttle.New(*cooldown)

	var auditor *audit.Log
	if *auditPath != "" {
		auditor, err = audit.New(*auditPath)
		if err != nil {
			return fmt.Errorf("opening audit log %q: %w", *auditPath, err)
		}
	}

	var fil *filter.Filter
	if len(cfg.Include) > 0 || len(cfg.Exclude) > 0 {
		fil = filter.New(cfg.Include, cfg.Exclude)
	}

	// Assemble the runner with all dependencies.
	r, err := runner.New(runner.Options{
		Config:    cfg,
		Notifier:  notify,
		Reporter:  rpt,
		Throttle:  throt,
		Auditor:   auditor,
		Filter:    fil,
	})
	if err != nil {
		return fmt.Errorf("initialising runner: %w", err)
	}

	// Honour SIGINT / SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Fprintf(os.Stderr, "driftwatch: watching %d path(s) every %s\n",
		len(cfg.WatchPaths), cfg.Interval)

	if err := r.Run(ctx); err != nil && err != context.Canceled {
		return fmt.Errorf("runner exited with error: %w", err)
	}

	fmt.Fprintln(os.Stderr, "driftwatch: shutdown complete")
	return nil
}
