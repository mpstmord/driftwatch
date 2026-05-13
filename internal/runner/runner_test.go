package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/driftwatch/internal/config"
	"github.com/example/driftwatch/internal/runner"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return p
}

func newTestConfig(t *testing.T, paths []string) *config.Config {
	t.Helper()
	dir := t.TempDir()
	return &config.Config{
		WatchPaths:   paths,
		BaselinePath: filepath.Join(dir, "baseline.json"),
		Format:       "text",
	}
}

// newRunnerWithBaseline is a test helper that creates a Runner and immediately
// establishes a baseline, so tests can focus on drift-detection behaviour.
func newRunnerWithBaseline(t *testing.T, paths []string) *runner.Runner {
	t.Helper()
	cfg := newTestConfig(t, paths)
	r, err := runner.New(cfg)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if err := r.UpdateBaseline(); err != nil {
		t.Fatalf("UpdateBaseline(): %v", err)
	}
	return r
}

func TestNew_CreatesRunnerWithoutExistingBaseline(t *testing.T) {
	dir := t.TempDir()
	f := writeTempFile(t, dir, "app.conf", "key=value")
	cfg := newTestConfig(t, []string{f})

	r, err := runner.New(cfg)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("New() returned nil runner")
	}
}

func TestRun_NoDriftOnFreshBaseline(t *testing.T) {
	dir := t.TempDir()
	f := writeTempFile(t, dir, "app.conf", "key=value")

	r := newRunnerWithBaseline(t, []string{f})

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
}

func TestRun_RespectsContextCancellation(t *testing.T) {
	dir := t.TempDir()
	f := writeTempFile(t, dir, "app.conf", "key=value")
	cfg := newTestConfig(t, []string{f})

	r, err := runner.New(cfg)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	if err := r.Run(ctx); err == nil {
		t.Fatal("Run() expected error on cancelled context, got nil")
	}
}

func TestUpdateBaseline_PersistsFile(t *testing.T) {
	dir := t.TempDir()
	f := writeTempFile(t, dir, "app.conf", "key=value")
	cfg := newTestConfig(t, []string{f})

	r, err := runner.New(cfg)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	if err := r.UpdateBaseline(); err != nil {
		t.Fatalf("UpdateBaseline(): %v", err)
	}

	if _, err := os.Stat(cfg.BaselinePath); os.IsNotExist(err) {
		t.Fatalf("expected baseline file at %s, not found", cfg.BaselinePath)
	}
}
