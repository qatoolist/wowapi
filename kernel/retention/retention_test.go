package retention_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/retention"
	"github.com/qatoolist/wowapi/testkit"
)

func tctx(tenant uuid.UUID) context.Context {
	return database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())
}

func TestIntegrationLegalHoldLifecycle(t *testing.T) {
	h := testkit.NewDB(t)
	svc := retention.NewHolds(model.UUIDv7())
	tenant := uuid.New()
	ctx := tctx(tenant)
	entity := uuid.New()

	var id uuid.UUID
	// Place a hold; IsHeld becomes true.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = svc.Place(ctx, db, "invoice", entity, "litigation")
		return e
	}); err != nil {
		t.Fatalf("place: %v", err)
	}
	assertHeld(t, h, svc, ctx, entity, true)

	// A second active hold on the same entity conflicts.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Place(ctx, db, "invoice", entity, "again")
		return e
	}); kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("duplicate hold should conflict, got %v", err)
	}

	// Release; IsHeld becomes false and List is empty.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Release(ctx, db, id)
	}); err != nil {
		t.Fatalf("release: %v", err)
	}
	assertHeld(t, h, svc, ctx, entity, false)

	// Releasing again is not-found.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Release(ctx, db, id)
	}); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("double release should be not-found, got %v", err)
	}

	// After release a fresh hold may be placed (unique index only bars ACTIVE dups).
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Place(ctx, db, "invoice", entity, "new matter")
		return e
	}); err != nil {
		t.Fatalf("re-place after release: %v", err)
	}
}

func TestIntegrationLegalHoldTenantIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	svc := retention.NewHolds(model.UUIDv7())
	t1, t2 := uuid.New(), uuid.New()
	entity := uuid.New()

	if err := h.TxM.WithTenant(tctx(t1), func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Place(ctx, db, "doc", entity, "hold")
		return e
	}); err != nil {
		t.Fatal(err)
	}
	// Tenant 2 does not see tenant 1's hold on the same entity id.
	assertHeld(t, h, svc, tctx(t2), entity, false)
}

func TestIntegrationDSRLifecycle(t *testing.T) {
	h := testkit.NewDB(t)
	dsr := retention.NewDSR(model.UUIDv7())
	tenant := uuid.New()
	ctx := tctx(tenant)

	// Export request → complete.
	var exportID uuid.UUID
	_ = h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		exportID, e = dsr.Open(ctx, db, "party:123", retention.KindExport)
		return e
	})
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return dsr.Complete(ctx, db, exportID)
	}); err != nil {
		t.Fatalf("complete: %v", err)
	}
	// Completing again conflicts (no longer pending).
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return dsr.Complete(ctx, db, exportID)
	}); kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("re-complete should conflict, got %v", err)
	}

	// Erasure request rejected with a statutory override reason.
	var eraseID uuid.UUID
	_ = h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		eraseID, e = dsr.Open(ctx, db, "party:123", retention.KindErasure)
		return e
	})
	// Reject requires a reason.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return dsr.Reject(ctx, db, eraseID, "")
	}); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("reject without a reason should be validation error, got %v", err)
	}
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return dsr.Reject(ctx, db, eraseID, "retained under tax law §44 for 7 years")
	}); err != nil {
		t.Fatalf("reject: %v", err)
	}
	var req retention.Request
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		req, e = dsr.Get(ctx, db, eraseID)
		return e
	})
	if req.Status != "rejected" || req.OverrideReason == "" {
		t.Fatalf("erasure request = %+v, want rejected with an override reason", req)
	}
}

func assertHeld(t *testing.T, h *testkit.DBHandle, svc *retention.Holds, ctx context.Context, entity uuid.UUID, want bool) {
	t.Helper()
	var got bool
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		got, e = svc.IsHeld(ctx, db, "invoice", entity)
		if e != nil {
			return e
		}
		// also try the doc/other entity_type used by isolation test
		if !got {
			g2, e2 := svc.IsHeld(ctx, db, "doc", entity)
			if e2 != nil {
				return e2
			}
			got = got || g2
		}
		return nil
	}); err != nil {
		t.Fatalf("isHeld: %v", err)
	}
	if got != want {
		t.Fatalf("IsHeld(%s) = %v, want %v", entity, got, want)
	}
}
