package quota

import (
	"context"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/watcher"
)

func driftResult(path string) watcher.DriftResult {
	return watcher.DriftResult{Path: path, Drifted: true}
}

func TestMiddleware_ForwardsWhenAllowed(t *testing.T) {
	q := New(2, time.Minute)
	var received []watcher.DriftResult
	mw := NewMiddleware(q, func(_ context.Context, r watcher.DriftResult) {
		received = append(received, r)
	})

	mw.Handle(context.Background(), driftResult("/etc/hosts"))
	if len(received) != 1 {
		t.Fatalf("expected 1 forwarded result, got %d", len(received))
	}
}

func TestMiddleware_DropsWhenQuotaExceeded(t *testing.T) {
	q := New(1, time.Minute)
	var count int
	mw := NewMiddleware(q, func(_ context.Context, _ watcher.DriftResult) {
		count++
	})

	path := "/etc/hosts"
	mw.Handle(context.Background(), driftResult(path))
	mw.Handle(context.Background(), driftResult(path))
	mw.Handle(context.Background(), driftResult(path))

	if count != 1 {
		t.Fatalf("expected 1 forwarded result, got %d", count)
	}
}

func TestMiddleware_DifferentPathsTrackedSeparately(t *testing.T) {
	q := New(1, time.Minute)
	var count int
	mw := NewMiddleware(q, func(_ context.Context, _ watcher.DriftResult) {
		count++
	})

	mw.Handle(context.Background(), driftResult("/etc/hosts"))
	mw.Handle(context.Background(), driftResult("/etc/resolv.conf"))

	if count != 2 {
		t.Fatalf("expected 2 forwarded results, got %d", count)
	}
}

func TestMiddleware_AllowsAgainAfterWindowExpiry(t *testing.T) {
	now := time.Now()
	q := New(1, time.Second)
	q.now = fixedClock(now)

	var count int
	mw := NewMiddleware(q, func(_ context.Context, _ watcher.DriftResult) {
		count++
	})

	path := "/etc/hosts"
	mw.Handle(context.Background(), driftResult(path))
	mw.Handle(context.Background(), driftResult(path)) // blocked

	q.now = fixedClock(now.Add(2 * time.Second))
	mw.Handle(context.Background(), driftResult(path)) // new window

	if count != 2 {
		t.Fatalf("expected 2 forwarded results, got %d", count)
	}
}
