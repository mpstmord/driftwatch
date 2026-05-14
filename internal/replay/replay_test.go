package replay_test

import (
	"context"
	"testing"
	"time"

	"github.com/driftwatch/internal/eventlog"
	"github.com/driftwatch/internal/replay"
	"github.com/driftwatch/internal/watcher"
)

func driftedResult(path string) watcher.DriftResult {
	return watcher.DriftResult{Path: path, Drifted: true}
}

func populatedLog(t *testing.T, paths ...string) *eventlog.EventLog {
	t.Helper()
	el := eventlog.New(50)
	for _, p := range paths {
		el.Record(driftedResult(p))
	}
	return el
}

func TestRun_DeliversAllEvents(t *testing.T) {
	el := populatedLog(t, "/etc/a", "/etc/b", "/etc/c")
	r := replay.New(el, replay.Options{})

	var got []string
	err := r.Run(context.Background(), func(res watcher.DriftResult) {
		got = append(got, res.Path)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 events, got %d", len(got))
	}
}

func TestRun_SinceFiltersOldEvents(t *testing.T) {
	el := eventlog.New(50)
	el.Record(driftedResult("/etc/old"))
	cutoff := time.Now()
	time.Sleep(2 * time.Millisecond)
	el.Record(driftedResult("/etc/new"))

	r := replay.New(el, replay.Options{Since: cutoff})

	var got []string
	_ = r.Run(context.Background(), func(res watcher.DriftResult) {
		got = append(got, res.Path)
	})
	if len(got) != 1 || got[0] != "/etc/new" {
		t.Fatalf("expected only /etc/new, got %v", got)
	}
}

func TestRun_CancelledContextStopsEarly(t *testing.T) {
	el := populatedLog(t, "/etc/a", "/etc/b", "/etc/c", "/etc/d", "/etc/e")
	r := replay.New(el, replay.Options{Delay: 10 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	var count int
	_ = r.Run(ctx, func(res watcher.DriftResult) {
		count++
		if count == 2 {
			cancel()
		}
	})
	if count >= 5 {
		t.Fatal("expected replay to stop early after cancel")
	}
}

func TestRun_EmptyLogProducesNoEvents(t *testing.T) {
	el := eventlog.New(50)
	r := replay.New(el, replay.Options{})

	var count int
	err := r.Run(context.Background(), func(_ watcher.DriftResult) { count++ })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 events, got %d", count)
	}
}
