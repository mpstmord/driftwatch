package checksum_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/driftwatch/driftwatch/internal/checksum"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.conf")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestCompute_BasicFields(t *testing.T) {
	path := writeTempFile(t, "key=value\n")

	cs, err := checksum.Compute(path)
	if err != nil {
		t.Fatalf("Compute() unexpected error: %v", err)
	}

	if cs.Path != path {
		t.Errorf("Path = %q; want %q", cs.Path, path)
	}
	if len(cs.SHA256) != 64 {
		t.Errorf("SHA256 length = %d; want 64", len(cs.SHA256))
	}
	if cs.Size != int64(len("key=value\n")) {
		t.Errorf("Size = %d; want %d", cs.Size, len("key=value\n"))
	}
	if cs.ComputedAt.IsZero() {
		t.Error("ComputedAt should not be zero")
	}
}

func TestCompute_Deterministic(t *testing.T) {
	path := writeTempFile(t, "stable content")

	a, _ := checksum.Compute(path)
	b, _ := checksum.Compute(path)

	if a.SHA256 != b.SHA256 {
		t.Errorf("expected identical checksums; got %s and %s", a.SHA256, b.SHA256)
	}
}

func TestCompute_DifferentContent(t *testing.T) {
	p1 := writeTempFile(t, "version=1")
	p2 := writeTempFile(t, "version=2")

	a, _ := checksum.Compute(p1)
	b, _ := checksum.Compute(p2)

	if checksum.Equal(a, b) {
		t.Error("Equal() returned true for files with different content")
	}
}

func TestCompute_MissingFile(t *testing.T) {
	_, err := checksum.Compute("/nonexistent/path/to/file.conf")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestEqual_NilHandling(t *testing.T) {
	path := writeTempFile(t, "data")
	cs, _ := checksum.Compute(path)

	if checksum.Equal(nil, cs) {
		t.Error("Equal(nil, cs) should return false")
	}
	if checksum.Equal(cs, nil) {
		t.Error("Equal(cs, nil) should return false")
	}
	if checksum.Equal(nil, nil) {
		t.Error("Equal(nil, nil) should return false")
	}
}
