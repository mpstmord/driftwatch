package pipeline_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/config"
	"github.com/example/driftwatch/internal/pipeline"
	"github.com/example/driftwatch/internal/watcher"
)

func minimalConfig(paths []string) *config.Config {
	return &config.Config{
		WatchPaths:     paths,
		Interval:       time.Second,
		DedupeWindow:   time.Minute,
		CooldownPeriod: time.Millisecond,
	}
}

func TestFromConfig_ReturnsNonNilPipeline(t *testing.T) {
	cfg := minimalConfig([]string{"/etc/hosts"})
	p := pipeline.FromConfig(cfg, nil)
	if p == nil {
		t.Fatal("expected non-nil Pipeline")
	}
}

func TestFromConfig_NotifiesOnDrift(t *testing.T) {
	var sink bytes.Buffer
	cfg := minimalConfig([]string{"/etc/hosts"})
	p := pipeline.FromConfig(cfg, &sink)

	results := []watcher.DriftResult{
		{Path: "/etc/hosts", Drifted: true, CurrentSum: "cafebabe"},
	}
	p.Process(context.Background(), results)

	if !strings.Contains(sink.String(), "/etc/hosts") {
		t.Errorf("expected /etc/hosts in output, got %q", sink.String())
	}
}

func TestFromConfig_ExcludePatternFiltersPath(t *testing.T) {
	var sink bytes.Buffer
	cfg := minimalConfig([]string{"/etc/hosts"})
	cfg.ExcludePatterns = []string{"/etc/*"}
	p := pipeline.FromConfig(cfg, &sink)

	results := []watcher.DriftResult{
		{Path: "/etc/hosts", Drifted: true, CurrentSum: "cafebabe"},
	}
	p.Process(context.Background(), results)

	if sink.Len() != 0 {
		t.Errorf("excluded path should produce no output, got %q", sink.String())
	}
}

func TestFromConfig_NilLogSinkDefaultsToStderr(t *testing.T) {
	// Just ensure no panic when logSink is nil.
	cfg := minimalConfig([]string{"/tmp"})
	pipeline.FromConfig(cfg, nil)
}
