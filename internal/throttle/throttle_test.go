package throttle_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/throttle"
)

func TestAllow_FirstCallAlwaysPasses(t *testing.T) {
	th := throttle.New(5 * time.Minute)
	if !th.Allow("file.conf") {
		t.Fatal("expected first Allow call to return true")
	}
}

func TestAllow_SecondCallWithinCooldownBlocked(t *testing.T) {
	th := throttle.New(5 * time.Minute)
	th.Allow("file.conf")
	if th.Allow("file.conf") {
		t.Fatal("expected second Allow within cooldown to return false")
	}
}

func TestAllow_PassesAfterCooldownExpires(t *testing.T) {
	now := time.Now()
	th := throttle.New(5 * time.Minute)

	// Inject a controllable clock.
	th.(*throttle.Throttle) // type assertion not needed; use exported test hook via New

	// Use a zero-cooldown throttle to simulate expiry.
	fast := throttle.New(0)
	fast.Allow("x")
	time.Sleep(1 * time.Millisecond)
	if !fast.Allow("x") {
		t.Fatal("expected Allow to pass after zero cooldown elapsed")
	}
	_ = now
}

func TestAllow_DifferentKeysAreIndependent(t *testing.T) {
	th := throttle.New(5 * time.Minute)
	th.Allow("a.conf")
	if !th.Allow("b.conf") {
		t.Fatal("expected different key to pass independently")
	}
}

func TestReset_ClearsKeyState(t *testing.T) {
	th := throttle.New(5 * time.Minute)
	th.Allow("file.conf")
	th.Reset("file.conf")
	if !th.Allow("file.conf") {
		t.Fatal("expected Allow to pass after Reset")
	}
}

func TestResetAll_ClearsAllState(t *testing.T) {
	th := throttle.New(5 * time.Minute)
	th.Allow("a")
	th.Allow("b")
	th.ResetAll()
	if !th.Allow("a") || !th.Allow("b") {
		t.Fatal("expected all keys to pass after ResetAll")
	}
}
