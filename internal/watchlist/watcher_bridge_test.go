package watchlist

import (
	"context"
	"testing"
	"time"
)

func TestBridge_AddNotifiesSubscribers(t *testing.T) {
	wl := New()
	b := NewBridge(wl)

	sub := b.Subscribe(4)
	defer sub.Cancel()

	if err := b.Add("/etc/hosts"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	select {
	case got := <-sub.Ch:
		if got != "/etc/hosts" {
			t.Errorf("expected /etc/hosts, got %q", got)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for notification")
	}
}

func TestBridge_RemoveNotifiesSubscribers(t *testing.T) {
	wl := FromPaths([]string{"/etc/hosts"})
	b := NewBridge(wl)

	sub := b.Subscribe(4)
	defer sub.Cancel()

	if err := b.Remove("/etc/hosts"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	select {
	case got := <-sub.Ch:
		if got != "/etc/hosts" {
			t.Errorf("expected /etc/hosts, got %q", got)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for notification")
	}
}

func TestBridge_MultipleSubscribersAllReceive(t *testing.T) {
	wl := New()
	b := NewBridge(wl)

	s1 := b.Subscribe(4)
	s2 := b.Subscribe(4)
	defer s1.Cancel()
	defer s2.Cancel()

	_ = b.Add("/etc/passwd")

	for _, sub := range []*Subscription{s1, s2} {
		select {
		case p := <-sub.Ch:
			if p != "/etc/passwd" {
				t.Errorf("unexpected path %q", p)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timed out")
		}
	}
}

func TestBridge_CancelRemovesSubscription(t *testing.T) {
	wl := New()
	b := NewBridge(wl)

	sub := b.Subscribe(4)
	sub.Cancel()

	// After cancel the channel is closed; Add should not panic.
	_ = b.Add("/etc/hosts")

	b.mu.Lock()
	count := len(b.subs)
	b.mu.Unlock()

	if count != 0 {
		t.Errorf("expected 0 subscribers after cancel, got %d", count)
	}
}

func TestDrain_CollectsNotifications(t *testing.T) {
	wl := New()
	b := NewBridge(wl)

	sub := b.Subscribe(8)
	defer sub.Cancel()

	paths := []string{"/a", "/b", "/c"}
	for _, p := range paths {
		_ = b.Add(p)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	got := Drain(ctx, sub)
	if len(got) != len(paths) {
		t.Errorf("expected %d notifications, got %d", len(paths), len(got))
	}
}
