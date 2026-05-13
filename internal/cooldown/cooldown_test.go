package cooldown_test

import (
	"testing"
	"time"

	"github.com/example/driftwatch/internal/cooldown"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestReady_FirstCallAlwaysPasses(t *testing.T) {
	c := cooldown.New(5 * time.Second)
	if !c.Ready("file.conf") {
		t.Fatal("expected first call to be ready")
	}
}

func TestReady_SecondCallWithinCooldownBlocked(t *testing.T) {
	now := time.Now()
	c := cooldown.New(10 * time.Second)
	// Inject fixed clock via unexported field workaround: use the public API only.
	// We call Ready twice in quick succession; the second should be blocked.
	c.Ready("cfg") // first call records now
	if c.Ready("cfg") {
		t.Fatal("expected second call within cooldown to be blocked")
	}
	_ = now
}

func TestReady_PassesAfterCooldownExpires(t *testing.T) {
	c := cooldown.New(1 * time.Millisecond)
	c.Ready("cfg")
	time.Sleep(5 * time.Millisecond)
	if !c.Ready("cfg") {
		t.Fatal("expected call after cooldown to pass")
	}
}

func TestReady_DifferentKeysAreIndependent(t *testing.T) {
	c := cooldown.New(10 * time.Second)
	c.Ready("a")
	if !c.Ready("b") {
		t.Fatal("expected independent key 'b' to be ready")
	}
}

func TestReset_ClearsKeyState(t *testing.T) {
	c := cooldown.New(10 * time.Second)
	c.Ready("cfg") // records timestamp
	if c.Ready("cfg") {
		t.Fatal("expected key to be in cooldown before reset")
	}
	c.Reset("cfg")
	if !c.Ready("cfg") {
		t.Fatal("expected key to be ready after reset")
	}
}

func TestRemaining_ZeroWhenNotSeen(t *testing.T) {
	c := cooldown.New(5 * time.Second)
	if r := c.Remaining("unknown"); r != 0 {
		t.Fatalf("expected 0 remaining for unseen key, got %v", r)
	}
}

func TestRemaining_PositiveWithinCooldown(t *testing.T) {
	c := cooldown.New(10 * time.Second)
	c.Ready("cfg")
	r := c.Remaining("cfg")
	if r <= 0 || r > 10*time.Second {
		t.Fatalf("expected remaining within (0, 10s], got %v", r)
	}
}

func TestRemaining_ZeroAfterExpiry(t *testing.T) {
	c := cooldown.New(1 * time.Millisecond)
	c.Ready("cfg")
	time.Sleep(5 * time.Millisecond)
	if r := c.Remaining("cfg"); r != 0 {
		t.Fatalf("expected 0 remaining after expiry, got %v", r)
	}
}
