package rollup_test

import (
	"strings"
	"testing"

	"github.com/example/driftwatch/internal/rollup"
	"github.com/example/driftwatch/internal/watcher"
)

func TestSummarise_NoDrift(t *testing.T) {
	results := []watcher.DriftResult{
		{Path: "/etc/hosts", Drifted: false},
		{Path: "/etc/passwd", Drifted: false},
	}
	s := rollup.Summarise(results)
	if s.Total != 2 {
		t.Errorf("expected Total=2, got %d", s.Total)
	}
	if s.Drifted != 0 {
		t.Errorf("expected Drifted=0, got %d", s.Drifted)
	}
	if len(s.Paths) != 0 {
		t.Errorf("expected no drifted paths, got %v", s.Paths)
	}
}

func TestSummarise_WithDrift(t *testing.T) {
	results := []watcher.DriftResult{
		{Path: "/etc/hosts", Drifted: true},
		{Path: "/etc/passwd", Drifted: false},
		{Path: "/etc/nginx/nginx.conf", Drifted: true},
	}
	s := rollup.Summarise(results)
	if s.Total != 3 {
		t.Errorf("expected Total=3, got %d", s.Total)
	}
	if s.Drifted != 2 {
		t.Errorf("expected Drifted=2, got %d", s.Drifted)
	}
	if len(s.Paths) != 2 {
		t.Errorf("expected 2 drifted paths, got %v", s.Paths)
	}
}

func TestSummary_String_NoDrift(t *testing.T) {
	s := rollup.Summary{Total: 3, Drifted: 0}
	if !strings.Contains(s.String(), "no drift") {
		t.Errorf("expected 'no drift' in output, got: %s", s.String())
	}
}

func TestSummary_String_WithDrift(t *testing.T) {
	s := rollup.Summary{
		Total:   3,
		Drifted: 2,
		Paths:   []string{"/etc/hosts", "/etc/nginx/nginx.conf"},
	}
	out := s.String()
	if !strings.Contains(out, "2/3") {
		t.Errorf("expected '2/3' in output, got: %s", out)
	}
	if !strings.Contains(out, "/etc/hosts") {
		t.Errorf("expected path in output, got: %s", out)
	}
}
