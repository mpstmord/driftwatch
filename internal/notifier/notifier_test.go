package notifier_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/driftwatch/driftwatch/internal/notifier"
)

func TestNew_DefaultsToStderr(t *testing.T) {
	n := notifier.New(nil)
	if n == nil {
		t.Fatal("expected non-nil LogNotifier")
	}
}

func TestNotify_WritesFormattedLine(t *testing.T) {
	var buf bytes.Buffer
	n := notifier.New(&buf)

	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	err := n.Notify(notifier.Event{
		Level:     notifier.LevelWarn,
		File:      "/etc/app/config.yaml",
		Message:   "checksum mismatch",
		Timestamp: ts,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "WARN") {
		t.Errorf("expected WARN in output, got: %s", got)
	}
	if !strings.Contains(got, "/etc/app/config.yaml") {
		t.Errorf("expected file path in output, got: %s", got)
	}
	if !strings.Contains(got, "checksum mismatch") {
		t.Errorf("expected message in output, got: %s", got)
	}
	if !strings.Contains(got, "2024-06-01T12:00:00Z") {
		t.Errorf("expected timestamp in output, got: %s", got)
	}
}

func TestNotify_FillsTimestampWhenZero(t *testing.T) {
	var buf bytes.Buffer
	n := notifier.New(&buf)

	err := n.Notify(notifier.Event{
		Level:   notifier.LevelInfo,
		File:    "/etc/hosts",
		Message: "no drift",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "INFO") {
		t.Errorf("expected INFO level, got: %s", got)
	}
	// Timestamp should be present and non-empty (not zero value)
	if strings.Contains(got, "0001-01-01") {
		t.Errorf("zero timestamp should have been replaced, got: %s", got)
	}
}

func TestNotify_MultipleEvents(t *testing.T) {
	var buf bytes.Buffer
	n := notifier.New(&buf)

	events := []notifier.Event{
		{Level: notifier.LevelWarn, File: "/etc/a", Message: "drift a"},
		{Level: notifier.LevelError, File: "/etc/b", Message: "drift b"},
	}
	for _, e := range events {
		if err := n.Notify(e); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %s", len(lines), buf.String())
	}
}
