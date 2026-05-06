package audit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/driftwatch/driftwatch/internal/audit"
)

func tempLogPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "audit.json")
}

func TestNew_CreatesEmptyLog(t *testing.T) {
	log, err := audit.New(tempLogPath(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries := log.Entries(); len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestRecord_AppendsEntry(t *testing.T) {
	log, _ := audit.New(tempLogPath(t))

	if err := log.Record("/etc/app.conf", true, "checksum mismatch"); err != nil {
		t.Fatalf("Record failed: %v", err)
	}

	entries := log.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.FilePath != "/etc/app.conf" {
		t.Errorf("expected file path /etc/app.conf, got %s", e.FilePath)
	}
	if !e.Drifted {
		t.Error("expected Drifted=true")
	}
	if e.Message != "checksum mismatch" {
		t.Errorf("unexpected message: %s", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestRecord_PersistsToDisk(t *testing.T) {
	path := tempLogPath(t)
	log, _ := audit.New(path)
	log.Record("/etc/hosts", false, "")

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}
}

func TestNew_LoadsExistingEntries(t *testing.T) {
	path := tempLogPath(t)

	// Write two entries with the first instance.
	log1, _ := audit.New(path)
	log1.Record("/etc/ssh/sshd_config", true, "drift detected")
	log1.Record("/etc/hosts", false, "")

	// Re-open and verify entries are reloaded.
	log2, err := audit.New(path)
	if err != nil {
		t.Fatalf("failed to reload log: %v", err)
	}
	entries := log2.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after reload, got %d", len(entries))
	}
	if entries[0].FilePath != "/etc/ssh/sshd_config" {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
}

func TestRecord_MultipleEntries(t *testing.T) {
	log, _ := audit.New(tempLogPath(t))
	paths := []string{"/a", "/b", "/c"}
	for _, p := range paths {
		if err := log.Record(p, false, ""); err != nil {
			t.Fatalf("Record(%s) failed: %v", p, err)
		}
	}
	if got := len(log.Entries()); got != 3 {
		t.Errorf("expected 3 entries, got %d", got)
	}
}
