package retention_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/retention"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// engineWith builds an Engine over a single record class plus a fresh DSR ledger.
func engineWith(t *testing.T, c retention.RecordClass) (*retention.Engine, *retention.DSR) {
	t.Helper()
	reg := retention.NewRegistry()
	reg.Register(c)
	if err := reg.Err(); err != nil {
		t.Fatalf("register %q: %v", c.Key, err)
	}
	dsr := retention.NewDSR(nil) // nil idgen → default UUIDv7 generator
	artifacts := retention.NewFileArtifactWriter(t.TempDir(), retention.TestKey(), nil)
	return retention.NewEngineWithCompliance(reg, dsr, nil, artifacts, nil), dsr
}

// TestIntegrationHoldsList exercises Holds.List (formerly uncovered) and the
// default-idgen branch of NewHolds(nil): List returns exactly the active holds,
// drops released ones, and is empty for a fresh tenant.
func TestIntegrationHoldsList(t *testing.T) {
	h := testkit.NewDB(t)
	svc := retention.NewHolds(nil) // nil idgen → default UUIDv7
	tenant := uuid.New()
	ctx := tctx(tenant)
	entA, entB := uuid.New(), uuid.New()

	// Fresh tenant: no holds.
	var initial []retention.Hold
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		initial, e = svc.List(ctx, db)
		return e
	}); err != nil {
		t.Fatalf("initial list: %v", err)
	}
	if len(initial) != 0 {
		t.Fatalf("fresh tenant list = %d holds, want 0", len(initial))
	}

	// Place two holds on distinct entities.
	var idA uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		if idA, e = svc.Place(ctx, db, "invoice", entA, "audit"); e != nil {
			return e
		}
		_, e = svc.Place(ctx, db, "contract", entB, "litigation")
		return e
	}); err != nil {
		t.Fatalf("place: %v", err)
	}

	var listed []retention.Hold
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		listed, e = svc.List(ctx, db)
		return e
	}); err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("list = %d holds, want 2", len(listed))
	}
	byType := map[string]retention.Hold{}
	for _, hd := range listed {
		byType[hd.EntityType] = hd
	}
	if got := byType["invoice"]; got.EntityID != entA || got.Reason != "audit" {
		t.Fatalf("invoice hold = %+v, want entity %s reason audit", got, entA)
	}
	if got := byType["contract"]; got.EntityID != entB || got.Reason != "litigation" {
		t.Fatalf("contract hold = %+v, want entity %s reason litigation", got, entB)
	}

	// Releasing one removes it from List; the other remains.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Release(ctx, db, idA)
	}); err != nil {
		t.Fatalf("release: %v", err)
	}
	var after []retention.Hold
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		after, e = svc.List(ctx, db)
		return e
	}); err != nil {
		t.Fatalf("list after release: %v", err)
	}
	if len(after) != 1 || after[0].EntityType != "contract" {
		t.Fatalf("list after release = %+v, want only the contract hold", after)
	}
}

// TestIntegrationHoldPlacedByNilWithoutActor covers the actorOrNil nil branch:
// a tenant context with no actor id records placed_by as SQL NULL.
func TestIntegrationHoldPlacedByNilWithoutActor(t *testing.T) {
	h := testkit.NewDB(t)
	svc := retention.NewHolds(nil)
	tenant := uuid.New()
	// Tenant only — deliberately no WithActorID.
	ctx := database.WithTenantID(context.Background(), tenant)
	entity := uuid.New()

	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = svc.Place(ctx, db, "invoice", entity, "no-actor")
		return e
	}); err != nil {
		t.Fatalf("place: %v", err)
	}

	var placedBy *uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT placed_by FROM legal_holds WHERE id = $1`, id).Scan(&placedBy); err != nil {
		t.Fatalf("read placed_by: %v", err)
	}
	if placedBy != nil {
		t.Fatalf("placed_by = %v, want NULL when no actor in context", *placedBy)
	}
}

// TestIntegrationPlaceAndOpenValidation covers the input-validation branches of
// Holds.Place and DSR.Open (and the NewDSR(nil) default-idgen branch).
func TestIntegrationPlaceAndOpenValidation(t *testing.T) {
	h := testkit.NewDB(t)
	holds := retention.NewHolds(nil)
	dsr := retention.NewDSR(nil)
	ctx := tctx(uuid.New())

	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if _, e := holds.Place(ctx, db, "", uuid.New(), "r"); kerr.KindOf(e) != kerr.KindValidation {
			t.Errorf("empty entity type: got %v, want validation", e)
		}
		if _, e := holds.Place(ctx, db, "invoice", uuid.Nil, "r"); kerr.KindOf(e) != kerr.KindValidation {
			t.Errorf("nil entity id: got %v, want validation", e)
		}
		if _, e := holds.Place(ctx, db, "invoice", uuid.New(), ""); kerr.KindOf(e) != kerr.KindValidation {
			t.Errorf("empty reason: got %v, want validation", e)
		}
		if _, e := dsr.Open(ctx, db, "", retention.KindExport); kerr.KindOf(e) != kerr.KindValidation {
			t.Errorf("empty subject ref: got %v, want validation", e)
		}
		if _, e := dsr.Open(ctx, db, "party:1", retention.Kind("bogus")); kerr.KindOf(e) != kerr.KindValidation {
			t.Errorf("bad kind: got %v, want validation", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("tx: %v", err)
	}
}

// TestIntegrationDSRRejectRequiresReasonAndGetNotFound covers Reject's empty-reason
// validation and DSR.Get's not-found branch.
func TestIntegrationDSRRejectRequiresReasonAndGetNotFound(t *testing.T) {
	h := testkit.NewDB(t)
	dsr := retention.NewDSR(nil)
	ctx := tctx(uuid.New())

	// Get of a random id → not found.
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := dsr.Get(ctx, db, uuid.New())
		if kerr.KindOf(e) != kerr.KindNotFound {
			t.Errorf("get unknown: got %v, want not-found", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("tx: %v", err)
	}

	// Reject with an empty reason → validation, before any DB row is touched.
	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = dsr.Open(ctx, db, "party:7", retention.KindErasure)
		if e != nil {
			return e
		}
		if e := dsr.Reject(ctx, db, id, ""); kerr.KindOf(e) != kerr.KindValidation {
			t.Errorf("reject without reason: got %v, want validation", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("tx: %v", err)
	}
}

// TestRegistryValidation covers Register's empty-key rejection and the
// short-circuit that retains the first error on subsequent Register calls.
func TestRegistryValidation(t *testing.T) {
	reg := retention.NewRegistry()
	reg.Register(retention.RecordClass{Key: ""})
	if kerr.KindOf(reg.Err()) != kerr.KindInternal {
		t.Fatalf("empty key: got %v, want internal", reg.Err())
	}
	first := reg.Err()
	// After an error is retained, further Register calls short-circuit and the
	// original error is preserved.
	reg.Register(retention.RecordClass{Key: "ok"})
	if reg.Err() != first {
		t.Fatalf("Err changed after short-circuit: %v", reg.Err())
	}
}

// TestIntegrationSweepDispositionCallbackError covers SweepDisposition's error
// path: a class whose Dispose fails aborts the sweep with a wrapped error.
func TestIntegrationSweepDispositionCallbackError(t *testing.T) {
	h := testkit.NewDB(t)
	sentinel := errors.New("dispose exploded")
	eng, _ := engineWith(t, retention.RecordClass{
		Key: "boom",
		Dispose: func(ctx context.Context, db database.TenantDB, before time.Time) (int, error) {
			return 0, sentinel
		},
	})
	ctx := tctx(uuid.New())

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := eng.SweepDisposition(ctx, db, time.Now())
		return e
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("sweep error = %v, want wrapped sentinel", err)
	}
}

// TestIntegrationRunExportBranches covers RunExport's guard branches: wrong kind,
// not-pending, unknown request (Get error), and a failing Export callback.
func TestIntegrationRunExportBranches(t *testing.T) {
	h := testkit.NewDB(t)
	eng, dsr := newEngine(t) // peopleClass with real Export/Erase/Dispose
	ctx := tctx(uuid.New())

	// wrong_kind: an erasure DSR cannot be run as an export.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		id, e := dsr.Open(ctx, db, "party:1", retention.KindErasure)
		if e != nil {
			return e
		}
		if _, e := eng.RunExportDetailed(ctx, db, id); kerr.KindOf(e) != kerr.KindConflict {
			t.Errorf("run export on erasure DSR: got %v, want conflict", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("wrong_kind tx: %v", err)
	}

	// not_pending: a completed export cannot be re-run.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		id, e := dsr.Open(ctx, db, "party:2", retention.KindExport)
		if e != nil {
			return e
		}
		if e := dsr.Complete(ctx, db, id); e != nil {
			return e
		}
		if _, e := eng.RunExportDetailed(ctx, db, id); kerr.KindOf(e) != kerr.KindConflict {
			t.Errorf("run export on completed DSR: got %v, want conflict", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("not_pending tx: %v", err)
	}

	// Get error: unknown request id → not-found propagated.
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		if _, e := eng.RunExportDetailed(ctx, db, uuid.New()); kerr.KindOf(e) != kerr.KindNotFound {
			t.Errorf("run export on unknown id: got %v, want not-found", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("get-error tx: %v", err)
	}

	// Export callback error: aborts the run with a wrapped error.
	sentinel := errors.New("export exploded")
	badEng, badDSR := engineWith(t, retention.RecordClass{
		Key: "x",
		Export: func(ctx context.Context, db database.TenantDB, subject string) (map[string]any, error) {
			return nil, sentinel
		},
	})
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		id, e := badDSR.Open(ctx, db, "party:3", retention.KindExport)
		if e != nil {
			return e
		}
		_, e = badEng.RunExportDetailed(ctx, db, id)
		if !errors.Is(e, sentinel) {
			t.Errorf("run export callback error = %v, want wrapped sentinel", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("callback-error tx: %v", err)
	}
}

// TestIntegrationRunErasureBranches mirrors the export guard coverage for
// RunErasure: wrong kind, not-pending, unknown request, and a failing Erase.
func TestIntegrationRunErasureBranches(t *testing.T) {
	h := testkit.NewDB(t)
	eng, dsr := newEngine(t)
	ctx := tctx(uuid.New())

	// wrong_kind: an export DSR cannot be run as an erasure.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		id, e := dsr.Open(ctx, db, "party:1", retention.KindExport)
		if e != nil {
			return e
		}
		if _, e := eng.RunErasureDetailed(ctx, db, id); kerr.KindOf(e) != kerr.KindConflict {
			t.Errorf("run erasure on export DSR: got %v, want conflict", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("wrong_kind tx: %v", err)
	}

	// not_pending: a completed erasure cannot be re-run.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		id, e := dsr.Open(ctx, db, "party:2", retention.KindErasure)
		if e != nil {
			return e
		}
		if e := dsr.Complete(ctx, db, id); e != nil {
			return e
		}
		if _, e := eng.RunErasureDetailed(ctx, db, id); kerr.KindOf(e) != kerr.KindConflict {
			t.Errorf("run erasure on completed DSR: got %v, want conflict", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("not_pending tx: %v", err)
	}

	// Get error: unknown request id → not-found propagated.
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		if _, e := eng.RunErasureDetailed(ctx, db, uuid.New()); kerr.KindOf(e) != kerr.KindNotFound {
			t.Errorf("run erasure on unknown id: got %v, want not-found", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("get-error tx: %v", err)
	}

	// Erase callback error: aborts the run with a wrapped error.
	sentinel := errors.New("erase exploded")
	badEng, badDSR := engineWith(t, retention.RecordClass{
		Key: "x",
		Erase: func(ctx context.Context, db database.TenantDB, subject string) (int, error) {
			return 0, sentinel
		},
	})
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		id, e := badDSR.Open(ctx, db, "party:3", retention.KindErasure)
		if e != nil {
			return e
		}
		_, e = badEng.RunErasureDetailed(ctx, db, id)
		if !errors.Is(e, sentinel) {
			t.Errorf("run erasure callback error = %v, want wrapped sentinel", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("callback-error tx: %v", err)
	}
}

// TestIntegrationReadOnlyTxSurfacesWriteErrors drives the generic db-error wrap
// paths of Place, Release, Open, and the DSR status transition: attempting each
// write inside a read-only transaction returns a non-conflict internal error
// (proving, for Place, that a read-only rejection is not misread as a
// unique-violation conflict — the isUniqueViolation false branch).
func TestIntegrationReadOnlyTxSurfacesWriteErrors(t *testing.T) {
	h := testkit.NewDB(t)
	holds := retention.NewHolds(nil)
	dsr := retention.NewDSR(nil)
	ctx := tctx(uuid.New())

	// Seed one active hold and one pending DSR in a writable tx.
	var holdID, dsrID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		if holdID, e = holds.Place(ctx, db, "invoice", uuid.New(), "keep"); e != nil {
			return e
		}
		dsrID, e = dsr.Open(ctx, db, "party:ro", retention.KindExport)
		return e
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	writes := map[string]func(ctx context.Context, db database.TenantDB) error{
		"place": func(ctx context.Context, db database.TenantDB) error {
			_, e := holds.Place(ctx, db, "doc", uuid.New(), "x")
			return e
		},
		"release": func(ctx context.Context, db database.TenantDB) error { return holds.Release(ctx, db, holdID) },
		"open": func(ctx context.Context, db database.TenantDB) error {
			_, e := dsr.Open(ctx, db, "s", retention.KindErasure)
			return e
		},
		"complete": func(ctx context.Context, db database.TenantDB) error { return dsr.Complete(ctx, db, dsrID) },
	}
	for name, op := range writes {
		err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
			return op(ctx, db)
		})
		if err == nil {
			t.Errorf("%s: expected write in read-only tx to fail", name)
			continue
		}
		if kerr.KindOf(err) == kerr.KindConflict {
			t.Errorf("%s: read-only failure misclassified as conflict: %v", name, err)
		}
	}
}
