package watchlist_test

import (
	"sort"
	"testing"

	"github.com/driftwatch/internal/watchlist"
)

func TestNew_StartsEmpty(t *testing.T) {
	wl := watchlist.New()
	if wl.Len() != 0 {
		t.Fatalf("expected empty watchlist, got len %d", wl.Len())
	}
}

func TestFromPaths_PopulatesEntries(t *testing.T) {
	paths := []string{"/etc/hosts", "/etc/resolv.conf"}
	wl := watchlist.FromPaths(paths)
	if wl.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", wl.Len())
	}
	for _, p := range paths {
		if !wl.Contains(p) {
			t.Errorf("expected %q to be present", p)
		}
	}
}

func TestAdd_InsertsPath(t *testing.T) {
	wl := watchlist.New()
	if err := wl.Add("/etc/hosts"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !wl.Contains("/etc/hosts") {
		t.Error("expected path to be present after Add")
	}
}

func TestAdd_DuplicateReturnsError(t *testing.T) {
	wl := watchlist.New()
	_ = wl.Add("/etc/hosts")
	err := wl.Add("/etc/hosts")
	if err != watchlist.ErrDuplicatePath {
		t.Fatalf("expected ErrDuplicatePath, got %v", err)
	}
}

func TestRemove_DeletesPath(t *testing.T) {
	wl := watchlist.FromPaths([]string{"/etc/hosts"})
	if err := wl.Remove("/etc/hosts"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wl.Contains("/etc/hosts") {
		t.Error("expected path to be absent after Remove")
	}
}

func TestRemove_MissingPathReturnsError(t *testing.T) {
	wl := watchlist.New()
	err := wl.Remove("/nonexistent")
	if err != watchlist.ErrPathNotFound {
		t.Fatalf("expected ErrPathNotFound, got %v", err)
	}
}

func TestPaths_ReturnsCopy(t *testing.T) {
	input := []string{"/a", "/b", "/c"}
	wl := watchlist.FromPaths(input)
	got := wl.Paths()
	sort.Strings(got)
	expected := []string{"/a", "/b", "/c"}
	for i, p := range expected {
		if got[i] != p {
			t.Errorf("paths[%d]: got %q, want %q", i, got[i], p)
		}
	}
	// Mutating the returned slice must not affect the watchlist.
	got[0] = "/mutated"
	if wl.Contains("/mutated") {
		t.Error("mutating returned slice should not affect watchlist")
	}
}

func TestLen_ReflectsCurrentCount(t *testing.T) {
	wl := watchlist.New()
	_ = wl.Add("/a")
	_ = wl.Add("/b")
	if wl.Len() != 2 {
		t.Fatalf("expected len 2, got %d", wl.Len())
	}
	_ = wl.Remove("/a")
	if wl.Len() != 1 {
		t.Fatalf("expected len 1 after remove, got %d", wl.Len())
	}
}
