package authz

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// countingStore counts ActiveAssignments/OrgAncestors calls and lets a test
// mutate the returned assignments to simulate a grant/revoke.
type countingStore struct {
	calls        int
	orgAncestors int
	asgs         []Assignment
}

func (c *countingStore) ActiveAssignments(context.Context, database.TenantDB, Actor, time.Time) ([]Assignment, error) {
	c.calls++
	return c.asgs, nil
}

func (c *countingStore) OrgAncestors(context.Context, database.TenantDB, uuid.UUID) ([]uuid.UUID, error) {
	c.orgAncestors++
	return nil, nil
}

func (c *countingStore) OrgSubtree(context.Context, database.TenantDB, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (c *countingStore) Policies(context.Context, database.TenantDB, Actor, string, string) ([]Policy, error) {
	return nil, nil
}

func (c *countingStore) ResourceOrg(context.Context, database.TenantDB, resource.Ref) (uuid.UUID, error) {
	return uuid.Nil, nil
}

type fakeClk struct{ t time.Time }

func (c *fakeClk) now() time.Time { return c.t }

func TestCachingStoreHitInvalidateTTL(t *testing.T) {
	clk := &fakeClk{t: time.Unix(1000, 0)}
	inner := &countingStore{asgs: []Assignment{{RoleKey: "r", Perms: []string{"p"}}}}
	c := newCachingStore(inner, time.Second, clk.now)
	a := Actor{TenantID: uuid.New(), CapacityID: uuid.New()}
	read := func() []Assignment {
		got, err := c.ActiveAssignments(context.Background(), nil, a, clk.now())
		if err != nil {
			t.Fatal(err)
		}
		return got
	}

	// Two reads within the TTL → the DB (inner) is hit exactly once.
	read()
	read()
	if inner.calls != 1 {
		t.Fatalf("inner called %d times, want 1 (second read served from cache)", inner.calls)
	}

	// Simulate a revoke on the DB. Within the TTL and WITHOUT invalidation the old
	// grant is still served (bounded staleness — the documented trade-off).
	inner.asgs = nil
	if len(read()) != 1 {
		t.Fatal("within TTL a revoke without Invalidate is bounded-stale (still cached)")
	}

	// Explicit invalidation makes the revoke take effect immediately — no stale
	// allow (the R1 correctness requirement).
	c.Invalidate(a.TenantID, a.CapacityID)
	if got := read(); len(got) != 0 {
		t.Fatalf("after Invalidate the revoke must apply, got %d assignments (stale-allow!)", len(got))
	}
	if inner.calls != 2 {
		t.Fatalf("inner called %d times, want 2 (invalidation forced a reload)", inner.calls)
	}

	// A fresh grant plus TTL expiry also refreshes without an explicit invalidate.
	inner.asgs = []Assignment{{RoleKey: "r2", Perms: []string{"p"}}}
	clk.t = clk.t.Add(2 * time.Second)
	if len(read()) != 1 {
		t.Fatal("after the TTL elapses the cache must reload")
	}
}

// TestCachingStoreOrgAncestorsRoutesToComposedInner is the sentinel dependency
// injection proof for AR-06 T1: OrgAncestors on the composed store (the
// CachingStore decorator) must route to the SAME inner instance a caller
// wraps it around — never a fresh, independently constructed Store. This is
// the exact seam kernel.New's orgAncestry closure now closes over
// (kernel/kernel.go), instead of calling authz.NewStore() a second time.
func TestCachingStoreOrgAncestorsRoutesToComposedInner(t *testing.T) {
	sentinel := &countingStore{}
	c := newCachingStore(sentinel, time.Second, time.Now)

	if _, err := c.OrgAncestors(context.Background(), nil, uuid.New()); err != nil {
		t.Fatalf("OrgAncestors: %v", err)
	}
	if sentinel.orgAncestors != 1 {
		t.Fatalf("sentinel inner store's OrgAncestors called %d times, want 1 — "+
			"the composed store must delegate to its wrapped instance, not a fresh one",
			sentinel.orgAncestors)
	}

	// A second, independently constructed Store must NOT observe the call —
	// proving the decorator is instance-bound, not type-bound.
	other := &countingStore{}
	if other.orgAncestors != 0 {
		t.Fatalf("unrelated store instance saw %d calls, want 0", other.orgAncestors)
	}
}
