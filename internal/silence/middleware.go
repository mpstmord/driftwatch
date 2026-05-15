package silence

import (
	"context"

	"github.com/yourusername/driftwatch/internal/watcher"
)

// Middleware wraps a drift-result channel and filters out results whose
// paths are currently silenced. Non-silenced results are forwarded to the
// returned channel.
//
// The returned channel is closed when ctx is cancelled or in is closed.
func Middleware(ctx context.Context, in <-chan watcher.DriftResult, sr *Silence) <-chan watcher.DriftResult {
	out := make(chan watcher.DriftResult)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case r, ok := <-in:
				if !ok {
					return
				}
				if sr.IsSilenced(r.Path) {
					continue
				}
				select {
				case out <- r:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
