package quota

import (
	"context"

	"github.com/example/driftwatch/internal/watcher"
)

// Middleware wraps a drift-result handler and gates each result through the
// Quota. Results that exceed the quota are silently dropped.
type Middleware struct {
	quota *Quota
	next  func(context.Context, watcher.DriftResult)
}

// NewMiddleware returns a Middleware that enforces q before forwarding results
// to next.
func NewMiddleware(q *Quota, next func(context.Context, watcher.DriftResult)) *Middleware {
	return &Middleware{quota: q, next: next}
}

// Handle checks the quota for result.Path. If the quota permits the alert the
// result is forwarded to the wrapped handler; otherwise it is dropped.
func (m *Middleware) Handle(ctx context.Context, result watcher.DriftResult) {
	if !m.quota.Allow(result.Path) {
		return
	}
	m.next(ctx, result)
}
