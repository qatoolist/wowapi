package app_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/testkit"
)

// TestStartWorkerRunsAuditAnchorExport is the CA-11 wiring regression: the
// scheduled anchor-export task registered by registerMaintenance runs on the
// leader-safe scheduler under StartWorker and durably persists the tenant's audit
// chain head into audit_anchors — proving the sweep is wired (not built-but-not-
// wired) and fires on its interval. The persisted anchor must equal the live
// chain head.
func TestStartWorkerRunsAuditAnchorExport(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}

	// Seed a tenant's audit chain via the runtime path (app_rt), before the worker
	// starts, so the first anchor tick has a head to snapshot.
	w := audit.New(model.UUIDv7(), nil)
	tenant := uuid.New()
	tctx := httpx.WithRequestID(database.WithActorID(
		database.WithTenantID(context.Background(), tenant), uuid.New()), "r")
	if err := h.TxM.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
		for i := 0; i < 3; i++ {
			if err := w.Record(ctx, db, audit.Entry{Action: "step", EntityType: "e"}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("seed audit rows: %v", err)
	}

	booted, err := app.New().Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- app.StartWorker(ctx, booted, app.WorkerConfigOpts{
			RelayPoll:           80 * time.Millisecond,
			JobPoll:             80 * time.Millisecond,
			SchedulerPoll:       40 * time.Millisecond,
			SLAInterval:         time.Hour,
			IdempotencyInterval: time.Hour,
			DLQDepthInterval:    time.Hour,
			AuditAnchorInterval: 100 * time.Millisecond,
			ShutdownDrain:       3 * time.Second,
		})
	}()

	// Wait until the anchor-export sweep has persisted the tenant's head.
	deadline := time.After(12 * time.Second)
	var anchored int64
	for anchored == 0 {
		select {
		case <-deadline:
			t.Fatal("audit anchor-export never persisted an anchor for the tenant")
		case err := <-done:
			t.Fatalf("StartWorker returned early: %v", err)
		case <-time.After(50 * time.Millisecond):
			if err := h.Admin.QueryRow(context.Background(),
				`SELECT count(*) FROM audit_anchors WHERE tenant_id = $1`, tenant).Scan(&anchored); err != nil {
				t.Fatalf("count anchors: %v", err)
			}
		}
	}

	// The persisted anchor must equal the live chain head.
	var gotSeq, gotRows int64
	var gotHead string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT anchor_seq, chain_head_hash, row_count FROM audit_anchors
		  WHERE tenant_id = $1 ORDER BY anchor_seq DESC LIMIT 1`, tenant).
		Scan(&gotSeq, &gotHead, &gotRows); err != nil {
		t.Fatalf("read anchor: %v", err)
	}
	var wantSeq int64
	var wantHead string
	if err := h.TxM.WithTenantRO(tctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		wantSeq, wantHead, e = w.Anchor(ctx, db)
		return e
	}); err != nil {
		t.Fatalf("Anchor: %v", err)
	}
	if gotSeq != wantSeq || gotHead != wantHead || gotSeq != 3 || gotRows != 3 {
		t.Fatalf("anchored (seq %d, head %q, rows %d), want (3, %q, 3)", gotSeq, gotHead, gotRows, wantHead)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("StartWorker returned error on shutdown: %v", err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("StartWorker did not drain within the shutdown window")
	}
}
