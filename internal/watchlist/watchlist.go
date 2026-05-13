// Package watchlist manages the set of file paths that driftwatch monitors,
// supporting dynamic add/remove operations with thread-safe access.
package watchlist

import (
	"errors"
	"sync"
)

// ErrPathNotFound is returned when a requested path is not in the watchlist.
var ErrPathNotFound = errors.New("watchlist: path not found")

// ErrDuplicatePath is returned when adding a path that already exists.
var ErrDuplicatePath = errors.New("watchlist: path already exists")

// WatchList holds the set of file paths currently being monitored.
type WatchList struct {
	mu    sync.RWMutex
	paths map[string]struct{}
}

// New creates an empty WatchList.
func New() *WatchList {
	return &WatchList{
		paths: make(map[string]struct{}),
	}
}

// FromPaths creates a WatchList pre-populated with the given paths.
func FromPaths(paths []string) *WatchList {
	wl := New()
	for _, p := range paths {
		wl.paths[p] = struct{}{}
	}
	return wl
}

// Add inserts a path into the watchlist. Returns ErrDuplicatePath if it already exists.
func (wl *WatchList) Add(path string) error {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	if _, ok := wl.paths[path]; ok {
		return ErrDuplicatePath
	}
	wl.paths[path] = struct{}{}
	return nil
}

// Remove deletes a path from the watchlist. Returns ErrPathNotFound if missing.
func (wl *WatchList) Remove(path string) error {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	if _, ok := wl.paths[path]; !ok {
		return ErrPathNotFound
	}
	delete(wl.paths, path)
	return nil
}

// Contains reports whether the given path is in the watchlist.
func (wl *WatchList) Contains(path string) bool {
	wl.mu.RLock()
	defer wl.mu.RUnlock()
	_, ok := wl.paths[path]
	return ok
}

// Paths returns a snapshot of all currently watched paths.
func (wl *WatchList) Paths() []string {
	wl.mu.RLock()
	defer wl.mu.RUnlock()
	out := make([]string, 0, len(wl.paths))
	for p := range wl.paths {
		out = append(out, p)
	}
	return out
}

// Len returns the number of paths in the watchlist.
func (wl *WatchList) Len() int {
	wl.mu.RLock()
	defer wl.mu.RUnlock()
	return len(wl.paths)
}
