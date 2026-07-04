package audit_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/testkit"
)

func auditCtx(tenant, actor uuid.UUID, reqID string) context.Context {
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tenant), actor)
	return httpx.WithRequestID(ctx, reqID)
}

func TestIntegrationAuditRecordAndQuery(t *testing.T) {
	h := testkit.NewDB(t)
	w := audit.New(model.UUIDv7(), nil)
	tenant, actor := uuid.New(), uuid.New()
	entity := uuid.New()
	ctx := auditCtx(tenant, actor, "req-abc")

	// Record two field changes on the same entity.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if err := w.Record(ctx, db, audit.Entry{
			Action: "receipt.update", EntityType: "receipt", EntityID: entity,
			Field: "amount", OldValue: "100", NewValue: "150", ActorKind: "user",
		}); err != nil {
			return err
		}
		return w.Record(ctx, db, audit.Entry{
			Action: "receipt.void", EntityType: "receipt", EntityID: entity, Reason: "duplicate",
		})
	}); err != nil {
		t.Fatalf("record: %v", err)
	}

	var logs []audit.Log
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		logs, e = w.Query(ctx, db, audit.Filter{EntityType: "receipt", EntityID: entity})
		return e
	}); err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("got %d audit rows, want 2", len(logs))
	}
	// Newest first: the void (recorded second) leads.
	if logs[0].Action != "receipt.void" {
		t.Errorf("newest action = %q, want receipt.void", logs[0].Action)
	}
	// Field-level capture + actor + request id are recorded.
	var amount audit.Log
	for _, l := range logs {
		if l.Field == "amount" {
			amount = l
		}
	}
	if amount.OldValue != "100" || amount.NewValue != "150" {
		t.Errorf("field change = %s→%s, want 100→150", amount.OldValue, amount.NewValue)
	}
	if amount.ActorID == nil || *amount.ActorID != actor {
		t.Errorf("actor id not captured: %v", amount.ActorID)
	}
	if amount.RequestID != "req-abc" {
		t.Errorf("request id = %q, want req-abc", amount.RequestID)
	}
}

// TestIntegrationAuditAppendOnly proves the runtime role cannot mutate history:
// app_rt has no UPDATE/DELETE on audit_logs (grant-enforced append-only).
func TestIntegrationAuditAppendOnly(t *testing.T) {
	h := testkit.NewDB(t)
	w := audit.New(model.UUIDv7(), nil)
	tenant := uuid.New()
	ctx := auditCtx(tenant, uuid.New(), "r")

	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return w.Record(ctx, db, audit.Entry{Action: "thing.done", EntityType: "thing"})
	}); err != nil {
		t.Fatalf("record: %v", err)
	}

	// UPDATE must be denied for app_rt.
	updErr := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := db.Exec(ctx, `UPDATE audit_logs SET action = 'tampered'`)
		return e
	})
	if updErr == nil || !strings.Contains(strings.ToLower(updErr.Error()), "denied") {
		t.Fatalf("app_rt UPDATE on audit_logs must be denied, got %v", updErr)
	}
	// DELETE must be denied for app_rt.
	delErr := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := db.Exec(ctx, `DELETE FROM audit_logs`)
		return e
	})
	if delErr == nil || !strings.Contains(strings.ToLower(delErr.Error()), "denied") {
		t.Fatalf("app_rt DELETE on audit_logs must be denied, got %v", delErr)
	}
}

func TestIntegrationAuditRedaction(t *testing.T) {
	h := testkit.NewDB(t)
	// Redactor masks the values of the "ssn" field.
	w := audit.New(model.UUIDv7(), func(e *audit.Entry) {
		if e.Field == "ssn" {
			e.OldValue = "***"
			e.NewValue = "***"
		}
	})
	tenant, entity := uuid.New(), uuid.New()
	ctx := auditCtx(tenant, uuid.New(), "r")

	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return w.Record(ctx, db, audit.Entry{
			Action: "person.update", EntityType: "person", EntityID: entity,
			Field: "ssn", OldValue: "111-11-1111", NewValue: "222-22-2222",
		})
	}); err != nil {
		t.Fatalf("record: %v", err)
	}

	var logs []audit.Log
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		logs, e = w.Query(ctx, db, audit.Filter{EntityID: entity})
		return e
	})
	if len(logs) != 1 {
		t.Fatalf("want 1 row, got %d", len(logs))
	}
	if logs[0].OldValue != "***" || logs[0].NewValue != "***" {
		t.Fatalf("sensitive values not redacted: %s / %s", logs[0].OldValue, logs[0].NewValue)
	}
}

// recordN records n simple audit rows for the tenant.
func recordN(t *testing.T, h *testkit.DBHandle, w *audit.Writer, ctx context.Context, n int) {
	t.Helper()
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		for i := 0; i < n; i++ {
			if err := w.Record(ctx, db, audit.Entry{Action: "step", EntityType: "e", NewValue: string(rune('a' + i))}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("record: %v", err)
	}
}

func TestIntegrationAuditChainVerifies(t *testing.T) {
	h := testkit.NewDB(t)
	w := audit.New(model.UUIDv7(), nil)
	tenant := uuid.New()
	ctx := auditCtx(tenant, uuid.New(), "r")

	recordN(t, h, w, ctx, 5)

	var res audit.VerifyResult
	var seq int64
	var head string
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		res, e = w.Verify(ctx, db)
		if e != nil {
			return e
		}
		seq, head, e = w.Anchor(ctx, db)
		return e
	})
	if !res.OK || res.Count != 5 || res.HeadSeq != 5 {
		t.Fatalf("verify = %+v, want OK/Count5/Head5", res)
	}
	if seq != 5 || head == "" {
		t.Fatalf("anchor = (seq %d, head %q), want (5, <hash>)", seq, head)
	}
}

// TestIntegrationAuditChainDetectsMutation mutates a committed row via the admin
// pool (bypassing app_rt's append-only grant) and proves Verify flags it.
func TestIntegrationAuditChainDetectsMutation(t *testing.T) {
	h := testkit.NewDB(t)
	w := audit.New(model.UUIDv7(), nil)
	tenant := uuid.New()
	ctx := auditCtx(tenant, uuid.New(), "r")
	recordN(t, h, w, ctx, 4)

	// Tamper: change the action of seq 2 out-of-band.
	if _, err := h.Admin.Exec(context.Background(),
		`UPDATE audit_logs SET action = 'tampered' WHERE tenant_id = $1 AND seq = 2`, tenant); err != nil {
		t.Fatal(err)
	}
	var res audit.VerifyResult
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		res, e = w.Verify(ctx, db)
		return e
	})
	if res.OK {
		t.Fatal("Verify passed on a mutated chain — tamper-evidence failed")
	}
	if res.BrokenSeq != 2 {
		t.Fatalf("broken at seq %d, want 2 (%s)", res.BrokenSeq, res.Reason)
	}
}

// TestIntegrationAuditChainDetectsDeletion deletes a row and proves the seq gap
// is caught.
func TestIntegrationAuditChainDetectsDeletion(t *testing.T) {
	h := testkit.NewDB(t)
	w := audit.New(model.UUIDv7(), nil)
	tenant := uuid.New()
	ctx := auditCtx(tenant, uuid.New(), "r")
	recordN(t, h, w, ctx, 4)

	if _, err := h.Admin.Exec(context.Background(),
		`DELETE FROM audit_logs WHERE tenant_id = $1 AND seq = 2`, tenant); err != nil {
		t.Fatal(err)
	}
	var res audit.VerifyResult
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		res, e = w.Verify(ctx, db)
		return e
	})
	if res.OK {
		t.Fatal("Verify passed despite a deleted row")
	}
	if res.BrokenSeq != 3 { // seq 2 gone → at the row where seq==3 we expected 2
		t.Fatalf("broken at seq %d, want 3 (gap); reason=%s", res.BrokenSeq, res.Reason)
	}
}

func TestIntegrationAuditTenantIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	w := audit.New(model.UUIDv7(), nil)
	t1, t2 := uuid.New(), uuid.New()

	if err := h.TxM.WithTenant(auditCtx(t1, uuid.New(), "r"), func(ctx context.Context, db database.TenantDB) error {
		return w.Record(ctx, db, audit.Entry{Action: "t1.action", EntityType: "x"})
	}); err != nil {
		t.Fatal(err)
	}
	// Tenant 2 must not see tenant 1's audit rows.
	var logs []audit.Log
	_ = h.TxM.WithTenantRO(auditCtx(t2, uuid.New(), "r"), func(ctx context.Context, db database.TenantDB) error {
		var e error
		logs, e = w.Query(ctx, db, audit.Filter{Action: "t1.action"})
		return e
	})
	if len(logs) != 0 {
		t.Fatalf("tenant 2 saw %d of tenant 1's audit rows (RLS breach)", len(logs))
	}
}
