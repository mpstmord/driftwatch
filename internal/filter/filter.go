// Package filter provides path-based filtering for driftwatch,
// allowing users to include or exclude specific files or glob patterns
// from drift detection.
package filter

import (
	"path/filepath"
	"strings"
)

// Filter holds compiled include and exclude glob patterns.
type Filter struct {
	includes []string
	excludes []string
}

// New creates a Filter from the given include and exclude pattern slices.
// An empty includes list means "allow all" (before excludes are applied).
func New(includes, excludes []string) *Filter {
	return &Filter{
		includes: includes,
		excludes: excludes,
	}
}

// Allow reports whether the given path should be processed.
// A path is allowed when:
//  1. It matches at least one include pattern (or no includes are defined), AND
//  2. It does not match any exclude pattern.
func (f *Filter) Allow(path string) (bool, error) {
	path = filepath.ToSlash(strings.TrimSpace(path))

	if len(f.includes) > 0 {
		matched, err := matchAny(path, f.includes)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}

	if len(f.excludes) > 0 {
		matched, err := matchAny(path, f.excludes)
		if err != nil {
			return false, err
		}
		if matched {
			return false, nil
		}
	}

	return true, nil
}

// matchAny returns true if path matches any of the provided glob patterns.
func matchAny(path string, patterns []string) (bool, error) {
	for _, pattern := range patterns {
		ok, err := filepath.Match(filepath.ToSlash(pattern), path)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}
