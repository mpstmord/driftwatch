package differ_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/driftwatch/internal/differ"
	"github.com/example/driftwatch/internal/snapshot"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return p
}

func TestDiff_NoDriftWhenSnapshotsMatch(t *testing.T) {
	dir := t.TempDir()
	p := writeTempFile(t, dir, "app.conf", "key=value")

	base, err := snapshot.New([]string{p})
	if err != nil {
		t.Fatalf("snapshot.New: %v", err)
	}
	current, err := snapshot.New([]string{p})
	if err != nil {
		t.Fatalf("snapshot.New: %v", err)
	}

	results := differ.Diff(base, current)
	if len(results) != 0 {
		t.Errorf("expected no drift, got %d results", len(results))
	}
}

func TestDiff_DetectsModifiedFile(t *testing.T) {
	dir := t.TempDir()
	p := writeTempFile(t, dir, "app.conf", "key=original")

	base, err := snapshot.New([]string{p})
	if err != nil {
		t.Fatalf("snapshot.New baseline: %v", err)
	}

	writeTempFile(t, dir, "app.conf", "key=changed")

	current, err := snapshot.New([]string{p})
	if err != nil {
		t.Fatalf("snapshot.New current: %v", err)
	}

	results := differ.Diff(base, current)
	if len(results) != 1 {
		t.Fatalf("expected 1 drift result, got %d", len(results))
	}
	if !results[0].Drifted {
		t.Error("expected Drifted=true")
	}
	if results[0].Path != p {
		t.Errorf("expected path %s, got %s", p, results[0].Path)
	}
}

func TestDiff_DetectsRemovedFile(t *testing.T) {
	dir := t.TempDir()
	p := writeTempFile(t, dir, "removed.conf", "data")

	base, err := snapshot.New([]string{p})
	if err != nil {
		t.Fatalf("snapshot.New: %v", err)
	}

	// Build an empty current snapshot by using a different existing file.
	p2 := writeTempFile(t, dir, "other.conf", "other")
	current, err := snapshot.New([]string{p2})
	if err != nil {
		t.Fatalf("snapshot.New current: %v", err)
	}

	results := differ.Diff(base, current)
	if len(results) == 0 {
		t.Fatal("expected drift results for removed file")
	}
	found := false
	for _, r := range results {
		if r.Path == p && r.Drifted {
			found = true
		}
	}
	if !found {
		t.Errorf("expected drift result for removed file %s", p)
	}
}
