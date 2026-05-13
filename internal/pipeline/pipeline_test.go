package pipeline_test

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/cooldown"
	"github.com/example/driftwatch/internal/dedupe"
	"github.com/example/driftwatch/internal/filter"
	"github.com/example/driftwatch/internal/notifier"
	"github.com/example/driftwatch/internal/pipeline"
	"github.com/example/driftwatch/internal/watcher"
)

func newPipeline(sink *bytes.Buffer) *pipeline.Pipeline {
	f := filter.New(filter.Config{})
	d := dedupe.New(dedupe.Config{Window: time.Minute})
	c := cooldown.New(cooldown.Config{Period: time.Millisecond})
	n := notifier.New(notifier.Config{Sink: sink})
	return pipeline.New(pipeline.Config{
		Filter:   f,
		Dedupe:   d,
		Cooldown: c,
		Notifier: n,
		Logger:   log.New(bytes.NewBuffer(nil), "", 0),
	})
}

func TestProcess_NoDriftProducesNoOutput(t *testing.T) {
	var buf bytes.Buffer
	p := newPipeline(&buf)
	results := []watcher.DriftResult{{Path: "/etc/hosts", Drifted: false}}
	p.Process(context.Background(), results)
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}

func TestProcess_DriftedFileNotifiesOnce(t *testing.T) {
	var buf bytes.Buffer
	p := newPipeline(&buf)
	results := []watcher.DriftResult{
		{Path: "/etc/hosts", Drifted: true, CurrentSum: "abc123"},
	}
	p.Process(context.Background(), results)
	if !strings.Contains(buf.String(), "/etc/hosts") {
		t.Errorf("expected notification for /etc/hosts, got %q", buf.String())
	}
}

func TestProcess_DedupeBlocksRepeat(t *testing.T) {
	var buf bytes.Buffer
	p := newPipeline(&buf)
	r := []watcher.DriftResult{{Path: "/etc/passwd", Drifted: true, CurrentSum: "deadbeef"}}
	p.Process(context.Background(), r)
	first := buf.String()
	buf.Reset()
	p.Process(context.Background(), r)
	if buf.Len() != 0 {
		t.Errorf("dedupe should block second identical event, got %q (first was %q)", buf.String(), first)
	}
}

func TestProcess_CancelledContextStopsEarly(t *testing.T) {
	var buf bytes.Buffer
	p := newPipeline(&buf)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := []watcher.DriftResult{
		{Path: "/a", Drifted: true, CurrentSum: "1"},
		{Path: "/b", Drifted: true, CurrentSum: "2"},
	}
	p.Process(ctx, results)
	// With a pre-cancelled context nothing should be notified.
	if buf.Len() != 0 {
		t.Errorf("cancelled context should produce no output, got %q", buf.String())
	}
}

func TestNew_PanicsOnNilField(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil Config field")
		}
	}()
	pipeline.New(pipeline.Config{}) // all nil → panic
}
