package bulk_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/foundation/bulk"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// fixedGen is a deterministic IDGen that always mints the same id, so a test can
// provoke primary-key collisions on the operation and item inserts.
type fixedGen struct{ id uuid.UUID }

func (f fixedGen) New() uuid.UUID { return f.id }

// hookTxM wraps a real TxManager and invokes after(n) once each WithTenant /
// WithTenantRO call returns (n is the running call count). It lets a test change
// the world — cancel a context, revoke a grant — between the internal
// transactions of a single Process run, exercising the mid-run branches.
type hookTxM struct {
	inner database.TxManager
	after func(n int)
	n     int
}

func (h *hookTxM) WithTenant(ctx context.Context, fn func(ctx context.Context, db database.TenantDB) error) error {
	err := h.inner.WithTenant(ctx, fn)
	h.n++
	if h.after != nil {
		h.after(h.n)
	}
	return err
}

func (h *hookTxM) WithTenantRO(ctx context.Context, fn func(ctx context.Context, db database.TenantDB) error) error {
	err := h.inner.WithTenantRO(ctx, fn)
	h.n++
	if h.after != nil {
		h.after(h.n)
	}
	return err
}

func (h *hookTxM) Platform(ctx context.Context, fn func(ctx context.Context, db database.DB) error) error {
	return h.inner.Platform(ctx, fn)
}

// adminExec runs owner DDL/DML (grants, revokes) for error-injection tests.
func adminExec(t *testing.T, h *testkit.DBHandle, sql string) {
	t.Helper()
	if _, err := h.Admin.Exec(context.Background(), sql); err != nil {
		t.Fatalf("admin exec %q: %v", sql, err)
	}
}

// startKind runs svc.Start with an explicit kind inside a tenant tx and returns
// the op id and Start's error (unlike the start helper, which fails the test).
func startKind(t *testing.T, h *testkit.DBHandle, svc *bulk.Service, ctx context.Context, kind string, ps []json.RawMessage) (uuid.UUID, error) {
	t.Helper()
	var (
		id   uuid.UUID
		serr error
	)
	txErr := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		id, serr = svc.Start(ctx, db, kind, ps)
		return serr
	})
	if serr == nil && txErr != nil {
		t.Fatalf("WithTenant: %v", txErr)
	}
	return id, serr
}

// TestIntegrationBulkNewDefaultsIDGen proves New(nil) installs the production
// UUIDv7 generator: an operation started with it processes end-to-end.
func TestIntegrationBulkNewDefaultsIDGen(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(nil) // nil -> default UUIDv7 generator
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}, item{Val: 2}))
	if id == uuid.Nil {
		t.Fatal("Start returned nil id with default generator")
	}
	n, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc)
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	if n != 2 {
		t.Fatalf("processed %d, want 2", n)
	}
	if p := progress(t, h, svc, ctx, id); p.Done != 2 || p.Status != "completed" {
		t.Fatalf("progress = %+v, want Done2/completed", p)
	}
}

// TestIntegrationBulkStartRejectsEmptyKind covers the kind validation guard.
func TestIntegrationBulkStartRejectsEmptyKind(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	ctx := database.WithTenantID(context.Background(), uuid.New())

	_, err := startKind(t, h, svc, ctx, "", payloads(item{Val: 1}))
	if err == nil {
		t.Fatal("Start with empty kind: want error, got nil")
	}
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("Start empty kind kind = %v, want KindValidation (err=%v)", kerr.KindOf(err), err)
	}
}

// TestIntegrationBulkStartDefaultsEmptyPayload proves a zero-length payload is
// stored as an empty JSON object rather than NULL/invalid.
func TestIntegrationBulkStartDefaultsEmptyPayload(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, []json.RawMessage{{}}) // one empty (len 0) payload

	var payload string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT payload::text FROM bulk_items WHERE bulk_id = $1`, id).Scan(&payload); err != nil {
		t.Fatalf("read item payload: %v", err)
	}
	if payload != "{}" {
		t.Fatalf("stored payload = %q, want {}", payload)
	}
}

// TestIntegrationBulkStartRecordsActor proves the actor from context is written
// to created_by (the actorOrNil id branch).
func TestIntegrationBulkStartRecordsActor(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	actor := uuid.New()
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tenant), actor)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}))

	var createdBy uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT created_by FROM bulk_operations WHERE id = $1`, id).Scan(&createdBy); err != nil {
		t.Fatalf("read created_by: %v", err)
	}
	if createdBy != actor {
		t.Fatalf("created_by = %v, want %v", createdBy, actor)
	}
}

// TestIntegrationBulkStartDuplicateOperationErrors covers the operation-insert
// error path: a second Start reusing the same id collides on the primary key.
func TestIntegrationBulkStartDuplicateOperationErrors(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(fixedGen{id: uuid.New()})
	ctx := database.WithTenantID(context.Background(), uuid.New())

	if _, err := startKind(t, h, svc, ctx, "test.bulk", payloads(item{Val: 1})); err != nil {
		t.Fatalf("first Start: %v", err)
	}
	// Same fixed id -> duplicate bulk_operations primary key.
	_, err := startKind(t, h, svc, ctx, "test.bulk", payloads(item{Val: 2}))
	if err == nil {
		t.Fatal("second Start with duplicate op id: want error, got nil")
	}
}

// TestIntegrationBulkStartDuplicateItemErrors covers the item-insert error path:
// with a fixed id the operation inserts, then the second item collides on the
// item primary key.
func TestIntegrationBulkStartDuplicateItemErrors(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(fixedGen{id: uuid.New()})
	ctx := database.WithTenantID(context.Background(), uuid.New())

	_, err := startKind(t, h, svc, ctx, "test.bulk", payloads(item{Val: 1}, item{Val: 2}))
	if err == nil {
		t.Fatal("Start with colliding item ids: want error, got nil")
	}
}

// TestIntegrationBulkProgressNotFound covers the not-found branch of Progress.
func TestIntegrationBulkProgressNotFound(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	ctx := database.WithTenantID(context.Background(), uuid.New())

	var perr error
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, perr = svc.Progress(ctx, db, uuid.New()) // no such operation
		return nil
	})
	if perr == nil {
		t.Fatal("Progress of missing op: want error, got nil")
	}
	if kerr.KindOf(perr) != kerr.KindNotFound {
		t.Fatalf("Progress missing op kind = %v, want KindNotFound (err=%v)", kerr.KindOf(perr), perr)
	}
}

// TestIntegrationBulkProgressOperationReadError covers the non-NoRows read error
// on the operation row: revoking SELECT makes the first query fail with a
// permission error (which is not ErrNoRows).
func TestIntegrationBulkProgressOperationReadError(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}))
	adminExec(t, h, `REVOKE SELECT ON bulk_operations FROM app_rt`)

	var perr error
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, perr = svc.Progress(ctx, db, id)
		return nil
	})
	if perr == nil {
		t.Fatal("Progress with SELECT revoked on bulk_operations: want error, got nil")
	}
	if kerr.KindOf(perr) == kerr.KindNotFound {
		t.Fatalf("Progress error must not be NotFound for a permission failure: %v", perr)
	}
}

// TestIntegrationBulkProgressItemCountError covers the item-count read error: the
// operation row is readable but the item aggregate query is denied.
func TestIntegrationBulkProgressItemCountError(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}))
	adminExec(t, h, `REVOKE SELECT ON bulk_items FROM app_rt`)

	var perr error
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, perr = svc.Progress(ctx, db, id)
		return nil
	})
	if perr == nil {
		t.Fatal("Progress with SELECT revoked on bulk_items: want error, got nil")
	}
}

// TestIntegrationBulkProcessMarkRunningError covers the failure to mark the
// operation running (and, transitively, mark's Exec error path).
func TestIntegrationBulkProcessMarkRunningError(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}))
	adminExec(t, h, `REVOKE UPDATE ON bulk_operations FROM app_rt`)

	if _, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc); err == nil {
		t.Fatal("Process with UPDATE revoked on bulk_operations: want error, got nil")
	}
}

// TestIntegrationBulkProcessNextError covers the failure to read the next pending
// item (and next's non-NoRows scan-error path).
func TestIntegrationBulkProcessNextError(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}))
	// Keep UPDATE (so mark running succeeds) but deny reading items.
	adminExec(t, h, `REVOKE SELECT ON bulk_items FROM app_rt`)

	if _, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc); err == nil {
		t.Fatal("Process with SELECT revoked on bulk_items: want error, got nil")
	}
}

// TestIntegrationBulkProcessItemWriteError covers runItem's two write-error
// branches: the done-mark Exec fails, then the failure-ledger Exec fails too, so
// runItem returns an infrastructure error that stops Process.
func TestIntegrationBulkProcessItemWriteError(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}))
	// Keep SELECT (so next finds the item) but deny all writes to items, so both
	// the done UPDATE and the failure-ledger UPDATE inside runItem fail.
	adminExec(t, h, `REVOKE UPDATE ON bulk_items FROM app_rt`)

	if _, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc); err == nil {
		t.Fatal("Process with UPDATE revoked on bulk_items: want error, got nil")
	}
	// The item stays pending: neither the done nor failed write could land.
	var pending int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM bulk_items WHERE bulk_id=$1 AND status='pending'`, id).Scan(&pending); err != nil {
		t.Fatal(err)
	}
	if pending != 1 {
		t.Fatalf("pending items = %d, want 1 (writes must not have landed)", pending)
	}
}

// TestIntegrationBulkProcessContextCanceled covers the in-loop cancellation
// check: the context is cancelled right after the running-mark commits, so the
// loop returns before touching any item.
func TestIntegrationBulkProcessContextCanceled(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}, item{Val: 2}))

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Cancel after the first internal tx (mark running) commits, before the loop
	// reads an item.
	txm := &hookTxM{inner: h.TxM, after: func(n int) {
		if n == 1 {
			cancel()
		}
	}}

	n, err := svc.Process(runCtx, txm, tenant, id, 0, okFunc)
	if err == nil {
		t.Fatal("Process on a cancelled context: want error, got nil")
	}
	if n != 0 {
		t.Fatalf("processed %d after cancellation, want 0", n)
	}
	// The items must remain untouched.
	if p := progress(t, h, svc, ctx, id); p.Done != 0 || p.Pending != 2 {
		t.Fatalf("after cancellation: %+v, want Done0/Pending2", p)
	}
}

// TestIntegrationBulkProcessMarkCompletedError covers the failure to mark the
// operation completed once no pending items remain: UPDATE is revoked after the
// running-mark commits, so the completed-mark fails.
func TestIntegrationBulkProcessMarkCompletedError(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}))

	txm := &hookTxM{inner: h.TxM, after: func(n int) {
		if n == 1 { // after the running-mark commits
			adminExec(t, h, `REVOKE UPDATE ON bulk_operations FROM app_rt`)
		}
	}}

	if _, err := svc.Process(context.Background(), txm, tenant, id, 0, okFunc); err == nil {
		t.Fatal("Process completed-mark with UPDATE revoked: want error, got nil")
	}
}
