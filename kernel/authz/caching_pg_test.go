package authz_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/policy"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestIntegrationCachingStoreRevokeInvalidate is the CA-2(b) end-to-end proof
// for the actor-assignment (product-owned) write path: with the CachingStore
// enabled over the real PgStore, a role REVOKE is reflected immediately after the
// Invalidate handle fires — NOT after the TTL. A long TTL is used so expiry can
// never be the cause; only Invalidate can refresh the entry within the window.
func TestIntegrationCachingStoreRevokeInvalidate(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	seedTenant(t, h, ctx, tenant)
	cap1 := seedCapacity(t, h, ctx, tenant)

	seedPermission(t, h, ctx, "requests.request.approve")
	role := seedRole(t, h, ctx, &tenant, "approver")
	seedRolePerm(t, h, ctx, role, "requests.request.approve")
	seedAssignment(t, h, ctx, tenant, cap1, role, "tenant", nil, nil, "now() - interval '1 hour'", "NULL")

	reg := authz.NewRegistry()
	reg.Register(authz.Permission{Key: "requests.request.approve"})
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}

	cache := authz.NewCachingStore(authz.NewStore(), time.Hour)
	eval := authz.New(authz.Options{Store: cache, Registry: reg, Policies: policy.New()})
	actor := authz.Actor{Kind: authz.ActorUser, CapacityID: cap1, TenantID: tenant}

	allowed := func() bool {
		d := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) (authz.Decision, error) {
			return eval.Evaluate(ctx, db, actor, "requests.request.approve", authz.Target{Scope: authz.ScopeTenant})
		})
		return d.Allowed
	}

	// Warm the cache with an allow.
	if !allowed() {
		t.Fatal("precondition: the assignment must grant approve before revoke")
	}

	// Revoke the assignment in the DB. Within the (long) TTL and WITHOUT
	// invalidation the cached allow persists — bounded staleness, the documented
	// trade-off (and proof the decision really was cached).
	mustExec(t, h, ctx, `DELETE FROM actor_assignments WHERE capacity_id = $1`, cap1)
	if !allowed() {
		t.Fatal("within the TTL a revoke without Invalidate must be bounded-stale (still allowed)")
	}

	// Fire the exposed handle for this actor — the product's grant/revoke path.
	// The revoke now takes effect immediately, with no clock advance.
	cache.Invalidate(tenant, cap1)
	if allowed() {
		t.Fatal("after Invalidate the revoke must apply immediately (stale-allow!)")
	}
}

// TestIntegrationCachingOffRevokeImmediate proves the caching-off default is
// unbroken: with no CachingStore, a revoke is reflected on the very next
// Evaluate — there is nothing to invalidate and nothing goes stale.
func TestIntegrationCachingOffRevokeImmediate(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	seedTenant(t, h, ctx, tenant)
	cap1 := seedCapacity(t, h, ctx, tenant)

	seedPermission(t, h, ctx, "requests.request.approve")
	role := seedRole(t, h, ctx, &tenant, "approver")
	seedRolePerm(t, h, ctx, role, "requests.request.approve")
	seedAssignment(t, h, ctx, tenant, cap1, role, "tenant", nil, nil, "now() - interval '1 hour'", "NULL")

	reg := authz.NewRegistry()
	reg.Register(authz.Permission{Key: "requests.request.approve"})

	eval := authz.New(authz.Options{Store: authz.NewStore(), Registry: reg, Policies: policy.New()})
	actor := authz.Actor{Kind: authz.ActorUser, CapacityID: cap1, TenantID: tenant}
	allowed := func() bool {
		d := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) (authz.Decision, error) {
			return eval.Evaluate(ctx, db, actor, "requests.request.approve", authz.Target{Scope: authz.ScopeTenant})
		})
		return d.Allowed
	}

	if !allowed() {
		t.Fatal("precondition: assignment grants approve")
	}
	mustExec(t, h, ctx, `DELETE FROM actor_assignments WHERE capacity_id = $1`, cap1)
	if allowed() {
		t.Fatal("with caching off a revoke must be reflected on the next Evaluate")
	}
}
