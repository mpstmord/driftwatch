package acknowledge

import (
	"testing"
	"time"
)

// fixedClock returns a clock function whose value can be advanced.
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsAcknowledged_NotAcknowledgedByDefault(t *testing.T) {
	s := New()
	if s.IsAcknowledged("/etc/app.conf") {
		t.Fatal("expected path to be unacknowledged initially")
	}
}

func TestAcknowledge_ActiveWithinDuration(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)

	s.Acknowledge("/etc/app.conf", "planned change", 10*time.Minute)

	if !s.IsAcknowledged("/etc/app.conf") {
		t.Fatal("expected path to be acknowledged")
	}
}

func TestAcknowledge_ExpiredAfterDuration(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)

	s.Acknowledge("/etc/app.conf", "planned change", 5*time.Minute)

	// Advance clock past expiry.
	s.now = fixedClock(now.Add(6 * time.Minute))

	if s.IsAcknowledged("/etc/app.conf") {
		t.Fatal("expected acknowledgement to have expired")
	}
}

func TestLift_RemovesAcknowledgement(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)

	s.Acknowledge("/etc/app.conf", "ops ticket #42", 1*time.Hour)
	s.Lift("/etc/app.conf")

	if s.IsAcknowledged("/etc/app.conf") {
		t.Fatal("expected acknowledgement to be lifted")
	}
}

func TestReason_ReturnsReasonForActivePath(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)

	s.Acknowledge("/etc/nginx.conf", "scheduled deploy", 30*time.Minute)

	reason, ok := s.Reason("/etc/nginx.conf")
	if !ok {
		t.Fatal("expected reason to be present")
	}
	if reason != "scheduled deploy" {
		t.Fatalf("expected 'scheduled deploy', got %q", reason)
	}
}

func TestReason_MissingPathReturnsFalse(t *testing.T) {
	s := New()
	_, ok := s.Reason("/etc/missing.conf")
	if ok {
		t.Fatal("expected no reason for unknown path")
	}
}

func TestAcknowledge_ResetsExpiry(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)

	s.Acknowledge("/etc/app.conf", "first", 2*time.Minute)

	// Advance just past original expiry, then re-acknowledge.
	s.now = fixedClock(now.Add(3 * time.Minute))
	s.Acknowledge("/etc/app.conf", "renewed", 10*time.Minute)

	if !s.IsAcknowledged("/etc/app.conf") {
		t.Fatal("expected renewed acknowledgement to be active")
	}

	reason, _ := s.Reason("/etc/app.conf")
	if reason != "renewed" {
		t.Fatalf("expected reason 'renewed', got %q", reason)
	}
}
