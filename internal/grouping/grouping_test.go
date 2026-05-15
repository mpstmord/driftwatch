package grouping_test

import (
	"testing"
	"time"

	"github.com/youorg/driftwatch/internal/grouping"
	"github.com/youorg/driftwatch/internal/watcher"
)

func driftAt(path string) watcher.DriftResult {
	return watcher.DriftResult{
		Path:    path,
		Drifted: true,
		At:      time.Now(),
	}
}

func TestRecord_FallsBackToDefault(t *testing.T) {
	g := grouping.New()
	g.Record(driftAt("/etc/app/config.yaml"))

	grp, ok := g.Get("default")
	if !ok {
		t.Fatal("expected default group to exist")
	}
	if len(grp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(grp.Results))
	}
}

func TestRecord_MatchesRuleByPrefix(t *testing.T) {
	g := grouping.New()
	if err := g.AddRule("/etc/app", "app"); err != nil {
		t.Fatalf("AddRule: %v", err)
	}
	g.Record(driftAt("/etc/app/config.yaml"))

	grp, ok := g.Get("app")
	if !ok {
		t.Fatal("expected 'app' group to exist")
	}
	if len(grp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(grp.Results))
	}
}

func TestRecord_FirstRuleWins(t *testing.T) {
	g := grouping.New()
	_ = g.AddRule("/etc", "system")
	_ = g.AddRule("/etc/app", "app")
	g.Record(driftAt("/etc/app/config.yaml"))

	if _, ok := g.Get("app"); ok {
		t.Fatal("second rule should not have matched")
	}
	grp, ok := g.Get("system")
	if !ok {
		t.Fatal("expected 'system' group")
	}
	if len(grp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(grp.Results))
	}
}

func TestAll_ReturnsAllGroups(t *testing.T) {
	g := grouping.New()
	_ = g.AddRule("/etc/app", "app")
	_ = g.AddRule("/var/log", "logs")
	g.Record(driftAt("/etc/app/config.yaml"))
	g.Record(driftAt("/var/log/syslog"))
	g.Record(driftAt("/tmp/other"))

	all := g.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(all))
	}
}

func TestAddRule_EmptyPrefixReturnsError(t *testing.T) {
	g := grouping.New()
	if err := g.AddRule("", "label"); err == nil {
		t.Fatal("expected error for empty prefix")
	}
}

func TestAddRule_EmptyLabelReturnsError(t *testing.T) {
	g := grouping.New()
	if err := g.AddRule("/etc", ""); err == nil {
		t.Fatal("expected error for empty label")
	}
}

func TestGet_MissingGroupReturnsFalse(t *testing.T) {
	g := grouping.New()
	if _, ok := g.Get("nonexistent"); ok {
		t.Fatal("expected ok=false for missing group")
	}
}
