package reporter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/driftwatch/internal/reporter"
	"github.com/driftwatch/internal/watcher"
)

func makeResults(specs []struct {
	path    string
	drifted bool
}) []watcher.DriftResult {
	var out []watcher.DriftResult
	for _, s := range specs {
		out = append(out, watcher.DriftResult{Path: s.path, Drifted: s.drifted})
	}
	return out
}

func TestNew_DefaultsToTextFormat(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf, "")
	if r == nil {
		t.Fatal("expected non-nil reporter")
	}
}

func TestReport_TextFormat_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf, reporter.FormatText)
	results := makeResults([]struct {
		path    string
		drifted bool
	}{{path: "/etc/app.conf", drifted: false}})

	if err := r.Report(results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "0/1 files drifted") {
		t.Errorf("expected summary line, got: %s", out)
	}
	if !strings.Contains(out, "[ok]") {
		t.Errorf("expected ok status, got: %s", out)
	}
}

func TestReport_TextFormat_WithDrift(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf, reporter.FormatText)
	results := makeResults([]struct {
		path    string
		drifted bool
	}{
		{path: "/etc/app.conf", drifted: true},
		{path: "/etc/other.conf", drifted: false},
	})

	if err := r.Report(results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "1/2 files drifted") {
		t.Errorf("expected drift summary, got: %s", out)
	}
	if !strings.Contains(out, "[DRIFTED]") {
		t.Errorf("expected DRIFTED status, got: %s", out)
	}
}

func TestReport_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf, reporter.FormatJSON)
	results := makeResults([]struct {
		path    string
		drifted bool
	}{{path: "/etc/app.conf", drifted: true}})

	if err := r.Report(results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"drifted":true`) {
		t.Errorf("expected drifted:true in JSON, got: %s", out)
	}
	if !strings.Contains(out, `"path":"/etc/app.conf"`) {
		t.Errorf("expected path in JSON, got: %s", out)
	}
}

func TestNew_NilSinkDefaultsToStdout(t *testing.T) {
	// Should not panic when sink is nil.
	r := reporter.New(nil, reporter.FormatText)
	if r == nil {
		t.Fatal("expected non-nil reporter")
	}
}
