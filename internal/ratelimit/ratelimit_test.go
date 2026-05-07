package ratelimit_test

import (
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/ratelimit"
)

func TestAllow_FirstEventAlwaysPasses(t *testing.T) {
	l := ratelimit.New(time.Second, 3)
	if !l.Allow("/etc/hosts") {
		t.Fatal("expected first event to be allowed")
	}
}

func TestAllow_RespectsMaxWithinWindow(t *testing.T) {
	l := ratelimit.New(time.Second, 2)
	key := "/etc/hosts"

	if !l.Allow(key) {
		t.Fatal("expected first allow")
	}
	if !l.Allow(key) {
		t.Fatal("expected second allow")
	}
	if l.Allow(key) {
		t.Fatal("expected third call to be blocked")
	}
}

func TestAllow_DifferentKeysAreIndependent(t *testing.T) {
	l := ratelimit.New(time.Second, 1)

	if !l.Allow("/etc/hosts") {
		t.Fatal("expected /etc/hosts to be allowed")
	}
	if !l.Allow("/etc/resolv.conf") {
		t.Fatal("expected /etc/resolv.conf to be allowed independently")
	}
}

func TestAllow_PassesAfterWindowExpires(t *testing.T) {
	window := 50 * time.Millisecond
	l := ratelimit.New(window, 1)
	key := "/etc/passwd"

	if !l.Allow(key) {
		t.Fatal("expected first allow")
	}
	if l.Allow(key) {
		t.Fatal("expected second call within window to be blocked")
	}

	time.Sleep(window + 10*time.Millisecond)

	if !l.Allow(key) {
		t.Fatal("expected allow after window expiry")
	}
}

func TestReset_ClearsKeyState(t *testing.T) {
	l := ratelimit.New(time.Second, 1)
	key := "/etc/hosts"

	l.Allow(key)
	if l.Allow(key) {
		t.Fatal("expected second call to be blocked before reset")
	}

	l.Reset(key)

	if !l.Allow(key) {
		t.Fatal("expected allow after reset")
	}
}

func TestRemaining_ReflectsAvailableSlots(t *testing.T) {
	l := ratelimit.New(time.Second, 3)
	key := "/etc/ssh/sshd_config"

	if got := l.Remaining(key); got != 3 {
		t.Fatalf("expected 3 remaining, got %d", got)
	}

	l.Allow(key)

	if got := l.Remaining(key); got != 2 {
		t.Fatalf("expected 2 remaining after one event, got %d", got)
	}
}
