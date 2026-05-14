package watchlist

import (
	"context"
	"fmt"
	"sync"
)

// Subscription holds a channel that receives path strings when the watchlist
// changes (entries added or removed). Close the returned cancel func to
// unsubscribe.
type Subscription struct {
	Ch     <-chan string
	cancel func()
}

// Cancel unregisters the subscription and closes the channel.
func (s *Subscription) Cancel() { s.cancel() }

// Bridge sits in front of a WatchList and fans out change notifications to
// registered subscribers whenever an entry is added or removed.
type Bridge struct {
	mu   sync.Mutex
	wl   *WatchList
	subs map[uint64]chan string
	next uint64
}

// NewBridge wraps wl and returns a Bridge ready for use.
func NewBridge(wl *WatchList) *Bridge {
	return &Bridge{
		wl:   wl,
		subs: make(map[uint64]chan string),
	}
}

// Subscribe returns a Subscription whose channel receives the affected path
// string on every Add or Remove call. buf controls the channel buffer size.
func (b *Bridge) Subscribe(buf int) *Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan string, buf)
	id := b.next
	b.next++
	b.subs[id] = ch

	return &Subscription{
		Ch: ch,
		cancel: func() {
			b.mu.Lock()
			delete(b.subs, id)
			close(ch)
			b.mu.Unlock()
		},
	}
}

// Add delegates to the underlying WatchList and notifies subscribers on
// success.
func (b *Bridge) Add(path string) error {
	if err := b.wl.Add(path); err != nil {
		return fmt.Errorf("bridge add: %w", err)
	}
	b.notify(path)
	return nil
}

// Remove delegates to the underlying WatchList and notifies subscribers on
// success.
func (b *Bridge) Remove(path string) error {
	if err := b.wl.Remove(path); err != nil {
		return fmt.Errorf("bridge remove: %w", err)
	}
	b.notify(path)
	return nil
}

// WatchList returns the underlying WatchList for read-only inspection.
func (b *Bridge) WatchList() *WatchList { return b.wl }

// Drain consumes all pending notifications from sub until ctx is done.
func Drain(ctx context.Context, sub *Subscription) []string {
	var out []string
	for {
		select {
		case p, ok := <-sub.Ch:
			if !ok {
				return out
			}
			out = append(out, p)
		case <-ctx.Done():
			return out
		}
	}
}

func (b *Bridge) notify(path string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subs {
		select {
		case ch <- path:
		default:
		}
	}
}
