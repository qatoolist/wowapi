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

// TestIntegrationAuditTxIDCorrelation is the CA-11 regression: audit rows written
// in the SAME database transaction share a tx_id, and rows in different
// transactions get different ones — so a forensic query can correlate every
// change made by one unit of work.
func TestIntegrationAuditTxIDCorrelation(t *testing.T) {
	h := testkit.NewDB(t)
	w := audit.New(model.UUIDv7(), nil)
	tenant, actor := uuid.New(), uuid.New()
	ctx := auditCtx(tenant, actor, "req-1")
	ent := uuid.New()

	// Two records in ONE transaction.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if err := w.Record(ctx, db, audit.Entry{Action: "widget.create", EntityType: "widget", EntityID: ent}); err != nil {
			return err
		}
		return w.Record(ctx, db, audit.Entry{Action: "widget.update", EntityType: "widget", EntityID: ent})
	}); err != nil {
		t.Fatalf("tx1: %v", err)
	}
	// A third record in a SEPARATE transaction.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return w.Record(ctx, db, audit.Entry{Action: "widget.delete", EntityType: "widget", EntityID: ent})
	}); err != nil {
		t.Fatalf("tx2: %v", err)
	}

	var logs []audit.Log
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		logs, e = w.Query(ctx, db, audit.Filter{EntityType: "widget", EntityID: ent})
		return e
	}); err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("want 3 audit rows, got %d", len(logs))
	}
	byAction := map[string]string{}
	for _, l := range logs {
		if l.TxID == "" {
			t.Fatalf("tx_id must be populated, empty on %s", l.Action)
		}
		byAction[l.Action] = l.TxID
	}
	if byAction["widget.create"] != byAction["widget.update"] {
		t.Errorf("same-tx rows must share tx_id: create=%s update=%s", byAction["widget.create"], byAction["widget.update"])
	}
	if byAction["widget.delete"] == byAction["widget.create"] {
		t.Errorf("different-tx rows must differ in tx_id, both = %s", byAction["widget.delete"])
	}
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

// TestIntegrationAuditExportAnchors proves the CA-11 scheduled anchor-export
// primitive: the platform-role ExportAnchors snapshots the tenant's live chain
// head (last seq + head hash) into the append-only audit_anchors table, the
// snapshot matches Anchor(), app_rt can read its own anchor back (tenant-scoped),
// re-running is a no-op until the chain advances, and the anchored (seq, hash)
// verifies against the live chain.
func TestIntegrationAuditExportAnchors(t *testing.T) {
	h := testkit.NewDB(t)
	w := audit.New(model.UUIDv7(), nil)
	tenant := uuid.New()
	ctx := auditCtx(tenant, uuid.New(), "r")
	recordN(t, h, w, ctx, 5)

	// The scheduled export (app_platform, cross-tenant) writes one anchor.
	if n, err := audit.ExportAnchors(context.Background(), h.Platform); err != nil {
		t.Fatalf("ExportAnchors: %v", err)
	} else if n != 1 {
		t.Fatalf("ExportAnchors wrote %d anchors, want 1", n)
	}

	// The anchor must equal the live chain head (Anchor()).
	var wantSeq int64
	var wantHead string
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		wantSeq, wantHead, e = w.Anchor(ctx, db)
		return e
	}); err != nil {
		t.Fatalf("Anchor: %v", err)
	}

	// app_rt reads its own anchor row back (tenant-scoped SELECT grant).
	var gotSeq, gotRows int64
	var gotHead string
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx,
			`SELECT anchor_seq, chain_head_hash, row_count FROM audit_anchors
			  WHERE tenant_id = app_tenant_id() ORDER BY anchor_seq DESC LIMIT 1`).
			Scan(&gotSeq, &gotHead, &gotRows)
	}); err != nil {
		t.Fatalf("read anchor as app_rt: %v", err)
	}
	if gotSeq != wantSeq || gotHead != wantHead || gotRows != 5 || wantSeq != 5 {
		t.Fatalf("anchor = (seq %d, head %q, rows %d), want (5, %q, 5)", gotSeq, gotHead, gotRows, wantHead)
	}

	// Re-running without new audit activity is a no-op (bounded evidence).
	if n, err := audit.ExportAnchors(context.Background(), h.Platform); err != nil {
		t.Fatalf("ExportAnchors (2nd): %v", err)
	} else if n != 0 {
		t.Fatalf("2nd ExportAnchors wrote %d anchors, want 0 (chain unchanged)", n)
	}

	// The anchored (seq, hash) verifies against the untouched live chain.
	var present bool
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		present, e = w.CheckAnchor(ctx, db, gotSeq, gotHead)
		return e
	}); err != nil {
		t.Fatalf("CheckAnchor: %v", err)
	}
	if !present {
		t.Fatal("CheckAnchor false on an untouched chain — anchor should verify")
	}

	// app_rt is read-only: it may SELECT its anchors but never forge one.
	insErr := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := db.Exec(ctx,
			`INSERT INTO audit_anchors (tenant_id, anchor_seq, chain_head_hash, row_count)
			 VALUES (app_tenant_id(), 99, 'forged', 99)`)
		return e
	})
	if insErr == nil || !strings.Contains(strings.ToLower(insErr.Error()), "denied") {
		t.Fatalf("app_rt INSERT on audit_anchors must be denied, got %v", insErr)
	}
}

// TestIntegrationAuditAnchorDetectsTailTruncation is the CA-11 payoff: a tail
// truncation (drop the last rows AND rewind the chain head so the shortened
// chain stays internally consistent) is UNDETECTABLE by Verify alone, but the
// exported anchor catches it — CheckAnchor finds the anchored (seq, hash) gone.
func TestIntegrationAuditAnchorDetectsTailTruncation(t *testing.T) {
	h := testkit.NewDB(t)
	w := audit.New(model.UUIDv7(), nil)
	tenant := uuid.New()
	ctx := auditCtx(tenant, uuid.New(), "r")
	recordN(t, h, w, ctx, 5)

	if _, err := audit.ExportAnchors(context.Background(), h.Platform); err != nil {
		t.Fatalf("ExportAnchors: %v", err)
	}
	var anchoredSeq int64
	var anchoredHead string
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		anchoredSeq, anchoredHead, e = w.Anchor(ctx, db)
		return e
	}); err != nil {
		t.Fatalf("Anchor: %v", err)
	}

	// Simulate an attacker with write access truncating the tail: delete seq 4,5
	// and rewind audit_chain to seq 3 (its head_hash = row_hash of seq 3). Done via
	// Admin (app_rt cannot; audit_logs is append-only for it).
	bg := context.Background()
	var seq3Hash string
	if err := h.Admin.QueryRow(bg,
		`SELECT row_hash FROM audit_logs WHERE tenant_id = $1 AND seq = 3`, tenant).Scan(&seq3Hash); err != nil {
		t.Fatalf("read seq3 hash: %v", err)
	}
	if _, err := h.Admin.Exec(bg,
		`DELETE FROM audit_logs WHERE tenant_id = $1 AND seq > 3`, tenant); err != nil {
		t.Fatalf("truncate tail: %v", err)
	}
	if _, err := h.Admin.Exec(bg,
		`UPDATE audit_chain SET next_seq = 4, head_hash = $2 WHERE tenant_id = $1`, tenant, seq3Hash); err != nil {
		t.Fatalf("rewind chain head: %v", err)
	}

	// Verify alone is fooled: the shortened chain is internally consistent.
	var res audit.VerifyResult
	var present bool
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		if res, e = w.Verify(ctx, db); e != nil {
			return e
		}
		present, e = w.CheckAnchor(ctx, db, anchoredSeq, anchoredHead)
		return e
	}); err != nil {
		t.Fatalf("verify/check: %v", err)
	}
	if !res.OK || res.HeadSeq != 3 {
		t.Fatalf("Verify = %+v, want OK with HeadSeq 3 (truncation hidden from Verify)", res)
	}
	// The anchor catches what Verify could not.
	if present {
		t.Fatal("CheckAnchor passed after tail truncation — anchor evidence failed to detect it")
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
