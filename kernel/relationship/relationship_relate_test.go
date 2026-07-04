package relationship_test

import (
	"context"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/relationship"
	"github.com/qatoolist/wowapi/testkit"
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
	if err := h.PlatformTxM.WithTenant(database.WithTenantID(ctx, tenant),
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

	err := h.TxM.WithTenant(database.WithTenantID(ctx, tenant),
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

	if err := h.PlatformTxM.WithTenant(database.WithTenantID(ctx, tenantA),
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
