package relationship_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/relationship"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// relationship_relate_test.go — QA G4 / finding D1 (data integrity + security):
// relationship.Relate is an EXPORTED write path with no callers and no test. It
// inserts into `relationships`, on which app_rt has SELECT only and app_platform
// has INSERT (SEC-24 — edge creation is a kernel/platform capability). These
// tests pin the CORRECT usage (tenant-bound app_platform) end-to-end with Has,
// and the privilege boundary (app_rt cannot create edges), so the split cannot
// silently regress and callers have a worked example.

func seedRelType(t *testing.T, h *testkit.DBHandle, ctx context.Context, relType string) {
	t.Helper()
	mustExec(t, h, ctx, `INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		VALUES ($1,$2,'capacity','resource','owner') ON CONFLICT (key) DO NOTHING`, relType, "core")
}

func TestIntegrationRelateAsPlatformThenHas(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := testkit.CreateTenant(t, h).ID
	relType := "core.owner_of"
	seedResourceType(t, h, ctx, "requests.request")
	seedRelType(t, h, ctx, relType)
	cap1 := seedCapacity(t, h, ctx, tenant)
	obj := seedResource(t, h, ctx, tenant, "requests.request")
	gen := model.UUIDv7()

	// Create the edge via Relate on a tenant-bound app_platform tx (the only role
	// permitted to INSERT relationships).
	actorID := uuid.New()
	if err := h.PlatformTxM.WithTenant(database.WithActorID(database.WithTenantID(ctx, tenant), actorID),
		func(ctx context.Context, db database.TenantDB) error {
			return relationship.Relate(ctx, db, gen, relType, "capacity", cap1, "resource", obj.ID)
		}); err != nil {
		t.Fatalf("Relate (app_platform): %v", err)
	}

	// The ReBAC checker (app_rt, read-only) now sees the edge.
	checker := relationship.NewChecker()
	actor := authz.Actor{Kind: authz.ActorUser, CapacityID: cap1, TenantID: tenant}
	var has bool
	if err := h.TxM.WithTenantRO(database.WithTenantID(ctx, tenant),
		func(ctx context.Context, db database.TenantDB) error {
			var e error
			has, e = checker.Has(ctx, db, actor, relType, obj, time.Now().Add(time.Minute))
			return e
		}); err != nil {
		t.Fatalf("Has: %v", err)
	}
	if !has {
		t.Fatal("edge created by Relate is not visible to Has")
	}
}

// The module role (app_rt) MUST NOT be able to create relationship edges — a
// module cannot self-grant ReBAC access by writing its own edges (SEC-24).
func TestIntegrationRelateAsAppRtDenied(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := testkit.CreateTenant(t, h).ID
	relType := "core.owner_of"
	seedResourceType(t, h, ctx, "requests.request")
	seedRelType(t, h, ctx, relType)
	cap1 := seedCapacity(t, h, ctx, tenant)
	obj := seedResource(t, h, ctx, tenant, "requests.request")
	gen := model.UUIDv7()

	actorID := uuid.New()
	err := h.TxM.WithTenant(database.WithActorID(database.WithTenantID(ctx, tenant), actorID),
		func(ctx context.Context, db database.TenantDB) error {
			return relationship.Relate(ctx, db, gen, relType, "capacity", cap1, "resource", obj.ID)
		})
	if err == nil {
		t.Fatal("app_rt must NOT be able to create relationship edges (SEC-24)")
	}
	if kerr.KindOf(err) == kerr.KindNotFound {
		t.Fatalf("expected a permission error, got NotFound: %v", err)
	}
}

// Tenant isolation: an edge created in tenant A is invisible to tenant B's checker.
func TestIntegrationRelateTenantIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenantA := testkit.CreateTenant(t, h).ID
	tenantB := testkit.CreateTenant(t, h).ID
	relType := "core.owner_of"
	seedResourceType(t, h, ctx, "requests.request")
	seedRelType(t, h, ctx, relType)
	capA := seedCapacity(t, h, ctx, tenantA)
	objA := seedResource(t, h, ctx, tenantA, "requests.request")
	gen := model.UUIDv7()

	actorID := uuid.New()
	if err := h.PlatformTxM.WithTenant(database.WithActorID(database.WithTenantID(ctx, tenantA), actorID),
		func(ctx context.Context, db database.TenantDB) error {
			return relationship.Relate(ctx, db, gen, relType, "capacity", capA, "resource", objA.ID)
		}); err != nil {
		t.Fatalf("Relate: %v", err)
	}

	checker := relationship.NewChecker()
	// Tenant B, using the same actor + object refs, sees no edge (RLS).
	actor := authz.Actor{Kind: authz.ActorUser, CapacityID: capA, TenantID: tenantB}
	var has bool
	if err := h.TxM.WithTenantRO(database.WithTenantID(ctx, tenantB),
		func(ctx context.Context, db database.TenantDB) error {
			var e error
			has, e = checker.Has(ctx, db, actor, relType, objA, time.Now().Add(time.Minute))
			return e
		}); err != nil {
		t.Fatalf("Has (tenant B): %v", err)
	}
	if has {
		t.Fatal("tenant B must not see tenant A's relationship edge")
	}
}

// TestIntegrationRelateRequiresActor proves DATA-07 T4: Relate fails closed
// when no actor is bound in ctx.
func TestIntegrationRelateRequiresActor(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := testkit.CreateTenant(t, h).ID
	relType := "core.owner_of"
	seedResourceType(t, h, ctx, "requests.request")
	seedRelType(t, h, ctx, relType)
	cap1 := seedCapacity(t, h, ctx, tenant)
	obj := seedResource(t, h, ctx, tenant, "requests.request")
	gen := model.UUIDv7()

	err := h.PlatformTxM.WithTenant(database.WithTenantID(ctx, tenant),
		func(ctx context.Context, db database.TenantDB) error {
			return relationship.Relate(ctx, db, gen, relType, "capacity", cap1, "resource", obj.ID)
		})
	if err == nil {
		t.Fatal("Relate without actor must fail")
	}
	if kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("want KindForbidden, got %v", err)
	}
}

// TestIntegrationRelateAttributesAndVersions proves DATA-07 T4: created_by
// reflects the bound actor and re-relating the same active edge bumps version.
func TestIntegrationRelateAttributesAndVersions(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := testkit.CreateTenant(t, h).ID
	relType := "core.owner_of"
	seedResourceType(t, h, ctx, "requests.request")
	seedRelType(t, h, ctx, relType)
	cap1 := seedCapacity(t, h, ctx, tenant)
	obj := seedResource(t, h, ctx, tenant, "requests.request")
	gen := model.UUIDv7()
	actorID := uuid.New()

	relate := func() error {
		return h.PlatformTxM.WithTenant(database.WithActorID(database.WithTenantID(ctx, tenant), actorID),
			func(ctx context.Context, db database.TenantDB) error {
				return relationship.Relate(ctx, db, gen, relType, "capacity", cap1, "resource", obj.ID)
			})
	}
	if err := relate(); err != nil {
		t.Fatalf("first Relate: %v", err)
	}
	if err := relate(); err != nil {
		t.Fatalf("second Relate: %v", err)
	}

	var createdBy, updatedBy uuid.UUID
	var version int
	if err := h.Admin.QueryRow(ctx,
		`SELECT created_by, updated_by, version FROM relationships WHERE tenant_id = $1 AND subject_id = $2 AND object_id = $3`,
		tenant, cap1, obj.ID).Scan(&createdBy, &updatedBy, &version); err != nil {
		t.Fatalf("query edge: %v", err)
	}
	if createdBy != actorID {
		t.Fatalf("created_by = %v; want %v", createdBy, actorID)
	}
	if updatedBy != actorID {
		t.Fatalf("updated_by = %v; want %v", updatedBy, actorID)
	}
	if version != 2 {
		t.Fatalf("version = %d; want 2", version)
	}
}

// TestIntegrationRelateWritesAudit proves DATA-07 T4: Relate writes a durable
// audit row inside the same transaction.
func TestIntegrationRelateWritesAudit(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := testkit.CreateTenant(t, h).ID
	relType := "core.owner_of"
	seedResourceType(t, h, ctx, "requests.request")
	seedRelType(t, h, ctx, relType)
	cap1 := seedCapacity(t, h, ctx, tenant)
	obj := seedResource(t, h, ctx, tenant, "requests.request")
	gen := model.UUIDv7()
	actorID := uuid.New()

	if err := h.PlatformTxM.WithTenant(database.WithActorID(database.WithTenantID(ctx, tenant), actorID),
		func(ctx context.Context, db database.TenantDB) error {
			return relationship.Relate(ctx, db, gen, relType, "capacity", cap1, "resource", obj.ID)
		}); err != nil {
		t.Fatalf("Relate: %v", err)
	}

	var n int
	if err := h.Admin.QueryRow(ctx,
		`SELECT count(*) FROM audit_logs WHERE tenant_id = $1 AND action = 'relationship.relate'`,
		tenant).Scan(&n); err != nil {
		t.Fatalf("count audit rows: %v", err)
	}
	if n != 1 {
		t.Fatalf("want 1 audit row, got %d", n)
	}
}
