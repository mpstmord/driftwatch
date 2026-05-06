package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/snapshot"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "snap-*.conf")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestNew_BuildsRecordsForPaths(t *testing.T) {
	p1 := writeTempFile(t, "alpha=1")
	p2 := writeTempFile(t, "beta=2")

	snap, err := snapshot.New([]string{p1, p2})
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	if len(snap.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(snap.Records))
	}
	if snap.Version != 1 {
		t.Errorf("expected version 1, got %d", snap.Version)
	}
	if snap.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestNew_MissingFileReturnsError(t *testing.T) {
	_, err := snapshot.New([]string{"/nonexistent/path/file.conf"})
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	p := writeTempFile(t, "key=value")

	original, err := snapshot.New([]string{p})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	dest := filepath.Join(t.TempDir(), "state.json")
	if err := snapshot.Save(dest, original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := snapshot.Load(dest)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Version != original.Version {
		t.Errorf("version mismatch: got %d, want %d", loaded.Version, original.Version)
	}
	if len(loaded.Records) != len(original.Records) {
		t.Fatalf("record count mismatch: got %d, want %d", len(loaded.Records), len(original.Records))
	}
	if loaded.Records[0].Path != original.Records[0].Path {
		t.Errorf("path mismatch: got %q, want %q", loaded.Records[0].Path, original.Records[0].Path)
	}
	if loaded.Records[0].Checksum.SHA256 != original.Records[0].Checksum.SHA256 {
		t.Error("checksum SHA256 mismatch after round-trip")
	}
}

func TestLoad_MissingFileReturnsError(t *testing.T) {
	_, err := snapshot.Load("/no/such/snapshot.json")
	if err == nil {
		t.Fatal("expected error loading missing snapshot, got nil")
	}
}

func TestNew_RecordCapturedAtIsRecent(t *testing.T) {
	before := time.Now().UTC()
	p := writeTempFile(t, "x=1")
	snap, err := snapshot.New([]string{p})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	after := time.Now().UTC()

	ts := snap.Records[0].CapturedAt
	if ts.Before(before) || ts.After(after) {
		t.Errorf("CapturedAt %v not within expected range [%v, %v]", ts, before, after)
	}
}
