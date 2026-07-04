package bulk_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/bulk"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/testkit"
)

type item struct {
	Val  int  `json:"val"`
	Fail bool `json:"fail"`
}

func payloads(items ...item) []json.RawMessage {
	out := make([]json.RawMessage, len(items))
	for i, it := range items {
		b, _ := json.Marshal(it)
		out[i] = b
	}
	return out
}

func start(t *testing.T, h *testkit.DBHandle, svc *bulk.Service, ctx context.Context, ps []json.RawMessage) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = svc.Start(ctx, db, "test.bulk", ps)
		return e
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return id
}

func progress(t *testing.T, h *testkit.DBHandle, svc *bulk.Service, ctx context.Context, id uuid.UUID) bulk.Progress {
	t.Helper()
	var p bulk.Progress
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		p, e = svc.Progress(ctx, db, id)
		return e
	}); err != nil {
		t.Fatalf("Progress: %v", err)
	}
	return p
}

func TestIntegrationBulkAllSucceed(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}, item{Val: 2}, item{Val: 3}))
	n, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc)
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	if n != 3 {
		t.Fatalf("processed %d, want 3", n)
	}
	p := progress(t, h, svc, ctx, id)
	if p.Total != 3 || p.Done != 3 || p.Failed != 0 || p.Pending != 0 || p.Status != "completed" {
		t.Fatalf("progress = %+v, want Total3/Done3/Failed0/Pending0/completed", p)
	}
}

func TestIntegrationBulkPartialFailureLedger(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	// items 2 and 4 fail; the rest succeed.
	id := start(t, h, svc, ctx, payloads(
		item{Val: 1}, item{Val: 2, Fail: true}, item{Val: 3}, item{Val: 4, Fail: true}, item{Val: 5}))

	if _, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, markFunc(h)); err != nil {
		t.Fatalf("Process: %v", err)
	}
	p := progress(t, h, svc, ctx, id)
	if p.Done != 3 || p.Failed != 2 || p.Pending != 0 || p.Status != "completed" {
		t.Fatalf("progress = %+v, want Done3/Failed2/Pending0/completed (partial failure)", p)
	}
	// Failed items carry the error; one item's failure did not stop the others.
	var failedWithErr int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM bulk_items WHERE bulk_id=$1 AND status='failed' AND last_error <> ''`, id).
		Scan(&failedWithErr); err != nil {
		t.Fatal(err)
	}
	if failedWithErr != 2 {
		t.Fatalf("failed items with a recorded error = %d, want 2", failedWithErr)
	}
	// Atomicity: only the 3 SUCCEEDED items wrote their mark — the failed items'
	// writes rolled back with their transaction.
	var marks int
	if err := h.Admin.QueryRow(context.Background(), `SELECT count(*) FROM bulk_marks`).Scan(&marks); err != nil {
		t.Fatal(err)
	}
	if marks != 3 {
		t.Fatalf("bulk_marks rows = %d, want 3 (failed items' writes must roll back)", marks)
	}
}

func TestIntegrationBulkChunkedResumable(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}, item{Val: 2}, item{Val: 3}, item{Val: 4}, item{Val: 5}))

	// Chunk of 2, then 2, then the rest — simulating resumable batches.
	if n, _ := svc.Process(context.Background(), h.TxM, tenant, id, 2, okFunc); n != 2 {
		t.Fatalf("first chunk processed %d, want 2", n)
	}
	if p := progress(t, h, svc, ctx, id); p.Done != 2 || p.Pending != 3 {
		t.Fatalf("after chunk 1: %+v, want Done2/Pending3", p)
	}
	if n, _ := svc.Process(context.Background(), h.TxM, tenant, id, 2, okFunc); n != 2 {
		t.Fatalf("second chunk processed %d, want 2", n)
	}
	if n, _ := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc); n != 1 {
		t.Fatalf("final chunk processed %d, want 1", n)
	}
	if p := progress(t, h, svc, ctx, id); p.Done != 5 || p.Pending != 0 || p.Status != "completed" {
		t.Fatalf("after all chunks: %+v, want Done5/Pending0/completed", p)
	}
}

func TestIntegrationBulkEmptyIsCompleted(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, nil)
	p := progress(t, h, svc, ctx, id)
	if p.Total != 0 || p.Status != "completed" {
		t.Fatalf("empty bulk op = %+v, want Total0/completed", p)
	}
}

// okFunc succeeds for every item.
func okFunc(context.Context, database.TenantDB, []byte) error { return nil }

// markFunc writes a row into bulk_marks for each item, then fails the ones whose
// payload says so — proving the failed item's write rolls back.
func markFunc(h *testkit.DBHandle) bulk.ItemFunc {
	ensureMarks(h)
	return func(ctx context.Context, db database.TenantDB, payload []byte) error {
		var it item
		if err := json.Unmarshal(payload, &it); err != nil {
			return err
		}
		if _, err := db.Exec(ctx, `INSERT INTO bulk_marks (val) VALUES ($1)`, it.Val); err != nil {
			return err
		}
		if it.Fail {
			return errors.New("item asked to fail")
		}
		return nil
	}
}

func ensureMarks(h *testkit.DBHandle) {
	ctx := context.Background()
	_, _ = h.Admin.Exec(ctx, `CREATE TABLE IF NOT EXISTS bulk_marks (val int)`)
	_, _ = h.Admin.Exec(ctx, `GRANT INSERT, SELECT ON bulk_marks TO app_rt`)
}
