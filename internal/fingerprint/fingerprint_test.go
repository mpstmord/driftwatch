package fingerprint_test

import (
	"testing"
	"time"

	"github.com/example/driftwatch/internal/fingerprint"
)

func TestUpdate_FirstCallReturnsFalse(t *testing.T) {
	tr := fingerprint.New()
	changed := tr.Update("/etc/hosts", "abc123")
	if changed {
		t.Fatal("first observation should never be reported as a change")
	}
}

func TestUpdate_SameChecksumReturnsFalse(t *testing.T) {
	tr := fingerprint.New()
	tr.Update("/etc/hosts", "abc123")
	changed := tr.Update("/etc/hosts", "abc123")
	if changed {
		t.Fatal("identical checksum should not be reported as a change")
	}
}

func TestUpdate_DifferentChecksumReturnsTrue(t *testing.T) {
	tr := fingerprint.New()
	tr.Update("/etc/hosts", "abc123")
	changed := tr.Update("/etc/hosts", "def456")
	if !changed {
		t.Fatal("different checksum should be reported as a change")
	}
}

func TestGet_ReturnsStoredRecord(t *testing.T) {
	tr := fingerprint.New()
	before := time.Now()
	tr.Update("/etc/passwd", "sum1")
	after := time.Now()

	rec, ok := tr.Get("/etc/passwd")
	if !ok {
		t.Fatal("expected record to exist")
	}
	if rec.Checksum != "sum1" {
		t.Fatalf("unexpected checksum: %q", rec.Checksum)
	}
	if rec.RecordedAt.Before(before) || rec.RecordedAt.After(after) {
		t.Fatalf("RecordedAt %v outside expected range", rec.RecordedAt)
	}
}

func TestGet_MissingPathReturnsFalse(t *testing.T) {
	tr := fingerprint.New()
	_, ok := tr.Get("/nonexistent")
	if ok {
		t.Fatal("expected no record for unseen path")
	}
}

func TestDelete_RemovesRecord(t *testing.T) {
	tr := fingerprint.New()
	tr.Update("/etc/hosts", "abc")
	tr.Delete("/etc/hosts")
	_, ok := tr.Get("/etc/hosts")
	if ok {
		t.Fatal("expected record to be deleted")
	}
}

func TestPaths_ReturnsAllTrackedPaths(t *testing.T) {
	tr := fingerprint.New()
	tr.Update("/a", "1")
	tr.Update("/b", "2")
	tr.Update("/c", "3")

	paths := tr.Paths()
	if len(paths) != 3 {
		t.Fatalf("expected 3 paths, got %d", len(paths))
	}
}

func TestUpdate_AfterDelete_TreatedAsFirstObservation(t *testing.T) {
	tr := fingerprint.New()
	tr.Update("/etc/hosts", "abc")
	tr.Delete("/etc/hosts")
	changed := tr.Update("/etc/hosts", "xyz")
	if changed {
		t.Fatal("re-added path should be treated as first observation")
	}
}
