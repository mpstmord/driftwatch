package eventlog_test

import (
	"testing"

	"github.com/driftwatch/internal/eventlog"
	"github.com/driftwatch/internal/watcher"
)

func driftedResult(path, old, new string) watcher.DriftResult {
	return watcher.DriftResult{Path: path, Drifted: true, OldHash: old, NewHash: new}
}

func cleanResult(path string) watcher.DriftResult {
	return watcher.DriftResult{Path: path, Drifted: false}
}

func TestNew_DefaultsMaxLen(t *testing.T) {
	el := eventlog.New(0)
	if el == nil {
		t.Fatal("expected non-nil EventLog")
	}
}

func TestRecord_IgnoresNonDriftedResults(t *testing.T) {
	el := eventlog.New(10)
	el.Record(cleanResult("/etc/hosts"))
	if el.Len() != 0 {
		t.Fatalf("expected 0 entries, got %d", el.Len())
	}
}

func TestRecord_AppendsDriftedResult(t *testing.T) {
	el := eventlog.New(10)
	el.Record(driftedResult("/etc/hosts", "aaa", "bbb"))
	if el.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", el.Len())
	}
	entries := el.All()
	if entries[0].Path != "/etc/hosts" {
		t.Errorf("unexpected path: %s", entries[0].Path)
	}
	if entries[0].OldHash != "aaa" || entries[0].NewHash != "bbb" {
		t.Errorf("unexpected hashes: %s %s", entries[0].OldHash, entries[0].NewHash)
	}
	if entries[0].DetectedAt.IsZero() {
		t.Error("DetectedAt should not be zero")
	}
}

func TestRecord_EvictsOldestWhenFull(t *testing.T) {
	el := eventlog.New(3)
	el.Record(driftedResult("/a", "1", "2"))
	el.Record(driftedResult("/b", "1", "2"))
	el.Record(driftedResult("/c", "1", "2"))
	el.Record(driftedResult("/d", "1", "2"))

	if el.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", el.Len())
	}
	entries := el.All()
	if entries[0].Path != "/b" {
		t.Errorf("expected oldest evicted; got %s", entries[0].Path)
	}
	if entries[2].Path != "/d" {
		t.Errorf("expected newest last; got %s", entries[2].Path)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	el := eventlog.New(10)
	el.Record(driftedResult("/etc/hosts", "x", "y"))
	a := el.All()
	a[0].Path = "mutated"
	b := el.All()
	if b[0].Path == "mutated" {
		t.Error("All() should return an independent copy")
	}
}

func TestClear_RemovesAllEntries(t *testing.T) {
	el := eventlog.New(10)
	el.Record(driftedResult("/etc/hosts", "x", "y"))
	el.Record(driftedResult("/etc/passwd", "x", "y"))
	el.Clear()
	if el.Len() != 0 {
		t.Fatalf("expected 0 after Clear, got %d", el.Len())
	}
}
