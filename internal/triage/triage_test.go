package triage_test

import (
	"testing"

	"github.com/driftwatch/internal/triage"
	"github.com/driftwatch/internal/watcher"
)

func driftAt(path string) watcher.DriftResult {
	return watcher.DriftResult{Path: path, Drifted: true}
}

func TestClassify_DefaultsToInfo(t *testing.T) {
	tr, err := triage.New(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := tr.Classify(driftAt("/etc/hosts"))
	if r.Level != triage.LevelInfo {
		t.Errorf("expected INFO, got %s", r.Level)
	}
}

func TestClassify_WarnPatternElevates(t *testing.T) {
	tr, err := triage.New([]string{`/etc/.*`}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := tr.Classify(driftAt("/etc/ssh/sshd_config"))
	if r.Level != triage.LevelWarn {
		t.Errorf("expected WARN, got %s", r.Level)
	}
}

func TestClassify_CritPatternOverridesWarn(t *testing.T) {
	tr, err := triage.New([]string{`/etc/.*`}, []string{`/etc/passwd`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := tr.Classify(driftAt("/etc/passwd"))
	if r.Level != triage.LevelCrit {
		t.Errorf("expected CRIT, got %s", r.Level)
	}
}

func TestClassify_NonMatchingPathRemainsInfo(t *testing.T) {
	tr, err := triage.New([]string{`/etc/.*`}, []string{`/etc/passwd`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := tr.Classify(driftAt("/var/log/app.log"))
	if r.Level != triage.LevelInfo {
		t.Errorf("expected INFO, got %s", r.Level)
	}
}

func TestNew_InvalidPatternReturnsError(t *testing.T) {
	_, err := triage.New([]string{`[invalid`}, nil)
	if err == nil {
		t.Fatal("expected error for invalid warn pattern, got nil")
	}
}

func TestNew_InvalidCritPatternReturnsError(t *testing.T) {
	_, err := triage.New(nil, []string{`[invalid`})
	if err == nil {
		t.Fatal("expected error for invalid crit pattern, got nil")
	}
}

func TestLevel_String(t *testing.T) {
	cases := []struct {
		level triage.Level
		want  string
	}{
		{triage.LevelInfo, "INFO"},
		{triage.LevelWarn, "WARN"},
		{triage.LevelCrit, "CRIT"},
	}
	for _, tc := range cases {
		if got := tc.level.String(); got != tc.want {
			t.Errorf("Level(%d).String() = %q, want %q", tc.level, got, tc.want)
		}
	}
}
