package alert_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourorg/driftwatch/internal/alert"
	"github.com/yourorg/driftwatch/internal/watcher"
)

func TestHandle_NoDriftProducesNoOutput(t *testing.T) {
	var buf bytes.Buffer
	a := alert.New(&buf)
	h := alert.NewDriftHandler(a)

	result := watcher.DriftResult{
		Path:    "/etc/config.yaml",
		Drifted: false,
	}

	if err := h.Handle(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for non-drifted result, got: %s", buf.String())
	}
}

func TestHandle_DriftedProducesAlert(t *testing.T) {
	var buf bytes.Buffer
	a := alert.New(&buf)
	h := alert.NewDriftHandler(a)

	result := watcher.DriftResult{
		Path:     "/etc/config.yaml",
		Drifted:  true,
		Previous: "abc123",
		Current:  "def456",
	}

	if err := h.Handle(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "/etc/config.yaml") {
		t.Errorf("expected file path in alert output, got: %s", out)
	}
	if !strings.Contains(out, "abc123") {
		t.Errorf("expected previous checksum in alert output, got: %s", out)
	}
	if !strings.Contains(out, "def456") {
		t.Errorf("expected current checksum in alert output, got: %s", out)
	}
}
