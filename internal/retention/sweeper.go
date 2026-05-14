package retention

import (
	"context"
	"time"
)

// Sweeper periodically calls Evict on a Store at a fixed interval.
type Sweeper struct {
	store    *Store
	interval time.Duration
}

// NewSweeper creates a Sweeper that will evict expired entries from store
// every interval duration.
func NewSweeper(store *Store, interval time.Duration) *Sweeper {
	return &Sweeper{store: store, interval: interval}
}

// Run starts the sweep loop. It blocks until ctx is cancelled.
func (sw *Sweeper) Run(ctx context.Context) {
	ticker := time.NewTicker(sw.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sw.store.Evict()
		}
	}
}
