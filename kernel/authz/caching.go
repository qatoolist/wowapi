package authz

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/resource"
)

// CachingStore is an OPT-IN decorator that caches the hot authorization read —
// ActiveAssignments — per (tenant, actor) for a short TTL, so a burst of requests
// from one principal does not hit the DB on every Evaluate (roadmap R1). It is a
// pure add-on: unwrapped, the evaluator behaves exactly as before.
//
// Correctness against stale-allow: a role grant/revoke MUST call Invalidate (or
// InvalidateTenant) so the change takes effect immediately on this pod; the TTL
// is only the cross-pod bound (invalidation is in-process). The other Store reads
// (org ancestry, policies, resource org) pass straight through — they are not on
// the per-actor hot path and caching them would widen the invalidation surface.
//
// Read-replica routing (the second half of R1) is a deployment concern: point the
// Manager's read-only path (WithTenantRO) at a replica pool; the evaluator already
// runs its reads in that read-only transaction.
type CachingStore struct {
	inner Store
	ttl   time.Duration
	now   func() time.Time

	mu    sync.Mutex
	cache map[string]cachedAssignments
}

type cachedAssignments struct {
	at   time.Time
	asgs []Assignment
}

// NewCachingStore wraps inner with a per-actor ActiveAssignments cache of the
// given TTL. A TTL <= 0 defaults to 1s (keep it short — it bounds cross-pod
// staleness after a revocation on another pod).
func NewCachingStore(inner Store, ttl time.Duration) *CachingStore {
	return newCachingStore(inner, ttl, time.Now)
}

func newCachingStore(inner Store, ttl time.Duration, now func() time.Time) *CachingStore {
	if ttl <= 0 {
		ttl = time.Second
	}
	return &CachingStore{inner: inner, ttl: ttl, now: now, cache: map[string]cachedAssignments{}}
}

func actorCacheKey(a Actor) string {
	switch {
	case a.CapacityID != uuid.Nil:
		return a.TenantID.String() + "|c:" + a.CapacityID.String()
	case a.UserID != uuid.Nil:
		return a.TenantID.String() + "|u:" + a.UserID.String()
	default:
		return a.TenantID.String() + "|s:" + a.System
	}
}

// ActiveAssignments returns the actor's assignments from cache when fresh, else
// loads and caches them. Returned slices are cloned so a caller cannot mutate the
// cached entry.
func (c *CachingStore) ActiveAssignments(ctx context.Context, db database.TenantDB, a Actor, at time.Time) ([]Assignment, error) {
	key := actorCacheKey(a)
	c.mu.Lock()
	if e, ok := c.cache[key]; ok && c.now().Sub(e.at) < c.ttl {
		out := slices.Clone(e.asgs)
		c.mu.Unlock()
		return out, nil
	}
	c.mu.Unlock()

	asgs, err := c.inner.ActiveAssignments(ctx, db, a, at)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.cache[key] = cachedAssignments{at: c.now(), asgs: slices.Clone(asgs)}
	c.mu.Unlock()
	return slices.Clone(asgs), nil
}

// Invalidate drops one actor's cached assignments — call it right after changing
// that actor's role assignments so a revocation takes effect immediately.
func (c *CachingStore) Invalidate(tenantID, capacityOrUserID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, tenantID.String()+"|c:"+capacityOrUserID.String())
	delete(c.cache, tenantID.String()+"|u:"+capacityOrUserID.String())
}

// InvalidateTenant drops every cached actor for a tenant — for a bulk role change
// (role definition edit, mass revoke).
func (c *CachingStore) InvalidateTenant(tenantID uuid.UUID) {
	prefix := tenantID.String() + "|"
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.cache {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(c.cache, k)
		}
	}
}

// InvalidateAll drops the entire cache. It is the correct invalidation for a
// GLOBAL authorization-spine write — a seed sync of platform roles / their
// role_permissions — because those rows are cross-tenant: a changed role and its
// grants may be held by actors in ANY tenant, and the cached ActiveAssignments
// pre-join role_permissions, so a permission added to (or pruned from) a role is
// otherwise served stale until the TTL. seeds.Sync calls this after its writes
// commit when a live cache is wired (CA-2), so a spine change takes effect on
// this pod immediately rather than after the TTL.
func (c *CachingStore) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	clear(c.cache)
}

// --- pass-through reads (not on the per-actor hot path) ---

func (c *CachingStore) OrgAncestors(ctx context.Context, db database.TenantDB, orgID uuid.UUID) ([]uuid.UUID, error) {
	return c.inner.OrgAncestors(ctx, db, orgID)
}

func (c *CachingStore) OrgSubtree(ctx context.Context, db database.TenantDB, orgID uuid.UUID) ([]uuid.UUID, error) {
	return c.inner.OrgSubtree(ctx, db, orgID)
}

func (c *CachingStore) Policies(ctx context.Context, db database.TenantDB, a Actor, perm, rt string) ([]Policy, error) {
	return c.inner.Policies(ctx, db, a, perm, rt)
}

func (c *CachingStore) ResourceOrg(ctx context.Context, db database.TenantDB, ref resource.Ref) (uuid.UUID, error) {
	return c.inner.ResourceOrg(ctx, db, ref)
}
