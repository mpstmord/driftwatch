// Package grouping provides path-based grouping of drift results,
// allowing consumers to aggregate and inspect drift events by logical group.
package grouping

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/youorg/driftwatch/internal/watcher"
)

// Group holds the label and all drift results associated with it.
type Group struct {
	Label   string
	Results []watcher.DriftResult
}

// Grouper maps file paths to named groups via prefix rules and accumulates
// drift results per group.
type Grouper struct {
	mu      sync.RWMutex
	rules   []rule
	groups  map[string]*Group
}

type rule struct {
	prefix string
	label  string
}

// New returns a Grouper with no rules. Add rules via AddRule before recording results.
func New() *Grouper {
	return &Grouper{
		groups: make(map[string]*Group),
	}
}

// AddRule registers a label for all paths that share the given prefix.
// Rules are evaluated in insertion order; the first match wins.
func (g *Grouper) AddRule(prefix, label string) error {
	if prefix == "" {
		return fmt.Errorf("grouping: prefix must not be empty")
	}
	if label == "" {
		return fmt.Errorf("grouping: label must not be empty")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.rules = append(g.rules, rule{prefix: filepath.Clean(prefix), label: label})
	return nil
}

// Record classifies the result's path and appends it to the matching group.
// If no rule matches, the result is placed in a group labelled "default".
func (g *Grouper) Record(r watcher.DriftResult) {
	label := g.labelFor(r.Path)
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.groups[label]; !ok {
		g.groups[label] = &Group{Label: label}
	}
	g.groups[label].Results = append(g.groups[label].Results, r)
}

// All returns a snapshot of every group that has at least one result.
func (g *Grouper) All() []Group {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]Group, 0, len(g.groups))
	for _, grp := range g.groups {
		copy := *grp
		copy.Results = append([]watcher.DriftResult(nil), grp.Results...)
		out = append(out, copy)
	}
	return out
}

// Get returns the group for the given label and whether it exists.
func (g *Grouper) Get(label string) (Group, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	grp, ok := g.groups[label]
	if !ok {
		return Group{}, false
	}
	copy := *grp
	copy.Results = append([]watcher.DriftResult(nil), grp.Results...)
	return copy, true
}

func (g *Grouper) labelFor(path string) string {
	clean := filepath.Clean(path)
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, r := range g.rules {
		if strings.HasPrefix(clean, r.prefix) {
			return r.label
		}
	}
	return "default"
}
