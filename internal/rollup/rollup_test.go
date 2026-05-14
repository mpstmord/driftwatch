package rollup_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/rollup"
	"github.com/example/driftwatch/internal/watcher"
)

func makeResult(path string) watcher.DriftResult {
	return watcher.DriftResult{Path: path, Drifted: true}
}

func TestAdd_BuffersResults(t *testing.T) {
	r := rollup.New(time.Hour, func(_ []watcher.DriftResult) {})
	r.Add(makeResult("/etc/hosts"))
	r.Add(makeResult("/etc/resolv.conf"))
	if got := r.Len(); got != 2 {
		t.Fatalf("expected 2 buffered results, got %d", got)
	}
}

func TestRun_FlushesOnWindow(t *testing.T) {
	var mu sync.Mutex
	var received []watcher.DriftResult

	r := rollup.New(20*time.Millisecond, func(batch []watcher.DriftResult) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, batch...)
	})

	r.Add(makeResult("/etc/hosts"))
	r.Add(makeResult("/etc/nginx/nginx.conf"))

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if len(received) < 2 {
		t.Fatalf("expected at least 2 flushed results, got %d", len(received))
	}
}

func TestRun_NoFlushWhenEmpty(t *testing.T) {
	called := 0
	r := rollup.New(20*time.Millisecond, func(batch []watcher.DriftResult) {
		called++
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	r.Run(ctx)
	if called != 0 {
		t.Fatalf("handler should not be called when buffer is empty, called %d times", called)
	}
}

func TestRun_CancelFlushesRemainder(t *testing.T) {
	var mu sync.Mutex
	var received []watcher.DriftResult

	r := rollup.New(time.Hour, func(batch []watcher.DriftResult) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, batch...)
	})

	r.Add(makeResult("/etc/ssh/sshd_config"))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 {
		t.Fatalf("expected 1 result flushed on cancel, got %d", len(received))
	}
}
