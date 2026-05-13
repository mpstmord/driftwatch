package dedupe

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsDuplicate_FirstCallAlwaysPasses(t *testing.T) {
	d := New(5 * time.Minute)
	if d.IsDuplicate("/etc/app.conf", "abc123") {
		t.Fatal("expected first call to not be a duplicate")
	}
}

func TestIsDuplicate_SameHashWithinWindowIsBlocked(t *testing.T) {
	now := time.Now()
	d := New(5 * time.Minute)
	d.now = fixedClock(now)

	d.IsDuplicate("/etc/app.conf", "abc123")
	if !d.IsDuplicate("/etc/app.conf", "abc123") {
		t.Fatal("expected second call with same hash within window to be duplicate")
	}
}

func TestIsDuplicate_DifferentHashIsNotBlocked(t *testing.T) {
	now := time.Now()
	d := New(5 * time.Minute)
	d.now = fixedClock(now)

	d.IsDuplicate("/etc/app.conf", "abc123")
	if d.IsDuplicate("/etc/app.conf", "xyz999") {
		t.Fatal("expected different hash to not be a duplicate")
	}
}

func TestIsDuplicate_PassesAfterWindowExpires(t *testing.T) {
	base := time.Now()
	d := New(1 * time.Second)
	d.now = fixedClock(base)

	d.IsDuplicate("/etc/app.conf", "abc123")

	// Advance clock past the window.
	d.now = fixedClock(base.Add(2 * time.Second))
	if d.IsDuplicate("/etc/app.conf", "abc123") {
		t.Fatal("expected event to pass after window expires")
	}
}

func TestIsDuplicate_DifferentPathsAreIndependent(t *testing.T) {
	now := time.Now()
	d := New(5 * time.Minute)
	d.now = fixedClock(now)

	d.IsDuplicate("/etc/a.conf", "hash1")
	if d.IsDuplicate("/etc/b.conf", "hash1") {
		t.Fatal("expected different paths to be independent")
	}
}

func TestForget_AllowsNextEventThrough(t *testing.T) {
	now := time.Now()
	d := New(5 * time.Minute)
	d.now = fixedClock(now)

	d.IsDuplicate("/etc/app.conf", "abc123")
	d.Forget("/etc/app.conf")

	if d.IsDuplicate("/etc/app.conf", "abc123") {
		t.Fatal("expected event to pass after Forget")
	}
}

func TestReset_ClearsAllEntries(t *testing.T) {
	now := time.Now()
	d := New(5 * time.Minute)
	d.now = fixedClock(now)

	d.IsDuplicate("/etc/a.conf", "h1")
	d.IsDuplicate("/etc/b.conf", "h2")
	d.Reset()

	if d.IsDuplicate("/etc/a.conf", "h1") {
		t.Fatal("expected /etc/a.conf to pass after Reset")
	}
	if d.IsDuplicate("/etc/b.conf", "h2") {
		t.Fatal("expected /etc/b.conf to pass after Reset")
	}
}
