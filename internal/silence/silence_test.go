package silence

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsSilenced_NotSilencedByDefault(t *testing.T) {
	s := New()
	if s.IsSilenced("/etc/app.conf") {
		t.Fatal("expected path to not be silenced")
	}
}

func TestAdd_SilencesPath(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)
	s.Add("/etc/app.conf", 5*time.Minute)
	if !s.IsSilenced("/etc/app.conf") {
		t.Fatal("expected path to be silenced")
	}
}

func TestIsSilenced_ExpiredAfterDuration(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)
	s.Add("/etc/app.conf", 1*time.Minute)

	// advance clock past expiry
	s.now = fixedClock(now.Add(2 * time.Minute))
	if s.IsSilenced("/etc/app.conf") {
		t.Fatal("expected silence to have expired")
	}
}

func TestLift_RemovesSilence(t *testing.T) {
	s := New()
	s.Add("/etc/app.conf", 10*time.Minute)
	s.Lift("/etc/app.conf")
	if s.IsSilenced("/etc/app.conf") {
		t.Fatal("expected silence to be lifted")
	}
}

func TestLift_NoopOnUnknownPath(t *testing.T) {
	s := New()
	s.Lift("/nonexistent") // must not panic
}

func TestAdd_ReplacesExistingExpiry(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)
	s.Add("/etc/app.conf", 1*time.Minute)
	s.Add("/etc/app.conf", 10*time.Minute)

	// advance past first expiry but within second
	s.now = fixedClock(now.Add(2 * time.Minute))
	if !s.IsSilenced("/etc/app.conf") {
		t.Fatal("expected extended silence to still be active")
	}
}

func TestActive_ReturnsOnlyLiveSilences(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)
	s.Add("/etc/a.conf", 5*time.Minute)
	s.Add("/etc/b.conf", 1*time.Minute)

	// advance past b's expiry
	s.now = fixedClock(now.Add(2 * time.Minute))
	active := s.Active()
	if _, ok := active["/etc/a.conf"]; !ok {
		t.Error("expected /etc/a.conf to be active")
	}
	if _, ok := active["/etc/b.conf"]; ok {
		t.Error("expected /etc/b.conf to have expired")
	}
}

func TestActive_ReturnsDurationRemaining(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)
	s.Add("/etc/app.conf", 10*time.Minute)

	s.now = fixedClock(now.Add(3 * time.Minute))
	active := s.Active()
	remaining := active["/etc/app.conf"]
	if remaining < 6*time.Minute || remaining > 7*time.Minute {
		t.Errorf("unexpected remaining duration: %v", remaining)
	}
}
