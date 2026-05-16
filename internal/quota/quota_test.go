package quota

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_FirstCallAlwaysPasses(t *testing.T) {
	q := New(3, time.Minute)
	if !q.Allow("/etc/hosts") {
		t.Fatal("expected first call to pass")
	}
}

func TestAllow_RespectsMax(t *testing.T) {
	q := New(2, time.Minute)
	path := "/etc/hosts"
	if !q.Allow(path) {
		t.Fatal("first call should pass")
	}
	if !q.Allow(path) {
		t.Fatal("second call should pass")
	}
	if q.Allow(path) {
		t.Fatal("third call should be blocked")
	}
}

func TestAllow_DifferentPathsAreIndependent(t *testing.T) {
	q := New(1, time.Minute)
	if !q.Allow("/etc/hosts") {
		t.Fatal("first path should pass")
	}
	if !q.Allow("/etc/resolv.conf") {
		t.Fatal("second path should pass independently")
	}
}

func TestAllow_PassesAfterWindowExpires(t *testing.T) {
	now := time.Now()
	q := New(1, time.Second)
	q.now = fixedClock(now)
	path := "/etc/hosts"
	q.Allow(path)
	q.Allow(path) // exhaust

	// advance past the window
	q.now = fixedClock(now.Add(2 * time.Second))
	if !q.Allow(path) {
		t.Fatal("expected allow after window expiry")
	}
}

func TestRemaining_ReturnsMaxWhenUnused(t *testing.T) {
	q := New(5, time.Minute)
	if got := q.Remaining("/etc/hosts"); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestRemaining_DecreasesAfterAllow(t *testing.T) {
	q := New(3, time.Minute)
	path := "/etc/hosts"
	q.Allow(path)
	if got := q.Remaining(path); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestRemaining_ZeroWhenExhausted(t *testing.T) {
	q := New(1, time.Minute)
	path := "/etc/hosts"
	q.Allow(path)
	q.Allow(path)
	if got := q.Remaining(path); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestReset_ClearsState(t *testing.T) {
	q := New(1, time.Minute)
	path := "/etc/hosts"
	q.Allow(path)
	q.Allow(path) // exhaust
	q.Reset(path)
	if !q.Allow(path) {
		t.Fatal("expected allow after reset")
	}
}
