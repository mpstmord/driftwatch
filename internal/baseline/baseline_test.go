package baseline_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/baseline"
)

func TestNew_EmptyRecords(t *testing.T) {
	b := baseline.New()
	if b.Records == nil {
		t.Fatal("expected non-nil Records map")
	}
	if len(b.Records) != 0 {
		t.Fatalf("expected 0 records, got %d", len(b.Records))
	}
}

func TestSet_And_Get(t *testing.T) {
	b := baseline.New()
	b.Set("/etc/app.conf", "abc123")

	r, ok := b.Get("/etc/app.conf")
	if !ok {
		t.Fatal("expected record to be found")
	}
	if r.Checksum != "abc123" {
		t.Errorf("expected checksum abc123, got %s", r.Checksum)
	}
	if r.RecordedAt.IsZero() {
		t.Error("expected RecordedAt to be set")
	}
	if time.Since(r.RecordedAt) > 5*time.Second {
		t.Error("RecordedAt appears stale")
	}
}

func TestGet_MissingKey(t *testing.T) {
	b := baseline.New()
	_, ok := b.Get("/nonexistent")
	if ok {
		t.Fatal("expected record not to be found")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "baseline.json")

	b := baseline.New()
	b.Set("/etc/nginx.conf", "deadbeef")
	b.Set("/etc/hosts", "cafebabe")

	if err := baseline.Save(b, dest); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := baseline.Load(dest)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(loaded.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(loaded.Records))
	}
	r, ok := loaded.Get("/etc/nginx.conf")
	if !ok || r.Checksum != "deadbeef" {
		t.Errorf("unexpected record for /etc/nginx.conf: %+v", r)
	}
}

func TestLoad_MissingFile_ReturnsErrNoBaseline(t *testing.T) {
	_, err := baseline.Load("/tmp/driftwatch_nonexistent_baseline.json")
	if err != baseline.ErrNoBaseline {
		t.Fatalf("expected ErrNoBaseline, got %v", err)
	}
}

func TestLoad_CorruptFile_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(dest, []byte("not json{"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := baseline.Load(dest)
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}
}
