package scheduler_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/alert"
	"github.com/example/driftwatch/internal/scheduler"
	"github.com/example/driftwatch/internal/watcher"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "sched-*.txt")
	if err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writeTempFile write: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestRun_CancelStopsLoop(t *testing.T) {
	path := writeTempFile(t, "stable content")

	w, err := watcher.New([]string{path})
	if err != nil {
		t.Fatalf("watcher.New: %v", err)
	}

	var buf strings.Builder
	a := alert.New(&buf)
	h := alert.NewDriftHandler(a)
	s := scheduler.New(50*time.Millisecond, w, h)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Millisecond)
	defer cancel()

	err = s.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestRun_AlertsOnDrift(t *testing.T) {
	path := writeTempFile(t, "original")

	w, err := watcher.New([]string{path})
	if err != nil {
		t.Fatalf("watcher.New: %v", err)
	}

	// Mutate the file after the watcher has recorded the baseline.
	if err := os.WriteFile(path, []byte("changed"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var buf strings.Builder
	a := alert.New(&buf)
	h := alert.NewDriftHandler(a)
	s := scheduler.New(50*time.Millisecond, w, h)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	s.Run(ctx) //nolint:errcheck

	if !strings.Contains(buf.String(), "drift") {
		t.Errorf("expected drift alert in output, got: %q", buf.String())
	}
}
