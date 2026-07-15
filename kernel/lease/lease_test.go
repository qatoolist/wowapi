package lease

import (
	"testing"
	"time"
)

func TestNewLeaseHasTokenGenerationAndExpiry(t *testing.T) {
	l := New(time.Minute)
	if l.Token == "" {
		t.Fatal("new lease token is empty")
	}
	if l.Generation != 1 {
		t.Fatalf("new lease generation = %d, want 1", l.Generation)
	}
	if l.IsExpired(time.Now().Add(30 * time.Second)) {
		t.Fatal("lease should not be expired before TTL")
	}
	if !l.IsExpired(time.Now().Add(2 * time.Minute)) {
		t.Fatal("lease should be expired after TTL")
	}
}

func TestLeaseEqualsRequiresTokenAndGeneration(t *testing.T) {
	a := Lease{Token: "a", Generation: 3}
	if !a.Equals(a) {
		t.Fatal("lease should equal itself")
	}
	b := Lease{Token: "a", Generation: 4}
	if a.Equals(b) {
		t.Fatal("different generation should not equal")
	}
	c := Lease{Token: "b", Generation: 3}
	if a.Equals(c) {
		t.Fatal("different token should not equal")
	}
}

func TestLeaseIsCurrentRejectsStaleGeneration(t *testing.T) {
	now := time.Now()
	a := Lease{Token: "tok", Generation: 5, ExpiresAt: now.Add(time.Minute)}
	b := Lease{Token: "tok", Generation: 6, ExpiresAt: now.Add(time.Minute)}
	if a.IsCurrent(b, now) {
		t.Fatal("stale generation should not be current")
	}
	if !a.IsCurrent(a, now) {
		t.Fatal("same lease should be current")
	}
}

func TestLeaseIsCurrentRejectsExpiredLease(t *testing.T) {
	now := time.Now()
	a := Lease{Token: "tok", Generation: 5, ExpiresAt: now.Add(-time.Second)}
	if a.IsCurrent(a, now) {
		t.Fatal("expired lease should not be current")
	}
}

func TestNextEpochBumpsTokenAndGeneration(t *testing.T) {
	a := New(time.Minute)
	b := a.NextEpoch(time.Minute)
	if b.Token == a.Token {
		t.Fatal("next epoch should have a new token")
	}
	if b.Generation != a.Generation+1 {
		t.Fatalf("next epoch generation = %d, want %d", b.Generation, a.Generation+1)
	}
}

func TestBumpGenerationKeepsToken(t *testing.T) {
	a := Lease{Token: "tok", Generation: 5, ExpiresAt: time.Now().Add(time.Minute)}
	b := a.BumpGeneration()
	if b.Token != a.Token {
		t.Fatal("bump generation should keep token")
	}
	if b.Generation != 6 {
		t.Fatalf("bump generation = %d, want 6", b.Generation)
	}
}

func TestRenewKeepsTokenAndGeneration(t *testing.T) {
	now := time.Now()
	a := Lease{Token: "tok", Generation: 5, ExpiresAt: now.Add(time.Minute)}
	b := a.Renew(5 * time.Minute)
	if b.Token != a.Token || b.Generation != a.Generation {
		t.Fatal("renew should keep token and generation")
	}
	if !b.ExpiresAt.After(a.ExpiresAt) {
		t.Fatal("renew should extend expiry")
	}
}

func TestZeroLease(t *testing.T) {
	var z Lease
	if !z.Zero() {
		t.Fatal("zero value should be zero")
	}
	if New(time.Minute).Zero() {
		t.Fatal("new lease should not be zero")
	}
}
