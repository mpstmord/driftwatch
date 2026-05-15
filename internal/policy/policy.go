// Package policy evaluates a set of named rules against drift results,
// assigning a disposition (allow, warn, block) to each drifted path.
package policy

import (
	"regexp"
	"sync"

	"github.com/example/driftwatch/internal/watcher"
)

// Disposition describes the outcome of a policy evaluation.
type Disposition int

const (
	Allow Disposition = iota // drift is acknowledged; no escalation
	Warn                     // drift should produce a warning
	Block                    // drift should halt or page on-call
)

func (d Disposition) String() string {
	switch d {
	case Allow:
		return "allow"
	case Warn:
		return "warn"
	case Block:
		return "block"
	default:
		return "unknown"
	}
}

// Rule pairs a compiled path pattern with its disposition.
type Rule struct {
	Name    string
	Pattern *regexp.Regexp
	Dispose Disposition
}

// Policy holds an ordered list of rules and evaluates drift results.
type Policy struct {
	mu    sync.RWMutex
	rules []Rule
}

// New returns an empty Policy.
func New() *Policy {
	return &Policy{}
}

// AddRule appends a rule to the policy. Rules are evaluated in insertion order;
// the first match wins.
func (p *Policy) AddRule(r Rule) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rules = append(p.rules, r)
}

// Evaluate returns the Disposition for the given drift result. If no rule
// matches, Warn is returned as the safe default.
func (p *Policy) Evaluate(r watcher.DriftResult) Disposition {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, rule := range p.rules {
		if rule.Pattern.MatchString(r.Path) {
			return rule.Dispose
		}
	}
	return Warn
}

// Rules returns a snapshot of the current rule list.
func (p *Policy) Rules() []Rule {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]Rule, len(p.rules))
	copy(out, p.rules)
	return out
}
