package retention_test

import (
	"context"
	"testing"
	"time"

	"github.com/driftwatch/internal/retention"
)

func TestSweeper_EvictsExpiredEntriesOverTime(t *testing.T) {
	now := time.Now()
	s := retention.New(retention.Policy{MaxAge: 50 * time.Millisecond})
	s.SetClock(fixedClock(now.Add(-200 * time.Millisecond)))
	s.Add("/old", "stale")
	s.SetClock(fixedClock(now))
	s.Add("/new", "fresh")

	sw := retention.NewSweeper(s, 20*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	go sw.Run(ctx)
	<-ctx.Done()

	if s.Len() != 1 {
		t.Fatalf("expected 1 entry after sweep, got %d", s.Len())
	}
	if s.All()[0].Path != "/new" {
		t.Errorf("expected /new to survive, got %s", s.All()[0].Path)
	}
}

func TestSweeper_StopsOnContextCancel(t *testing.T) {
	s := retention.New(retention.Policy{MaxAge: time.Hour})
	sw := retention.NewSweeper(s, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		sw.Run(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
		// expected
	case <-time.After(200 * time.Millisecond):
		t.Fatal("sweeper did not stop after context cancel")
	}
}
