package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "driftwatch.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempConfig: %v", err)
	}
	return p
}

func TestLoad_ValidConfig(t *testing.T) {
	raw := `
interval: 10s
alert:
  level: error
  output: stdout
files:
  - path: /etc/hosts
    tag: hosts
`
	p := writeTempConfig(t, raw)
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 10*time.Second {
		t.Errorf("interval: got %v, want 10s", cfg.Interval)
	}
	if cfg.Alert.Level != "error" {
		t.Errorf("alert.level: got %q, want \"error\"", cfg.Alert.Level)
	}
	if len(cfg.Files) != 1 || cfg.Files[0].Path != "/etc/hosts" {
		t.Errorf("unexpected files: %+v", cfg.Files)
	}
}

func TestLoad_Defaults(t *testing.T) {
	raw := `
files:
  - path: /etc/passwd
`
	p := writeTempConfig(t, raw)
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 30*time.Second {
		t.Errorf("default interval: got %v, want 30s", cfg.Interval)
	}
	if cfg.Alert.Level != "warn" {
		t.Errorf("default level: got %q, want \"warn\"", cfg.Alert.Level)
	}
	if cfg.Alert.Output != "stdout" {
		t.Errorf("default output: got %q, want \"stdout\"", cfg.Alert.Output)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/driftwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_NoFilesError(t *testing.T) {
	raw := `interval: 5s\nalert:\n  level: info\n`
	p := writeTempConfig(t, "interval: 5s\n")
	_ = p
	p2 := writeTempConfig(t, "interval: 5s\nfiles: []\n")
	_, err := config.Load(p2)
	if err == nil {
		t.Fatal("expected validation error for empty files list")
	}
}

func TestLoad_EmptyPathInEntry(t *testing.T) {
	raw := `
files:
  - path: ""
    tag: empty
`
	p := writeTempConfig(t, raw)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error for empty file path")
	}
}
