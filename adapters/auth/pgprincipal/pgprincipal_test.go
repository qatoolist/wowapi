package pgprincipal_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/adapters/auth/pgprincipal"
	"github.com/qatoolist/wowapi/v2/kernel/auth"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// These are real DB integration tests (no mocks): they run against the migrated
// kernel schema over the app_platform (global users) and app_rt (RLS-scoped
// acting_capacities) pools, proving the role split and that cross-tenant
// capacities are invisible under RLS.

// seedUser inserts a global user with a known idp subject via the owner pool.
func seedUser(t *testing.T, h *testkit.DBHandle, subject, status string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := h.Admin.Exec(context.Background(),
		`INSERT INTO users (id, idp_subject, email, status, created_by) VALUES ($1,$2,$3,$4,$5)`,
		id, subject, uuid.NewString()[:8]+"@example.test", status, uuid.Nil)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func TestUserIDBySubject(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	subject := "idp|active-" + uuid.NewString()[:8]
	userID := seedUser(t, h, subject, "active")

	got, err := store.UserIDBySubject(ctx, subject)
	if err != nil {
		t.Fatalf("UserIDBySubject(active): %v", err)
	}
	if got != userID {
		t.Fatalf("user id: got %v want %v", got, userID)
	}

	// Unknown subject → opaque unauthenticated.
	if _, err := store.UserIDBySubject(ctx, "idp|nobody-"+uuid.NewString()[:8]); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("unknown subject: want KindUnauthenticated, got %v", err)
	}

	// Disabled user → unauthenticated with the SAME message (no oracle).
	disSubject := "idp|disabled-" + uuid.NewString()[:8]
	seedUser(t, h, disSubject, "disabled")
	if _, err := store.UserIDBySubject(ctx, disSubject); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("disabled subject: want KindUnauthenticated, got %v", err)
	}
}

// seedUserTenantAccess inserts a live user_tenant_access row for userID in tenant.
func seedUserTenantAccess(t *testing.T, h *testkit.DBHandle, userID, tenant uuid.UUID, status string, validTo any) {
	t.Helper()
	_, err := h.Admin.Exec(context.Background(),
		`INSERT INTO user_tenant_access (id, user_id, tenant_id, status, valid_to, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), userID, tenant, status, validTo, uuid.Nil)
	if err != nil {
		t.Fatalf("seed user_tenant_access: %v", err)
	}
}

func TestActiveTenantAccess(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	userID := seedUser(t, h, "idp|member-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, userID, tenant.ID, "active", nil)

	// Valid live membership.
	if err := store.ActiveTenantAccess(ctx, userID, tenant.ID); err != nil {
		t.Fatalf("ActiveTenantAccess(valid): %v", err)
	}

	// Unknown user → forbidden (no oracle distinguishing absent user from absent membership).
	if err := store.ActiveTenantAccess(ctx, uuid.New(), tenant.ID); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("unknown user: want KindForbidden, got %v", err)
	}

	// Revoked membership (valid_to set) → forbidden.
	revokedUser := seedUser(t, h, "idp|revoked-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, revokedUser, tenant.ID, "active", "2020-01-01")
	if err := store.ActiveTenantAccess(ctx, revokedUser, tenant.ID); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("revoked membership: want KindForbidden, got %v", err)
	}

	// Suspended membership status → forbidden.
	suspendedUser := seedUser(t, h, "idp|suspended-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, suspendedUser, tenant.ID, "suspended", nil)
	if err := store.ActiveTenantAccess(ctx, suspendedUser, tenant.ID); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("suspended membership: want KindForbidden, got %v", err)
	}

	// Foreign tenant: membership exists only in `tenant`, so a different tenant
	// sees no live row → forbidden.
	tenant2 := testkit.CreateTenant(t, h)
	if err := store.ActiveTenantAccess(ctx, userID, tenant2.ID); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("foreign-tenant membership: want KindForbidden, got %v", err)
	}
}

func TestValidateCapacity(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	userID := seedUser(t, h, "idp|cap-"+uuid.NewString()[:8], "active")
	capID := testkit.CreateCapacity(t, h, tenant.ID, userID)

	// Valid capacity for this user in this tenant.
	if err := store.ValidateCapacity(ctx, userID, tenant.ID, capID); err != nil {
		t.Fatalf("ValidateCapacity(valid): %v", err)
	}

	// Unknown capacity id → forbidden.
	if err := store.ValidateCapacity(ctx, userID, tenant.ID, uuid.New()); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("unknown capacity: want KindForbidden, got %v", err)
	}

	// Capacity for a different user → forbidden.
	otherUser := seedUser(t, h, "idp|other-"+uuid.NewString()[:8], "active")
	if err := store.ValidateCapacity(ctx, otherUser, tenant.ID, capID); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("foreign-user capacity: want KindForbidden, got %v", err)
	}

	// Cross-tenant: the capacity belongs to `tenant`, so under RLS it is invisible
	// when validating within a different tenant → forbidden (no cross-tenant leak).
	tenant2 := testkit.CreateTenant(t, h)
	if err := store.ValidateCapacity(ctx, userID, tenant2.ID, capID); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("cross-tenant capacity must be invisible: want KindForbidden, got %v", err)
	}
}

// TestIdentityGrantOneActivePerActorConcurrent proves that the partial unique
// index enforcing at most one active grant per actor holds under concurrent
// activation attempts: exactly one succeeds, the rest fail.
func TestIdentityGrantOneActivePerActorConcurrent(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	actor := uuid.New()

	const n = 8
	var wg sync.WaitGroup
	var successes int
	var mu sync.Mutex
	var errs []error

	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.Admin.Exec(ctx,
				`INSERT INTO identity_grant (id, status, tenant_id, actor_id, activated_at, expires_at)
				 VALUES ($1, 'active', $2, $3, now(), now() + interval '1 hour')`,
				uuid.New(), tenant.ID, actor)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
			} else {
				successes++
			}
		}()
	}
	wg.Wait()

	if successes != 1 {
		t.Fatalf("concurrent activations: got %d successes, want exactly 1 (errs: %v)", successes, errs)
	}

	var activeCount int
	if err := h.Admin.QueryRow(ctx,
		`SELECT count(*) FROM identity_grant WHERE actor_id = $1 AND status = 'active'`, actor).Scan(&activeCount); err != nil {
		t.Fatalf("count active grants: %v", err)
	}
	if activeCount != 1 {
		t.Fatalf("active grant count = %d, want 1", activeCount)
	}
}

// seedCapacity inserts an active acting capacity with an explicit label so
// tests can create multiple active capacities for the same user.
func seedCapacity(t *testing.T, h *testkit.DBHandle, tenantID, userID uuid.UUID, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := h.Admin.Exec(context.Background(),
		`INSERT INTO acting_capacities (id, tenant_id, user_id, label, created_by)
		 VALUES ($1, $2, $3, $4, $5)`, id, tenantID, userID, label, uuid.Nil)
	if err != nil {
		t.Fatalf("seed capacity: %v", err)
	}
	return id
}

// TestActiveCapacityCount proves the T4 capacity-count path: capacities are
// counted per user per tenant, and cross-tenant capacities are invisible under
// RLS.
func TestActiveCapacityCount(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	userID := seedUser(t, h, "idp|capcount-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, userID, tenant.ID, "active", nil)

	// No capacities yet.
	if n, err := store.ActiveCapacityCount(ctx, userID, tenant.ID); err != nil || n != 0 {
		t.Fatalf("count zero: got %d, %v", n, err)
	}

	// One capacity.
	cap1 := seedCapacity(t, h, tenant.ID, userID, "member")
	if n, err := store.ActiveCapacityCount(ctx, userID, tenant.ID); err != nil || n != 1 {
		t.Fatalf("count one: got %d, %v", n, err)
	}

	// Two capacities trigger T4 enforcement. Labels must differ because of the
	// partial unique index on (tenant_id, user_id, label) WHERE valid_to IS NULL.
	cap2 := seedCapacity(t, h, tenant.ID, userID, "admin")
	if n, err := store.ActiveCapacityCount(ctx, userID, tenant.ID); err != nil || n != 2 {
		t.Fatalf("count two: got %d, %v", n, err)
	}

	// A capacity from another tenant is invisible under RLS.
	tenant2 := testkit.CreateTenant(t, h)
	otherUser := seedUser(t, h, "idp|capcount-other-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, otherUser, tenant2.ID, "active", nil)
	seedCapacity(t, h, tenant2.ID, otherUser, "member")
	if n, err := store.ActiveCapacityCount(ctx, userID, tenant.ID); err != nil || n != 2 {
		t.Fatalf("count after cross-tenant seed: got %d, %v", n, err)
	}

	// Validating an explicit capacity still works.
	if err := store.ValidateCapacity(ctx, userID, tenant.ID, cap1); err != nil {
		t.Fatalf("ValidateCapacity(cap1): %v", err)
	}
	if err := store.ValidateCapacity(ctx, userID, tenant.ID, cap2); err != nil {
		t.Fatalf("ValidateCapacity(cap2): %v", err)
	}
	// cap1 from tenant is invisible in tenant2.
	if err := store.ValidateCapacity(ctx, userID, tenant2.ID, cap1); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("cross-tenant capacity visible: got %v", err)
	}
}

// seedIdentityGrant inserts an identity_grant row via the admin pool
// (app_platform), since app_rt has no grants on the table.
func seedIdentityGrant(t *testing.T, h *testkit.DBHandle, id, tenantID, actorID uuid.UUID, status string, impersonatedUserID, approverID *uuid.UUID, expiresAt time.Time) {
	t.Helper()
	_, err := h.Admin.Exec(context.Background(),
		`INSERT INTO identity_grant (id, status, tenant_id, actor_id, impersonated_user_id, approver_id, activated_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, now(), $7)`,
		id, status, tenantID, actorID, impersonatedUserID, approverID, expiresAt)
	if err != nil {
		t.Fatalf("seed identity_grant: %v", err)
	}
}

// TestResolveGrant_ImpersonationSuccess proves that an active impersonation
// grant maps the impersonated user to Actor.UserID and the support actor to
// Actor.ImpersonatorUserID.
func TestResolveGrant_ImpersonationSuccess(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	actor := seedUser(t, h, "idp|imp-actor-"+uuid.NewString()[:8], "active")
	target := seedUser(t, h, "idp|imp-target-"+uuid.NewString()[:8], "active")
	approver := seedUser(t, h, "idp|imp-approver-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, actor, tenant.ID, "active", nil)
	seedUserTenantAccess(t, h, target, tenant.ID, "active", nil)
	seedUserTenantAccess(t, h, approver, tenant.ID, "active", nil)

	grantID := uuid.New()
	seedIdentityGrant(t, h, grantID, tenant.ID, actor, "active", &target, &approver, time.Now().Add(time.Hour))

	grant, err := store.ResolveGrant(ctx, target, tenant.ID, grantID)
	if err != nil {
		t.Fatalf("ResolveGrant: %v", err)
	}
	if grant.ImpersonatorUserID != actor {
		t.Fatalf("impersonator: got %v want %v", grant.ImpersonatorUserID, actor)
	}
	if grant.BreakGlass {
		t.Fatalf("impersonation grant must not set break-glass")
	}
}

// TestResolveGrant_BreakGlassSuccess proves that an active break-glass grant
// with no impersonated_user_id sets BreakGlass=true and verifies against the
// actor.
func TestResolveGrant_BreakGlassSuccess(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	actor := seedUser(t, h, "idp|bg-actor-"+uuid.NewString()[:8], "active")
	approver := seedUser(t, h, "idp|bg-approver-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, actor, tenant.ID, "active", nil)
	seedUserTenantAccess(t, h, approver, tenant.ID, "active", nil)

	grantID := uuid.New()
	seedIdentityGrant(t, h, grantID, tenant.ID, actor, "active", nil, &approver, time.Now().Add(time.Hour))

	grant, err := store.ResolveGrant(ctx, actor, tenant.ID, grantID)
	if err != nil {
		t.Fatalf("ResolveGrant: %v", err)
	}
	if !grant.BreakGlass {
		t.Fatalf("expected break-glass true")
	}
	if grant.ImpersonatorUserID != uuid.Nil {
		t.Fatalf("break-glass must not set impersonator")
	}
}

// TestResolveGrant_ExpiredRejection proves that a grant whose expires_at has
// passed is rejected as GrantRejectionExpired.
func TestResolveGrant_ExpiredRejection(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	actor := seedUser(t, h, "idp|bg-exp-"+uuid.NewString()[:8], "active")
	approver := seedUser(t, h, "idp|bg-exp-approver-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, actor, tenant.ID, "active", nil)
	seedUserTenantAccess(t, h, approver, tenant.ID, "active", nil)

	grantID := uuid.New()
	seedIdentityGrant(t, h, grantID, tenant.ID, actor, "active", nil, &approver, time.Now().Add(-time.Hour))

	_, err := store.ResolveGrant(ctx, actor, tenant.ID, grantID)
	if !auth.IsGrantRejection(err, auth.GrantRejectionExpired) {
		t.Fatalf("want GrantRejectionExpired, got %v", err)
	}
}

// TestResolveGrant_RevokedRejection proves that a grant with status 'revoked'
// is rejected as GrantRejectionRevoked.
func TestResolveGrant_RevokedRejection(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	actor := seedUser(t, h, "idp|bg-rev-"+uuid.NewString()[:8], "active")
	approver := seedUser(t, h, "idp|bg-rev-approver-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, actor, tenant.ID, "active", nil)
	seedUserTenantAccess(t, h, approver, tenant.ID, "active", nil)

	grantID := uuid.New()
	seedIdentityGrant(t, h, grantID, tenant.ID, actor, "revoked", nil, &approver, time.Now().Add(time.Hour))

	_, err := store.ResolveGrant(ctx, actor, tenant.ID, grantID)
	if !auth.IsGrantRejection(err, auth.GrantRejectionRevoked) {
		t.Fatalf("want GrantRejectionRevoked, got %v", err)
	}
}

// TestResolveGrant_WrongTenantRejection proves that a grant from a different
// tenant is rejected as GrantRejectionWrongTenant (via the tenant_id filter in
// the lookup, so cross-tenant grant IDs are not oracle-leaky).
func TestResolveGrant_WrongTenantRejection(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	tenant1 := testkit.CreateTenant(t, h)
	tenant2 := testkit.CreateTenant(t, h)
	actor := seedUser(t, h, "idp|bg-wt-"+uuid.NewString()[:8], "active")
	approver := seedUser(t, h, "idp|bg-wt-approver-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, actor, tenant1.ID, "active", nil)
	seedUserTenantAccess(t, h, approver, tenant1.ID, "active", nil)

	grantID := uuid.New()
	seedIdentityGrant(t, h, grantID, tenant1.ID, actor, "active", nil, &approver, time.Now().Add(time.Hour))

	_, err := store.ResolveGrant(ctx, actor, tenant2.ID, grantID)
	if !auth.IsGrantRejection(err, auth.GrantRejectionWrongTenant) {
		t.Fatalf("want GrantRejectionWrongTenant, got %v", err)
	}
}

// TestResolveGrant_WrongActorRejection proves that a break-glass grant issued to
// a different actor is rejected as GrantRejectionWrongActor.
func TestResolveGrant_WrongActorRejection(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	actor := seedUser(t, h, "idp|bg-wa-actor-"+uuid.NewString()[:8], "active")
	other := seedUser(t, h, "idp|bg-wa-other-"+uuid.NewString()[:8], "active")
	approver := seedUser(t, h, "idp|bg-wa-approver-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, actor, tenant.ID, "active", nil)
	seedUserTenantAccess(t, h, other, tenant.ID, "active", nil)
	seedUserTenantAccess(t, h, approver, tenant.ID, "active", nil)

	grantID := uuid.New()
	seedIdentityGrant(t, h, grantID, tenant.ID, actor, "active", nil, &approver, time.Now().Add(time.Hour))

	_, err := store.ResolveGrant(ctx, other, tenant.ID, grantID)
	if !auth.IsGrantRejection(err, auth.GrantRejectionWrongActor) {
		t.Fatalf("want GrantRejectionWrongActor, got %v", err)
	}
}

// TestResolveGrant_NotFoundRejection proves that a forged/unknown grant ID is
// rejected as GrantRejectionNotFound.
func TestResolveGrant_NotFoundRejection(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)
	userID := seedUser(t, h, "idp|bg-nf-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, userID, tenant.ID, "active", nil)

	_, err := store.ResolveGrant(ctx, userID, tenant.ID, uuid.New())
	if !auth.IsGrantRejection(err, auth.GrantRejectionNotFound) {
		t.Fatalf("want GrantRejectionNotFound, got %v", err)
	}
}

// TestResolveGrant_UnauthorizedApproverRejection proves three unauthorized-
// approver conditions: missing approver, self-approval, and an approver who is
// not a tenant member. All are rejected as GrantRejectionUnauthorizedApprover.
// Each sub-case uses a distinct actor because identity_grant enforces at most
// one active grant per actor.
func TestResolveGrant_UnauthorizedApproverRejection(t *testing.T) {
	h := testkit.NewDB(t)
	store := pgprincipal.New(h.PlatformTxM, h.TxM)
	ctx := context.Background()

	tenant := testkit.CreateTenant(t, h)

	// Missing approver.
	actor1 := seedUser(t, h, "idp|bg-ua-actor1-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, actor1, tenant.ID, "active", nil)
	grantID := uuid.New()
	seedIdentityGrant(t, h, grantID, tenant.ID, actor1, "active", nil, nil, time.Now().Add(time.Hour))
	_, err := store.ResolveGrant(ctx, actor1, tenant.ID, grantID)
	if !auth.IsGrantRejection(err, auth.GrantRejectionUnauthorizedApprover) {
		t.Fatalf("missing approver: want GrantRejectionUnauthorizedApprover, got %v", err)
	}

	// Self-approval: approver is the actor.
	actor2 := seedUser(t, h, "idp|bg-ua-actor2-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, actor2, tenant.ID, "active", nil)
	grantID = uuid.New()
	selfApprover := actor2
	seedIdentityGrant(t, h, grantID, tenant.ID, actor2, "active", nil, &selfApprover, time.Now().Add(time.Hour))
	_, err = store.ResolveGrant(ctx, actor2, tenant.ID, grantID)
	if !auth.IsGrantRejection(err, auth.GrantRejectionUnauthorizedApprover) {
		t.Fatalf("self-approval: want GrantRejectionUnauthorizedApprover, got %v", err)
	}

	// Approver is a real user but lacks tenant membership.
	actor3 := seedUser(t, h, "idp|bg-ua-actor3-"+uuid.NewString()[:8], "active")
	seedUserTenantAccess(t, h, actor3, tenant.ID, "active", nil)
	foreignApprover := seedUser(t, h, "idp|bg-ua-foreign-"+uuid.NewString()[:8], "active")
	grantID = uuid.New()
	seedIdentityGrant(t, h, grantID, tenant.ID, actor3, "active", nil, &foreignApprover, time.Now().Add(time.Hour))
	_, err = store.ResolveGrant(ctx, actor3, tenant.ID, grantID)
	if !auth.IsGrantRejection(err, auth.GrantRejectionUnauthorizedApprover) {
		t.Fatalf("foreign approver: want GrantRejectionUnauthorizedApprover, got %v", err)
	}
}
