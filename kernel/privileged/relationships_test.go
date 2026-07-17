package privileged_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	kaudit "github.com/qatoolist/wowapi/v2/kernel/audit"
	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/privileged"
	"github.com/qatoolist/wowapi/v2/kernel/relationship"
	"github.com/qatoolist/wowapi/v2/kernel/resource"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// These integration tests pin, framework-side, the behaviour the product
// SECURITY DEFINER bridge identity_grant/revoke_committee_seat provided (GAP-006):
// tenant binding, ownership, subject/object existence, tenant isolation, and
// the soft-revoke / double-revoke semantics — proving a product no longer needs
// its own SECURITY DEFINER function to grant/revoke owned edges.

const seatRelType = "committee.seat_of"

func seedSeatRelType(t *testing.T, h *testkit.DBHandle) {
	t.Helper()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		 VALUES ($1,'committee','capacity','resource','seat') ON CONFLICT (key) DO NOTHING`,
		seatRelType); err != nil {
		t.Fatalf("seed rel type: %v", err)
	}
}

func authzActor(cap, tenant uuid.UUID) authz.Actor {
	return authz.Actor{Kind: authz.ActorUser, CapacityID: cap, TenantID: tenant}
}

// newRelSvc builds a privileged service bound to the "committee" module over the
// tenant-bindable app_platform manager — the production wiring.
func newRelSvc(h *testkit.DBHandle) *privileged.Relationships {
	svc := privileged.New("committee", h.PlatformTxM, nil, kaudit.New(model.UUIDv7(), nil), model.UUIDv7(), privileged.Config{})
	return svc.Relationships()
}

func TestIntegrationGrantThenCheckerSees(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	seedSeatRelType(t, h)
	user := testkit.CreateUser(t, h)
	cap1 := testkit.CreateCapacity(t, h, tenant, user)
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")

	rel := newRelSvc(h)
	ctx := testkit.TenantCtx(tenant)
	// Omit SubjectKind: it defaults to capacity (the identity a human actor carries).
	id, err := rel.Grant(ctx, privileged.GrantSpec{
		RelType: seatRelType, SubjectID: cap1, Object: obj, Actor: user,
	})
	if err != nil {
		t.Fatalf("Grant: %v", err)
	}
	if id == uuid.Nil {
		t.Fatal("Grant returned a nil id")
	}

	// The ReBAC checker (app_rt, read-only) now sees the edge.
	checker := relationship.NewChecker()
	actor := authzActor(cap1, tenant)
	var has bool
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		has, e = checker.Has(ctx, db, actor, seatRelType, obj, time.Now().Add(time.Minute))
		return e
	}); err != nil {
		t.Fatalf("Has: %v", err)
	}
	if !has {
		t.Fatal("granted edge is not visible to the ReBAC checker")
	}
}

func TestIntegrationGrantOwnershipPrivilegeDenied(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	// A relationship type owned by a DIFFERENT module ("core").
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		 VALUES ('core.owner_of','core','capacity','resource','owner') ON CONFLICT (key) DO NOTHING`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	user := testkit.CreateUser(t, h)
	cap1 := testkit.CreateCapacity(t, h, tenant, user)
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")

	rel := newRelSvc(h) // bound to "committee"
	_, err := rel.Grant(testkit.TenantCtx(tenant), privileged.GrantSpec{
		RelType: "core.owner_of", SubjectKind: relationship.KindCapacity, SubjectID: cap1, Object: obj,
	})
	if kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("want ownership_denied (Forbidden), got %v", err)
	}
}

func TestIntegrationGrantMissingSubjectRejected(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	seedSeatRelType(t, h)
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")

	rel := newRelSvc(h)
	_, err := rel.Grant(testkit.TenantCtx(tenant), privileged.GrantSpec{
		RelType: seatRelType, SubjectKind: relationship.KindCapacity, SubjectID: uuid.New(), Object: obj,
	})
	if kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("want subject_not_found (NotFound), got %v", err)
	}
}

func TestIntegrationGrantMissingObjectRejected(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	seedSeatRelType(t, h)
	user := testkit.CreateUser(t, h)
	cap1 := testkit.CreateCapacity(t, h, tenant, user)
	testkit.CreateResourceType(t, h, "committee.committee")

	rel := newRelSvc(h)
	_, err := rel.Grant(testkit.TenantCtx(tenant), privileged.GrantSpec{
		RelType: seatRelType, SubjectKind: relationship.KindCapacity, SubjectID: cap1,
		Object: resource.Ref{Type: "committee.committee", ID: uuid.New()},
	})
	if kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("want object_not_found (NotFound), got %v", err)
	}
}

func TestIntegrationGrantForeignTenantSubjectIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	tenantA := testkit.CreateTenant(t, h).ID
	tenantB := testkit.CreateTenant(t, h).ID
	seedSeatRelType(t, h)
	user := testkit.CreateUser(t, h)
	capB := testkit.CreateCapacity(t, h, tenantB, user) // capacity lives in B
	objA := testkit.CreateResourceTypeAndResource(t, h, tenantA, "committee.committee")

	rel := newRelSvc(h)
	// Grant in tenant A naming tenant B's capacity: RLS hides capB, so NotFound.
	_, err := rel.Grant(testkit.TenantCtx(tenantA), privileged.GrantSpec{
		RelType: seatRelType, SubjectKind: relationship.KindCapacity, SubjectID: capB, Object: objA,
	})
	if kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("cross-tenant subject must be rejected, got %v", err)
	}
}

func TestIntegrationGrantNoTenantFailsClosed(t *testing.T) {
	h := testkit.NewDB(t)
	rel := newRelSvc(h)
	_, err := rel.Grant(context.Background(), privileged.GrantSpec{
		RelType: seatRelType, SubjectID: uuid.New(),
		Object: resource.Ref{Type: "committee.committee", ID: uuid.New()},
	})
	if kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("want no_tenant (Unauthenticated), got %v", err)
	}
}

func TestIntegrationGrantInvalidWindowRejected(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	seedSeatRelType(t, h)
	user := testkit.CreateUser(t, h)
	cap1 := testkit.CreateCapacity(t, h, tenant, user)
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")

	from := time.Now()
	to := from.Add(-time.Hour)
	rel := newRelSvc(h)
	_, err := rel.Grant(testkit.TenantCtx(tenant), privileged.GrantSpec{
		RelType: seatRelType, SubjectKind: relationship.KindCapacity, SubjectID: cap1, Object: obj,
		ValidFrom: from, ValidTo: &to,
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("want invalid_window (Validation), got %v", err)
	}
}

func TestIntegrationRevokeThenCheckerBlind(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	seedSeatRelType(t, h)
	user := testkit.CreateUser(t, h)
	cap1 := testkit.CreateCapacity(t, h, tenant, user)
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")

	rel := newRelSvc(h)
	ctx := testkit.TenantCtx(tenant)
	id, err := rel.Grant(ctx, privileged.GrantSpec{
		RelType: seatRelType, SubjectKind: relationship.KindCapacity, SubjectID: cap1, Object: obj,
	})
	if err != nil {
		t.Fatalf("Grant: %v", err)
	}
	// Revoke with a nil actor id (unattributed) — updated_by falls to NULL.
	if err := rel.Revoke(ctx, id, uuid.Nil); err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	// After revoke, the checker no longer sees the edge (valid_to <= now).
	checker := relationship.NewChecker()
	var has bool
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		has, e = checker.Has(ctx, db, authzActor(cap1, tenant), seatRelType, obj, time.Now().Add(time.Minute))
		return e
	}); err != nil {
		t.Fatalf("Has: %v", err)
	}
	if has {
		t.Fatal("revoked edge must not be visible to the checker")
	}

	// The historical row still exists (soft revoke, not delete).
	var cnt int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM relationships WHERE id = $1`, id).Scan(&cnt); err != nil {
		t.Fatalf("count: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("revoke must preserve the historical row, got %d", cnt)
	}
}

func TestIntegrationDoubleRevokeConflict(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	seedSeatRelType(t, h)
	user := testkit.CreateUser(t, h)
	cap1 := testkit.CreateCapacity(t, h, tenant, user)
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")

	rel := newRelSvc(h)
	ctx := testkit.TenantCtx(tenant)
	id, err := rel.Grant(ctx, privileged.GrantSpec{
		RelType: seatRelType, SubjectKind: relationship.KindCapacity, SubjectID: cap1, Object: obj,
	})
	if err != nil {
		t.Fatalf("Grant: %v", err)
	}
	if err := rel.Revoke(ctx, id, user); err != nil {
		t.Fatalf("first Revoke: %v", err)
	}
	err = rel.Revoke(ctx, id, user)
	if kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("double revoke must be a conflict, got %v", err)
	}
}

func TestIntegrationRevokeForeignTenantIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	tenantA := testkit.CreateTenant(t, h).ID
	tenantB := testkit.CreateTenant(t, h).ID
	seedSeatRelType(t, h)
	user := testkit.CreateUser(t, h)
	capA := testkit.CreateCapacity(t, h, tenantA, user)
	objA := testkit.CreateResourceTypeAndResource(t, h, tenantA, "committee.committee")

	rel := newRelSvc(h)
	id, err := rel.Grant(testkit.TenantCtx(tenantA), privileged.GrantSpec{
		RelType: seatRelType, SubjectKind: relationship.KindCapacity, SubjectID: capA, Object: objA,
	})
	if err != nil {
		t.Fatalf("Grant: %v", err)
	}
	// Tenant B cannot revoke tenant A's edge (RLS hides it → NotFound).
	err = rel.Revoke(testkit.TenantCtx(tenantB), id, user)
	if kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("cross-tenant revoke must be NotFound, got %v", err)
	}
}

func TestIntegrationRevokeUnownedTypePrivilegeDenied(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	// Seed a "core"-owned type and an edge of it via the platform pool directly.
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		 VALUES ('core.owner_of','core','capacity','resource','owner') ON CONFLICT (key) DO NOTHING`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	user := testkit.CreateUser(t, h)
	cap1 := testkit.CreateCapacity(t, h, tenant, user)
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")
	id := uuid.New()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO relationships (id, tenant_id, rel_type, subject_kind, subject_id, object_kind, object_id, created_by)
		 VALUES ($1,$2,'core.owner_of','capacity',$3,'resource',$4,$5)`,
		id, tenant, cap1, obj.ID, uuid.Nil); err != nil {
		t.Fatalf("seed edge: %v", err)
	}

	rel := newRelSvc(h) // "committee" module
	err := rel.Revoke(testkit.TenantCtx(tenant), id, user)
	if kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("revoking an unowned edge type must be Forbidden, got %v", err)
	}
}
