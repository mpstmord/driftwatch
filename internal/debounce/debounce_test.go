package debounce_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/debounce"
)

const shortWait = 50 * time.Millisecond

func TestTrigger_CallsFnAfterQuietWindow(t *testing.T) {
	d := debounce.New(shortWait)

	var called int32
	d.Trigger("a", func() { atomic.StoreInt32(&called, 1) })

	if d.Pending("a") == false {
		t.Fatal("expected timer to be pending")
	}

	time.Sleep(shortWait * 3)

	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("expected callback to have been called")
	}
	if d.Pending("a") {
		t.Fatal("expected timer to be cleared after firing")
	}
}

func TestTrigger_ResetsTimerOnRepeatCall(t *testing.T) {
	d := debounce.New(shortWait)

	var count int32
	trigger := func() { atomic.AddInt32(&count, 1) }

	// Rapidly re-trigger — only ONE call should fire.
	for i := 0; i < 5; i++ {
		d.Trigger("b", trigger)
		time.Sleep(shortWait / 5)
	}

	time.Sleep(shortWait * 3)

	if n := atomic.LoadInt32(&count); n != 1 {
		t.Fatalf("expected 1 call, got %d", n)
	}
}

func TestCancel_StopsPendingTimer(t *testing.T) {
	d := debounce.New(shortWait)

	var called int32
	d.Trigger("c", func() { atomic.StoreInt32(&called, 1) })
	d.Cancel("c")

	time.Sleep(shortWait * 3)

	if atomic.LoadInt32(&called) != 0 {
		t.Fatal("expected callback NOT to be called after cancel")
	}
	if d.Pending("c") {
		t.Fatal("expected timer to be absent after cancel")
	}
}

func TestPending_DifferentKeysAreIndependent(t *testing.T) {
	d := debounce.New(shortWait)

	d.Trigger("x", func() {})

	if !d.Pending("x") {
		t.Fatal("expected x to be pending")
	}
	if d.Pending("y") {
		t.Fatal("expected y NOT to be pending")
	}

	d.Cancel("x")
}

func TestCancel_NoopOnMissingKey(t *testing.T) {
	d := debounce.New(shortWait)
	// Should not panic.
	d.Cancel("nonexistent")
}
