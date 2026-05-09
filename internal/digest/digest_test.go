package digest_test

import (
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/digest"
)

var (
	now  = time.Now()
	later = now.Add(time.Second)
)

func TestNew_DefaultsMaxLen(t *testing.T) {
	h := digest.New(0)
	for i := 0; i < 15; i++ {
		h.Record("/etc/hosts", "hash", now)
	}
	entries := h.All("/etc/hosts")
	if len(entries) != 10 {
		t.Fatalf("expected 10 entries (default cap), got %d", len(entries))
	}
}

func TestRecord_And_Latest(t *testing.T) {
	h := digest.New(5)
	h.Record("/etc/hosts", "aaa", now)
	h.Record("/etc/hosts", "bbb", later)

	e, ok := h.Latest("/etc/hosts")
	if !ok {
		t.Fatal("expected entry, got none")
	}
	if e.Hash != "bbb" {
		t.Fatalf("expected hash bbb, got %s", e.Hash)
	}
}

func TestLatest_MissingPath(t *testing.T) {
	h := digest.New(5)
	_, ok := h.Latest("/nonexistent")
	if ok {
		t.Fatal("expected false for missing path")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	h := digest.New(5)
	h.Record("/etc/ssh/sshd_config", "x1", now)
	h.Record("/etc/ssh/sshd_config", "x2", later)

	all := h.All("/etc/ssh/sshd_config")
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating the returned slice must not affect internal state.
	all[0].Hash = "mutated"
	again := h.All("/etc/ssh/sshd_config")
	if again[0].Hash == "mutated" {
		t.Fatal("All returned a reference to internal slice")
	}
}

func TestRecord_EvictsOldestWhenFull(t *testing.T) {
	h := digest.New(3)
	h.Record("/f", "h1", now)
	h.Record("/f", "h2", now)
	h.Record("/f", "h3", now)
	h.Record("/f", "h4", now)

	all := h.All("/f")
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Hash != "h2" {
		t.Fatalf("expected oldest to be h2, got %s", all[0].Hash)
	}
}

func TestChanged_DetectsDiff(t *testing.T) {
	h := digest.New(5)
	h.Record("/etc/hosts", "newHash", now)

	changed, err := h.Changed("/etc/hosts", "oldHash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}
}

func TestChanged_NoDiff(t *testing.T) {
	h := digest.New(5)
	h.Record("/etc/hosts", "sameHash", now)

	changed, err := h.Changed("/etc/hosts", "sameHash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Fatal("expected changed=false")
	}
}

func TestChanged_ErrorOnMissingPath(t *testing.T) {
	h := digest.New(5)
	_, err := h.Changed("/missing", "hash")
	if err == nil {
		t.Fatal("expected error for missing path, got nil")
	}
}
