package authz

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/resource"
)

func resourceRef(id uuid.UUID) resource.Ref {
	return resource.Ref{Type: "requests.request", ID: id}
}

// recordingStore is a Store whose pass-through reads are canned and counted, so
// the CachingStore's delegation of the non-hot-path reads is observable.
type recordingStore struct {
	ancestors []uuid.UUID
	subtree   []uuid.UUID
	policies  []Policy
	resOrg    uuid.UUID

	ancestorsCalls int
	subtreeCalls   int
	policiesCalls  int
	resOrgCalls    int
}

func (r *recordingStore) ActiveAssignments(context.Context, database.TenantDB, Actor, time.Time) ([]Assignment, error) {
	return nil, nil
}

func (r *recordingStore) OrgAncestors(context.Context, database.TenantDB, uuid.UUID) ([]uuid.UUID, error) {
	r.ancestorsCalls++
	return r.ancestors, nil
}

func (r *recordingStore) OrgSubtree(context.Context, database.TenantDB, uuid.UUID) ([]uuid.UUID, error) {
	r.subtreeCalls++
	return r.subtree, nil
}

func (r *recordingStore) Policies(context.Context, database.TenantDB, Actor, string, string) ([]Policy, error) {
	r.policiesCalls++
	return r.policies, nil
}

func (r *recordingStore) ResourceOrg(context.Context, database.TenantDB, resource.Ref) (uuid.UUID, error) {
	r.resOrgCalls++
	return r.resOrg, nil
}

// TestNewCachingStoreDefaultTTL exercises the public constructor and the
// ttl<=0 → 1s default: a non-positive TTL is clamped so the cache never
// degenerates into "never fresh".
func TestNewCachingStoreDefaultTTL(t *testing.T) {
	inner := &countingStore{}
	if c := NewCachingStore(inner, 0); c.ttl != time.Second {
		t.Fatalf("ttl<=0 must default to 1s, got %v", c.ttl)
	}
	if c := NewCachingStore(inner, -5*time.Second); c.ttl != time.Second {
		t.Fatalf("negative ttl must default to 1s, got %v", c.ttl)
	}
	if c := NewCachingStore(inner, 250*time.Millisecond); c.ttl != 250*time.Millisecond {
		t.Fatalf("positive ttl must be kept, got %v", c.ttl)
	}
}

// TestActorCacheKeyVariants drives ActiveAssignments with the three actor
// shapes so every actorCacheKey branch (capacity, user-only, system) is keyed
// distinctly and cached independently.
func TestActorCacheKeyVariants(t *testing.T) {
	clk := &fakeClk{t: time.Unix(2000, 0)}
	inner := &countingStore{asgs: []Assignment{{RoleKey: "r", Perms: []string{"p"}}}}
	c := newCachingStore(inner, time.Second, clk.now)
	tenant := uuid.New()

	cap := Actor{TenantID: tenant, CapacityID: uuid.New(), UserID: uuid.New()}
	user := Actor{TenantID: tenant, UserID: uuid.New()}
	sys := Actor{TenantID: tenant, System: "outbox-relay"}

	read := func(a Actor) {
		if _, err := c.ActiveAssignments(context.Background(), nil, a, clk.now()); err != nil {
			t.Fatal(err)
		}
	}
	// First read of each distinct principal misses; the repeat hits the cache.
	read(cap)
	read(user)
	read(sys)
	if inner.calls != 3 {
		t.Fatalf("three distinct principals must each miss once: calls=%d, want 3", inner.calls)
	}
	read(cap)
	read(user)
	read(sys)
	if inner.calls != 3 {
		t.Fatalf("repeat reads within TTL must all hit cache: calls=%d, want 3", inner.calls)
	}
	// Distinct keys: cache holds three entries.
	if len(c.cache) != 3 {
		t.Fatalf("cache must hold one entry per distinct actor key, got %d", len(c.cache))
	}
}

// TestInvalidateTenantDropsAllTenantActors covers the bulk invalidation path: a
// tenant-wide role change must evict every cached principal in that tenant while
// leaving another tenant's entries untouched.
func TestInvalidateTenantDropsAllTenantActors(t *testing.T) {
	clk := &fakeClk{t: time.Unix(3000, 0)}
	inner := &countingStore{asgs: []Assignment{{RoleKey: "r", Perms: []string{"p"}}}}
	c := newCachingStore(inner, time.Hour, clk.now)
	tA, tB := uuid.New(), uuid.New()
	a1 := Actor{TenantID: tA, CapacityID: uuid.New()}
	a2 := Actor{TenantID: tA, UserID: uuid.New()}
	other := Actor{TenantID: tB, CapacityID: uuid.New()}

	read := func(a Actor) {
		if _, err := c.ActiveAssignments(context.Background(), nil, a, clk.now()); err != nil {
			t.Fatal(err)
		}
	}
	read(a1)
	read(a2)
	read(other)
	if inner.calls != 3 {
		t.Fatalf("warmup: calls=%d, want 3", inner.calls)
	}

	c.InvalidateTenant(tA)

	// Both tenant-A principals must reload; the tenant-B principal stays cached.
	read(a1)
	read(a2)
	if inner.calls != 5 {
		t.Fatalf("both tenant-A actors must reload after InvalidateTenant: calls=%d, want 5", inner.calls)
	}
	read(other)
	if inner.calls != 5 {
		t.Fatalf("another tenant must be unaffected by InvalidateTenant: calls=%d, want 5", inner.calls)
	}
}

// TestCachingStorePassThroughReads asserts the non-hot-path reads delegate
// straight to the inner store (they are intentionally not cached).
func TestCachingStorePassThroughReads(t *testing.T) {
	orgID := uuid.New()
	inner := &recordingStore{
		ancestors: []uuid.UUID{orgID},
		subtree:   []uuid.UUID{orgID, uuid.New()},
		policies:  []Policy{{Key: "p"}},
		resOrg:    orgID,
	}
	c := NewCachingStore(inner, time.Second)
	ctx := context.Background()

	anc, err := c.OrgAncestors(ctx, nil, orgID)
	if err != nil || len(anc) != 1 || anc[0] != orgID {
		t.Fatalf("OrgAncestors passthrough = %v, %v", anc, err)
	}
	sub, err := c.OrgSubtree(ctx, nil, orgID)
	if err != nil || len(sub) != 2 {
		t.Fatalf("OrgSubtree passthrough = %v, %v", sub, err)
	}
	pols, err := c.Policies(ctx, nil, Actor{}, "perm", "rt")
	if err != nil || len(pols) != 1 || pols[0].Key != "p" {
		t.Fatalf("Policies passthrough = %v, %v", pols, err)
	}
	got, err := c.ResourceOrg(ctx, nil, resourceRef(orgID))
	if err != nil || got != orgID {
		t.Fatalf("ResourceOrg passthrough = %v, %v", got, err)
	}
	if inner.ancestorsCalls != 1 || inner.subtreeCalls != 1 || inner.policiesCalls != 1 || inner.resOrgCalls != 1 {
		t.Fatalf("each passthrough must hit inner exactly once: %+v", inner)
	}
}

// TestLastDotNoDot covers the defensive no-dot branch of lastDot (unreachable
// through Register, which regex-gates the key, but part of the helper contract).
func TestLastDotNoDot(t *testing.T) {
	if got := lastDot("nodothere"); got != -1 {
		t.Fatalf("lastDot(no dot) = %d, want -1", got)
	}
	if got := lastDot("a.b.c"); got != 3 {
		t.Fatalf("lastDot(a.b.c) = %d, want 3", got)
	}
}
