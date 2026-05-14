package escalation

import (
	"testing"
	"time"
)

func fixedClock(at time.Time) func() time.Time {
	return func() time.Time { return at }
}

func TestRecord_FirstHitIsInfo(t *testing.T) {
	tr := New(3, 5, time.Minute)
	level := tr.Record("/etc/hosts")
	if level != LevelInfo {
		t.Fatalf("expected INFO, got %s", level)
	}
}

func TestRecord_ReachesWarnThreshold(t *testing.T) {
	tr := New(3, 5, time.Minute)
	var level Level
	for i := 0; i < 3; i++ {
		level = tr.Record("/etc/hosts")
	}
	if level != LevelWarn {
		t.Fatalf("expected WARN at hit 3, got %s", level)
	}
}

func TestRecord_ReachesCritThreshold(t *testing.T) {
	tr := New(3, 5, time.Minute)
	var level Level
	for i := 0; i < 5; i++ {
		level = tr.Record("/etc/hosts")
	}
	if level != LevelCrit {
		t.Fatalf("expected CRIT at hit 5, got %s", level)
	}
}

func TestRecord_DifferentPathsAreIndependent(t *testing.T) {
	tr := New(2, 4, time.Minute)
	for i := 0; i < 4; i++ {
		tr.Record("/etc/hosts")
	}
	level := tr.Record("/etc/resolv.conf")
	if level != LevelInfo {
		t.Fatalf("expected INFO for fresh path, got %s", level)
	}
}

func TestRecord_WindowResetAfterExpiry(t *testing.T) {
	base := time.Now()
	tr := New(3, 5, time.Minute)
	tr.now = fixedClock(base)

	for i := 0; i < 4; i++ {
		tr.Record("/etc/hosts")
	}

	// Advance clock past the window.
	tr.now = fixedClock(base.Add(2 * time.Minute))
	level := tr.Record("/etc/hosts")
	if level != LevelInfo {
		t.Fatalf("expected INFO after window reset, got %s", level)
	}
}

func TestReset_ClearsCounter(t *testing.T) {
	tr := New(2, 4, time.Minute)
	for i := 0; i < 3; i++ {
		tr.Record("/etc/hosts")
	}
	tr.Reset("/etc/hosts")
	level := tr.Record("/etc/hosts")
	if level != LevelInfo {
		t.Fatalf("expected INFO after reset, got %s", level)
	}
}

func TestLevel_String(t *testing.T) {
	cases := []struct {
		level Level
		want  string
	}{
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelCrit, "CRIT"},
	}
	for _, c := range cases {
		if got := c.level.String(); got != c.want {
			t.Errorf("Level(%d).String() = %q, want %q", c.level, got, c.want)
		}
	}
}
