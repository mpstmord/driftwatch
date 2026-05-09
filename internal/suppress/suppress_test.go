package suppress

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsSuppressed_NotSuppressedByDefault(t *testing.T) {
	m := New()
	if m.IsSuppressed("/etc/app.conf") {
		t.Fatal("expected path to not be suppressed")
	}
}

func TestSuppress_ActiveWithinDuration(t *testing.T) {
	base := time.Now()
	m := New()
	m.now = fixedNow(base)

	m.Suppress("/etc/app.conf", 5*time.Minute)

	m.now = fixedNow(base.Add(4 * time.Minute))
	if !m.IsSuppressed("/etc/app.conf") {
		t.Fatal("expected path to be suppressed within window")
	}
}

func TestSuppress_ExpiredAfterDuration(t *testing.T) {
	base := time.Now()
	m := New()
	m.now = fixedNow(base)

	m.Suppress("/etc/app.conf", 5*time.Minute)

	m.now = fixedNow(base.Add(6 * time.Minute))
	if m.IsSuppressed("/etc/app.conf") {
		t.Fatal("expected suppression to have expired")
	}
}

func TestLift_RemovesSuppression(t *testing.T) {
	m := New()
	m.Suppress("/etc/app.conf", time.Hour)
	m.Lift("/etc/app.conf")

	if m.IsSuppressed("/etc/app.conf") {
		t.Fatal("expected suppression to be lifted")
	}
}

func TestPurge_RemovesExpiredOnly(t *testing.T) {
	base := time.Now()
	m := New()
	m.now = fixedNow(base)

	m.Suppress("/etc/a.conf", time.Minute)
	m.Suppress("/etc/b.conf", time.Hour)

	m.now = fixedNow(base.Add(2 * time.Minute))
	m.Purge()

	if m.IsSuppressed("/etc/a.conf") {
		t.Error("expected /etc/a.conf to be purged")
	}
	if !m.IsSuppressed("/etc/b.conf") {
		t.Error("expected /etc/b.conf to remain active")
	}
}

func TestActive_ReturnsOnlyLiveSuppresions(t *testing.T) {
	base := time.Now()
	m := New()
	m.now = fixedNow(base)

	m.Suppress("/etc/a.conf", time.Minute)
	m.Suppress("/etc/b.conf", time.Hour)

	m.now = fixedNow(base.Add(2 * time.Minute))
	active := m.Active()

	if len(active) != 1 {
		t.Fatalf("expected 1 active suppression, got %d", len(active))
	}
	if active[0].Path != "/etc/b.conf" {
		t.Errorf("expected /etc/b.conf, got %s", active[0].Path)
	}
}
