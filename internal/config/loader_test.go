package config_test

import (
	"testing"
	"time"

	"github.com/example/driftwatch/internal/config"
)

func TestLoadOrDefault_UsesFileWhenPresent(t *testing.T) {
	raw := `
interval: 60s
files:
  - path: /etc/resolv.conf
    tag: resolv
`
	p := writeTempConfig(t, raw)
	cfg, err := config.LoadOrDefault(p, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 60*time.Second {
		t.Errorf("interval: got %v, want 60s", cfg.Interval)
	}
	if len(cfg.Files) != 1 {
		t.Errorf("expected 1 file entry, got %d", len(cfg.Files))
	}
}

func TestLoadOrDefault_FallsBackToWatchPaths(t *testing.T) {
	paths := []string{"/etc/hosts", "/etc/passwd"}
	cfg, err := config.LoadOrDefault("/nonexistent/config.yaml", paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Files) != 2 {
		t.Errorf("expected 2 file entries, got %d", len(cfg.Files))
	}
	if cfg.Files[0].Path != "/etc/hosts" {
		t.Errorf("first path: got %q, want /etc/hosts", cfg.Files[0].Path)
	}
	if cfg.Interval != 30*time.Second {
		t.Errorf("default interval: got %v, want 30s", cfg.Interval)
	}
}

func TestLoadOrDefault_ErrorWhenNoFileAndNoPaths(t *testing.T) {
	_, err := config.LoadOrDefault("/nonexistent/config.yaml", nil)
	if err == nil {
		t.Fatal("expected error when no config file and no watch paths")
	}
}

func TestMustLoad_PanicsOnMissingFile(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing config file")
		}
	}()
	config.MustLoad("/nonexistent/driftwatch.yaml")
}
