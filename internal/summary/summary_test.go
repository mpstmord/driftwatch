package summary_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/driftwatch/internal/summary"
	"github.com/driftwatch/internal/watcher"
)

func drifted(path string) watcher.DriftResult {
	return watcher.DriftResult{Path: path, Drifted: true}
}

func clean(path string) watcher.DriftResult {
	return watcher.DriftResult{Path: path, Drifted: false}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	s := summary.New(time.Second, summary.FormatText, nil)
	if s == nil {
		t.Fatal("expected non-nil Summariser")
	}
}

func TestRun_FlushesOnWindowExpiry(t *testing.T) {
	var buf bytes.Buffer
	s := summary.New(50*time.Millisecond, summary.FormatText, &buf)

	s.Add(drifted("/etc/hosts"))
	s.Add(clean("/etc/resolv.conf"))

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()
	s.Run(ctx)

	out := buf.String()
	if !strings.Contains(out, "drifted=1") {
		t.Errorf("expected drifted=1 in output, got: %s", out)
	}
	if !strings.Contains(out, "clean=1") {
		t.Errorf("expected clean=1 in output, got: %s", out)
	}
	if !strings.Contains(out, "/etc/hosts") {
		t.Errorf("expected drifted path in output, got: %s", out)
	}
}

func TestRun_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	s := summary.New(40*time.Millisecond, summary.FormatJSON, &buf)
	s.Add(drifted("/etc/ssh/sshd_config"))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	s.Run(ctx)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 0 {
		t.Fatal("expected at least one JSON line")
	}
	var rep summary.Report
	if err := json.Unmarshal([]byte(lines[0]), &rep); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if rep.Drifted != 1 {
		t.Errorf("expected Drifted=1, got %d", rep.Drifted)
	}
}

func TestRun_NoDriftProducesCleanReport(t *testing.T) {
	var buf bytes.Buffer
	s := summary.New(40*time.Millisecond, summary.FormatText, &buf)
	s.Add(clean("/etc/hosts"))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	s.Run(ctx)

	out := buf.String()
	if !strings.Contains(out, "drifted=0") {
		t.Errorf("expected drifted=0, got: %s", out)
	}
	if strings.Contains(out, "/etc/hosts") {
		t.Errorf("clean path should not appear in drifted_paths: %s", out)
	}
}

func TestRun_CancelFlushesRemainder(t *testing.T) {
	var buf bytes.Buffer
	s := summary.New(10*time.Second, summary.FormatText, &buf)
	s.Add(drifted("/tmp/cfg"))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()
	s.Run(ctx)

	if !strings.Contains(buf.String(), "drifted=1") {
		t.Errorf("expected flush on cancel, got: %s", buf.String())
	}
}
