package retention_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/retention"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAdd_StoresEntry(t *testing.T) {
	s := retention.New(retention.Policy{MaxAge: time.Hour})
	s.Add("/etc/hosts", "abc123")
	if s.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", s.Len())
	}
}

func TestEvict_RemovesExpiredEntries(t *testing.T) {
	now := time.Now()
	s := retention.New(retention.Policy{MaxAge: time.Minute})
	s.SetClock(fixedClock(now.Add(-2 * time.Minute)))
	s.Add("/etc/hosts", "old")
	s.SetClock(fixedClock(now))
	s.Add("/etc/resolv.conf", "new")

	removed := s.Evict()
	if removed != 1 {
		t.Fatalf("expected 1 removed, got %d", removed)
	}
	if s.Len() != 1 {
		t.Fatalf("expected 1 remaining, got %d", s.Len())
	}
}

func TestEvict_KeepsAllWhenNoneExpired(t *testing.T) {
	s := retention.New(retention.Policy{MaxAge: time.Hour})
	s.Add("/a", "1")
	s.Add("/b", "2")
	removed := s.Evict()
	if removed != 0 {
		t.Fatalf("expected 0 removed, got %d", removed)
	}
	if s.Len() != 2 {
		t.Fatalf("expected 2 remaining, got %d", s.Len())
	}
}

func TestEvict_RemovesAllWhenAllExpired(t *testing.T) {
	now := time.Now()
	s := retention.New(retention.Policy{MaxAge: time.Second})
	s.SetClock(fixedClock(now.Add(-10 * time.Second)))
	s.Add("/a", "x")
	s.Add("/b", "y")
	s.SetClock(fixedClock(now))
	removed := s.Evict()
	if removed != 2 {
		t.Fatalf("expected 2 removed, got %d", removed)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := retention.New(retention.Policy{MaxAge: time.Hour})
	s.Add("/etc/hosts", "abc")
	entries := s.All()
	entries[0].Path = "mutated"
	if s.All()[0].Path != "/etc/hosts" {
		t.Fatal("All() should return a copy, not a reference")
	}
}

func TestAll_ReflectsCorrectFields(t *testing.T) {
	now := time.Now()
	s := retention.New(retention.Policy{MaxAge: time.Hour})
	s.SetClock(fixedClock(now))
	s.Add("/etc/passwd", "deadbeef")
	entries := s.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Path != "/etc/passwd" {
		t.Errorf("unexpected path: %s", e.Path)
	}
	if e.Checksum != "deadbeef" {
		t.Errorf("unexpected checksum: %s", e.Checksum)
	}
	if !e.RecordedAt.Equal(now) {
		t.Errorf("unexpected timestamp: %v", e.RecordedAt)
	}
}
