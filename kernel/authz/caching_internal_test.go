package authz

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// countingStore counts ActiveAssignments calls and lets a test mutate the
// returned assignments to simulate a grant/revoke.
type countingStore struct {
	calls int
	asgs  []Assignment
}

func (c *countingStore) ActiveAssignments(context.Context, database.TenantDB, Actor, time.Time) ([]Assignment, error) {
	c.calls++
	return c.asgs, nil
}
func (c *countingStore) OrgAncestors(context.Context, database.TenantDB, uuid.UUID) ([]uuid.UUID, error) {
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
