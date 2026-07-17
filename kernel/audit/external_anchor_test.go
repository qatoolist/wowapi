package audit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/audit"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/testkit"
)

func anchorCtx(tenant uuid.UUID) context.Context {
	return database.WithActorID(database.WithTenantID(context.Background(), tenant), uuid.New())
}

// TestIntegrationExternalAnchorTamperDetection proves the external anchor closes
// Verify's blind spot: after anchoring the chain, an attacker truncates the tail
// rows and rewinds audit_chain.head_hash; Verify detects that the anchored head
// is no longer present.
func TestIntegrationExternalAnchorTamperDetection(t *testing.T) {
	h := testkit.NewDB(t)
	w := audit.New(nil, nil)
	tenant := uuid.New()
	ctx := anchorCtx(tenant)
	store := audit.NewFileStore(t.TempDir())
	ea := audit.NewExternalAnchor(store, w)

	// Seed three audit rows.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		for range 3 {
			if err := w.Record(ctx, db, audit.Entry{Action: "step", EntityType: "e"}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("seed audit rows: %v", err)
	}

	// Anchor the chain head externally.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return ea.AnchorNow(ctx, db)
	}); err != nil {
		t.Fatalf("anchor: %v", err)
	}

	// Before tampering, Verify passes.
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return ea.Verify(ctx, db)
	}); err != nil {
		t.Fatalf("verify before tamper: %v", err)
	}

	// Capture the hash of seq 1 before tampering.
	var seq1Hash string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT row_hash FROM audit_logs WHERE tenant_id = $1 AND seq = 1`, tenant).Scan(&seq1Hash); err != nil {
		t.Fatalf("read seq1 hash: %v", err)
	}

	// Tamper: truncate all rows after seq 1 and rewind the chain head to seq 1's hash.
	if _, err := h.Admin.Exec(context.Background(),
		`DELETE FROM audit_logs WHERE tenant_id = $1 AND seq > 1`, tenant); err != nil {
		t.Fatalf("truncate audit tail: %v", err)
	}
	if _, err := h.Admin.Exec(context.Background(),
		`UPDATE audit_chain SET next_seq = 2, head_hash = $1 WHERE tenant_id = $2`, seq1Hash, tenant); err != nil {
		t.Fatalf("rewind chain head: %v", err)
	}

	// Local Verify still passes because the remaining chain is internally consistent.
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		res, err := w.Verify(ctx, db)
		if err != nil {
			return err
		}
		if !res.OK {
			t.Fatalf("local Verify failed unexpectedly: %s", res.Reason)
		}
		return nil
	}); err != nil {
		t.Fatalf("local verify after tamper: %v", err)
	}

	// External Verify detects the tamper.
	err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return ea.Verify(ctx, db)
	})
	if err == nil {
		t.Fatal("verify after tamper succeeded, want tamper error")
	}
	if !errors.Is(err, audit.ErrAnchorTampered) {
		t.Fatalf("verify error = %v, want ErrAnchorTampered", err)
	}
	if kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("verify error kind = %v, want conflict", kerr.KindOf(err))
	}
}
