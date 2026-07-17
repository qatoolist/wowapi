package kernel_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationStepUpStrongFactorsConfigurable is B8's kernel-composition
// regression: Deps.StepUpStrongFactors is the deployment config surface for
// the default strong-factor set, wired straight into the evaluator kernel.New
// builds — sms is opted back in via Deps alone, no code change.
func TestIntegrationStepUpStrongFactorsConfigurable(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tenant := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	capID := testkit.CreateCapacity(t, h, tenant.ID, userID)
	const perm = "billing.export.read"

	role := testkit.CreateRole(t, h, tenant.ID, "biller", perm)
	testkit.GrantRole(t, h, tenant.ID, capID, role, "tenant", nil, "")

	ctx := database.WithTenantID(context.Background(), tenant.ID)
	target := authz.Target{Scope: authz.ScopeTenant}
	newActorWithAMR := func(amr ...string) authz.Actor {
		return authz.Actor{Kind: authz.ActorUser, UserID: userID, CapacityID: capID, TenantID: tenant.ID, AMR: amr}
	}

	// --- Default config (no StepUpStrongFactors set): sms does NOT satisfy. ---
	kDefault, err := kernel.New(config.Defaults(), log, kernel.Deps{Pool: h.Runtime, Tx: h.TxM})
	if err != nil {
		t.Fatalf("kernel.New (default): %v", err)
	}
	kDefault.Perms.Register(authz.Permission{Key: perm, StepUp: true})
	if err := kDefault.Perms.Err(); err != nil {
		t.Fatal(err)
	}

	var d authz.Decision
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		d, e = kDefault.Authz.Evaluate(ctx, db, newActorWithAMR("pwd", "sms"), perm, target)
		return e
	}); err != nil {
		t.Fatalf("Evaluate (default, sms): %v", err)
	}
	if d.Allowed || !d.StepUpRequired {
		t.Fatalf("sms under default kernel config = %+v, want denied+StepUpRequired (sms is opt-in only)", d)
	}

	// --- Deployment config opts sms back in via Deps only, no code change. ---
	kSMS, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Tx: h.TxM,
		StepUpStrongFactors: append([]string{"sms"}, authz.DefaultStrongFactors...),
	})
	if err != nil {
		t.Fatalf("kernel.New (sms opt-in): %v", err)
	}
	kSMS.Perms.Register(authz.Permission{Key: perm, StepUp: true})
	if err := kSMS.Perms.Err(); err != nil {
		t.Fatal(err)
	}

	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		d, e = kSMS.Authz.Evaluate(ctx, db, newActorWithAMR("pwd", "sms"), perm, target)
		return e
	}); err != nil {
		t.Fatalf("Evaluate (sms opt-in): %v", err)
	}
	if !d.Allowed || d.StepUpRequired {
		t.Fatalf("sms with sms opted into Deps.StepUpStrongFactors = %+v, want allowed", d)
	}
}
