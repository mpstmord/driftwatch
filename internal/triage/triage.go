// Package triage classifies drift results by severity based on
// configurable path patterns and escalation history.
package triage

import (
	"regexp"
	"sync"

	"github.com/driftwatch/internal/watcher"
)

// Level represents a severity classification for a drift event.
type Level int

const (
	LevelInfo  Level = iota // default, low-priority drift
	LevelWarn               // elevated, matches a watched pattern
	LevelCrit               // critical, matches a critical pattern
)

// String returns the human-readable label for a Level.
func (l Level) String() string {
	switch l {
	case LevelWarn:
		return "WARN"
	case LevelCrit:
		return "CRIT"
	default:
		return "INFO"
	}
}

// Result pairs a drift result with its assigned severity level.
type Result struct {
	Drift watcher.DriftResult
	Level Level
}

// Triage classifies incoming drift results against compiled pattern sets.
type Triage struct {
	mu       sync.RWMutex
	warnPats []*regexp.Regexp
	critPats []*regexp.Regexp
}

// New creates a Triage instance with the given warn and critical glob-style
// patterns. Patterns use Go regexp syntax.
func New(warnPatterns, critPatterns []string) (*Triage, error) {
	warn, err := compileAll(warnPatterns)
	if err != nil {
		return nil, err
	}
	crit, err := compileAll(critPatterns)
	if err != nil {
		return nil, err
	}
	return &Triage{warnPats: warn, critPats: crit}, nil
}

// Classify assigns a severity Level to a drift result.
func (t *Triage) Classify(d watcher.DriftResult) Result {
	t.mu.RLock()
	defer t.mu.RUnlock()

	lvl := LevelInfo
	if matchAny(t.warnPats, d.Path) {
		lvl = LevelWarn
	}
	if matchAny(t.critPats, d.Path) {
		lvl = LevelCrit
	}
	return Result{Drift: d, Level: lvl}
}

func compileAll(patterns []string) ([]*regexp.Regexp, error) {
	out := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		out = append(out, re)
	}
	return out, nil
}

func matchAny(pats []*regexp.Regexp, path string) bool {
	for _, re := range pats {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}
