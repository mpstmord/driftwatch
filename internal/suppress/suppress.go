// Package suppress provides a mechanism to temporarily suppress drift alerts
// for specific file paths. This is useful during planned maintenance windows
// or known configuration changes.
package suppress

import (
	"sync"
	"time"
)

// Suppression holds the expiry time for a suppressed path.
type Suppression struct {
	Path      string
	ExpiresAt time.Time
}

// Manager tracks active suppressions keyed by file path.
type Manager struct {
	mu          sync.RWMutex
	suppressions map[string]Suppression
	now         func() time.Time
}

// New creates a new Manager.
func New() *Manager {
	return &Manager{
		suppressions: make(map[string]Suppression),
		now:         time.Now,
	}
}

// Suppress adds a suppression for the given path lasting for the given duration.
func (m *Manager) Suppress(path string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.suppressions[path] = Suppression{
		Path:      path,
		ExpiresAt: m.now().Add(duration),
	}
}

// IsSuppressed reports whether the given path currently has an active suppression.
func (m *Manager) IsSuppressed(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.suppressions[path]
	if !ok {
		return false
	}
	return m.now().Before(s.ExpiresAt)
}

// Lift removes a suppression for the given path immediately.
func (m *Manager) Lift(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.suppressions, path)
}

// Purge removes all expired suppressions.
func (m *Manager) Purge() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	for k, s := range m.suppressions {
		if !now.Before(s.ExpiresAt) {
			delete(m.suppressions, k)
		}
	}
}

// Active returns a snapshot of all currently active suppressions.
func (m *Manager) Active() []Suppression {
	m.mu.RLock()
	defer m.mu.RUnlock()
	now := m.now()
	out := make([]Suppression, 0, len(m.suppressions))
	for _, s := range m.suppressions {
		if now.Before(s.ExpiresAt) {
			out = append(out, s)
		}
	}
	return out
}
