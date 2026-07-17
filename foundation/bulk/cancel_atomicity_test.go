package bulk_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/qatoolist/wowapi/v2/foundation/bulk"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// Closure-review regression (adversarial closure review 2026-07-17, F-04):
// cancellation must atomically update aggregate and item state, or remain
// safely retryable. The previous two-transaction Cancel could commit the
// aggregate as cancelled and then fail the pending-item sweep — terminal
// aggregate, stranded pending items, and no legal transition to repair it.
func TestIntegrationCancelAtomicUnderInjectedFailure(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithBatchSize(1))
	tenant := uuid.New()
	tctx := database.WithTenantID(context.Background(), tenant)

	var bulkID uuid.UUID
	payload, _ := json.Marshal(map[string]int{"val": 1})
	if err := h.TxM.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		bulkID, e = svc.Start(ctx, db, "test.bulk", []json.RawMessage{payload, payload})
		return e
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Inject a failure between the aggregate transition and the item cleanup —
	// INSIDE Cancel's single tenant transaction. Both writes must roll back.
	boom := errors.New("injected failure between cancellation writes")
	bulk.SetCancelInterceptor(svc, func(ctx context.Context, db database.TenantDB) error { return boom })
	if err := svc.Cancel(context.Background(), h.TxM, tenant, bulkID); !errors.Is(err, boom) {
		t.Fatalf("Cancel with injected failure returned %v, want the injected error", err)
	}

	var status string
	var pending int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT status FROM bulk_operations WHERE id = $1`, bulkID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM bulk_items WHERE bulk_id = $1 AND status = 'pending'`, bulkID).Scan(&pending); err != nil {
		t.Fatal(err)
	}
	if status == "cancelled" {
		t.Fatalf("aggregate committed as cancelled despite the failed transaction (pending items: %d) — partial write", pending)
	}
	if pending != 2 {
		t.Fatalf("pending items = %d after rolled-back cancel, want 2", pending)
	}

	// Retry without the fault: must fully cancel aggregate AND items.
	bulk.SetCancelInterceptor(svc, nil)
	if err := svc.Cancel(context.Background(), h.TxM, tenant, bulkID); err != nil {
		t.Fatalf("retry Cancel: %v", err)
	}
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT status FROM bulk_operations WHERE id = $1`, bulkID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	var stranded int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM bulk_items WHERE bulk_id = $1 AND status = 'pending'`, bulkID).Scan(&stranded); err != nil {
		t.Fatal(err)
	}
	if status != "cancelled" || stranded != 0 {
		t.Fatalf("after retry: status=%q stranded-pending=%d, want cancelled/0", status, stranded)
	}

	// Idempotent re-cancel of an already-cancelled aggregate completes cleanly
	// (the interrupted-cancellation repair path); completed stays terminal —
	// covered by the lifecycle matrix.
	if err := svc.Cancel(context.Background(), h.TxM, tenant, bulkID); err != nil {
		t.Fatalf("idempotent re-cancel: %v", err)
	}
}

// The idempotent-cancel path must NOT weaken terminal completed state.
func TestIntegrationCancelOfCompletedStillRejected(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithBatchSize(1))
	tenant := uuid.New()
	tctx := database.WithTenantID(context.Background(), tenant)

	var bulkID uuid.UUID
	payload, _ := json.Marshal(map[string]int{"val": 1})
	if err := h.TxM.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		bulkID, e = svc.Start(ctx, db, "test.bulk", []json.RawMessage{payload})
		return e
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if _, err := svc.Process(context.Background(), h.TxM, tenant, bulkID, 0,
		func(context.Context, database.TenantDB, bulk.Item) error { return nil }); err != nil {
		t.Fatalf("Process: %v", err)
	}
	err := svc.Cancel(context.Background(), h.TxM, tenant, bulkID)
	if err == nil || kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("Cancel(completed) = %v, want KindConflict invalid transition", err)
	}
}
