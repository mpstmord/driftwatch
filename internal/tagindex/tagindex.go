// Package tagindex maintains a mapping from user-defined string tags to the
// set of watched paths that carry those tags.  It allows callers to quickly
// look up which files belong to a logical group (e.g. "nginx", "critical",
// "pci") without scanning every entry in the watchlist.
package tagindex

import (
	"fmt"
	"sync"
)

// Index maps tags to the paths that carry them.
type Index struct {
	mu      sync.RWMutex
	byTag   map[string]map[string]struct{} // tag -> set of paths
	byPath  map[string]map[string]struct{} // path -> set of tags
}

// New returns an empty Index.
func New() *Index {
	return &Index{
		byTag:  make(map[string]map[string]struct{}),
		byPath: make(map[string]map[string]struct{}),
	}
}

// Tag associates one or more tags with a path.  Duplicate tags are silently
// ignored.  Returns an error if path or any tag is empty.
func (idx *Index) Tag(path string, tags ...string) error {
	if path == "" {
		return fmt.Errorf("tagindex: path must not be empty")
	}
	for _, t := range tags {
		if t == "" {
			return fmt.Errorf("tagindex: tag must not be empty")
		}
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	if _, ok := idx.byPath[path]; !ok {
		idx.byPath[path] = make(map[string]struct{})
	}
	for _, t := range tags {
		idx.byPath[path][t] = struct{}{}
		if _, ok := idx.byTag[t]; !ok {
			idx.byTag[t] = make(map[string]struct{})
		}
		idx.byTag[t][path] = struct{}{}
	}
	return nil
}

// Untag removes all tag associations for the given path.
func (idx *Index) Untag(path string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for t := range idx.byPath[path] {
		delete(idx.byTag[t], path)
		if len(idx.byTag[t]) == 0 {
			delete(idx.byTag, t)
		}
	}
	delete(idx.byPath, path)
}

// PathsForTag returns a snapshot of all paths associated with the given tag.
// Returns nil when the tag is unknown.
func (idx *Index) PathsForTag(tag string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	set, ok := idx.byTag[tag]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	return out
}

// TagsForPath returns a snapshot of all tags associated with the given path.
// Returns nil when the path is unknown.
func (idx *Index) TagsForPath(path string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	set, ok := idx.byPath[path]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	return out
}
