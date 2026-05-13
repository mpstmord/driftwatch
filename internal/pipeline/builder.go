package pipeline

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/example/driftwatch/internal/config"
	"github.com/example/driftwatch/internal/cooldown"
	"github.com/example/driftwatch/internal/dedupe"
	"github.com/example/driftwatch/internal/filter"
	"github.com/example/driftwatch/internal/notifier"
)

// FromConfig constructs a ready-to-use Pipeline driven by the application
// Config. It is the primary entry-point used by cmd/driftwatch/main.go.
func FromConfig(cfg *config.Config, logSink io.Writer) *Pipeline {
	if logSink == nil {
		logSink = os.Stderr
	}

	f := filter.New(filter.Config{
		Include: cfg.IncludePatterns,
		Exclude: cfg.ExcludePatterns,
	})

	dedupeWindow := 5 * time.Minute
	if cfg.DedupeWindow > 0 {
		dedupeWindow = cfg.DedupeWindow
	}
	d := dedupe.New(dedupe.Config{Window: dedupeWindow})

	cooldownPeriod := 30 * time.Second
	if cfg.CooldownPeriod > 0 {
		cooldownPeriod = cfg.CooldownPeriod
	}
	c := cooldown.New(cooldown.Config{Period: cooldownPeriod})

	n := notifier.New(notifier.Config{Sink: logSink})

	logger := log.New(logSink, "[pipeline] ", log.LstdFlags)

	return New(Config{
		Filter:   f,
		Dedupe:   d,
		Cooldown: c,
		Notifier: n,
		Logger:   logger,
	})
}
