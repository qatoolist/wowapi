package kernel_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel"
	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestIntegrationDurableAuthzDenial proves the default audit sink writes a
// DURABLE audit_logs row for a sensitive-permission denial — not only a WARN log
// (finding F1). The durable write runs in its own tenant tx because Evaluate is
// read-only.
func TestIntegrationDurableAuthzDenial(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{Pool: h.Runtime, Tx: h.TxM})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	// A sensitive permission — denials of these are always audited (blueprint 07 §1).
	k.Perms.Register(authz.Permission{Key: "secret.thing.read", Sensitive: true})
	if err := k.Perms.Err(); err != nil {
		t.Fatalf("register perm: %v", err)
	}

	tenant := uuid.New()
	actor := authz.Actor{Kind: authz.ActorUser, UserID: uuid.New(), CapacityID: uuid.New(), TenantID: tenant}
	ctx := database.WithTenantID(context.Background(), tenant)

	// Evaluate in a read-only tx → default_deny (no grant) → sensitive → audited.
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		d, e := k.Authz.Evaluate(ctx, db, actor, "secret.thing.read", authz.Target{Scope: authz.ScopeTenant})
		if e != nil {
			return e
		}
		if d.Allowed {
			t.Fatal("expected a denial")
		}
		return nil
	}); err != nil {
		t.Fatalf("evaluate: %v", err)
	}

	// A durable audit_logs row must now exist for this tenant.
	var n int
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx,
			`SELECT count(*) FROM audit_logs WHERE action = 'authz.denied'`).Scan(&n)
	}); err != nil {
		t.Fatalf("read audit_logs: %v", err)
	}
	if n != 1 {
		t.Fatalf("durable authz-denial audit rows = %d, want 1 (denial must be durably audited, not just logged)", n)
	}
}
