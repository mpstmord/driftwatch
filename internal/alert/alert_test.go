package alert_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/alert"
)

func TestNew_DefaultsToStdout(t *testing.T) {
	a := alert.New()
	if a == nil {
		t.Fatal("expected non-nil Alerter")
	}
}

func TestSend_WritesToSink(t *testing.T) {
	var buf bytes.Buffer
	a := alert.New(&buf)

	e := alert.Event{
		Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		Level:     alert.LevelWarn,
		FilePath:  "/etc/app/config.yaml",
		Message:   "checksum mismatch",
	}

	if err := a.Send(e); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "WARN") {
		t.Errorf("expected WARN in output, got: %s", out)
	}
	if !strings.Contains(out, "/etc/app/config.yaml") {
		t.Errorf("expected file path in output, got: %s", out)
	}
	if !strings.Contains(out, "checksum mismatch") {
		t.Errorf("expected message in output, got: %s", out)
	}
}

func TestSend_MultipleWriters(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	a := alert.New(&buf1, &buf2)

	e := alert.Event{
		Timestamp: time.Now(),
		Level:     alert.LevelError,
		FilePath:  "/etc/hosts",
		Message:   "file removed",
	}

	if err := a.Send(e); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if buf1.String() != buf2.String() {
		t.Errorf("sinks received different output")
	}
}

func TestDrift_SendsWarnLevel(t *testing.T) {
	var buf bytes.Buffer
	a := alert.New(&buf)

	if err := a.Drift("/etc/nginx/nginx.conf", "unexpected change"); err != nil {
		t.Fatalf("Drift returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "WARN") {
		t.Errorf("expected WARN level in drift alert, got: %s", out)
	}
}
